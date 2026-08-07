package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"neoai/globals"
	"neoai/utils"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// oauth.go — LinuxDO and GitHub OAuth2 login.
//
// Both providers follow the standard authorization-code flow:
//
//   1. Frontend redirects user to the provider's authorize URL with our
//      client_id + redirect_uri + state.
//   2. Provider redirects back to /oauth/<provider>/callback?code=...&state=...
//   3. Backend exchanges the code for an access token.
//   4. Backend uses the access token to fetch the user profile.
//   5. Backend looks up (or auto-creates) a local user bound to that
//      provider identity and issues a JWT.
//
// The bind is stored in the `auth.bind_id` column (already used by the
// deeptrain flow) — we store provider+provider_uid encoded as the
// "external identity" key, so the same external identity always maps
// to the same local user.
//
// Config (in config.yaml under `system.oauth`):
//   system.oauth.linuxdo.client_id
//   system.oauth.linuxdo.client_secret
//   system.oauth.github.client_id
//   system.oauth.github.client_secret
//
// Redirect URIs (must be registered with the provider):
//   <backend_url>/api/oauth/linuxdo/callback
//   <backend_url>/api/oauth/github/callback

type OAuthProvider string

const (
	ProviderLinuxDo OAuthProvider = "linuxdo"
	ProviderGitHub  OAuthProvider = "github"
)

type oauthProviderConfig struct {
	AuthorizeURL string
	TokenURL     string
	UserInfoURL  string
	Scope        string
}

var providerConfigs = map[OAuthProvider]oauthProviderConfig{
	ProviderLinuxDo: {
		AuthorizeURL: "https://connect.linux.do/oauth2/authorize",
		TokenURL:     "https://connect.linux.do/oauth2/token",
		UserInfoURL:  "https://connect.linux.do/api/user",
		Scope:        "",
	},
	ProviderGitHub: {
		AuthorizeURL: "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		UserInfoURL:  "https://api.github.com/user",
		Scope:        "read:user user:email",
	},
}

// OAuthState is the per-request anti-CSRF state. We just use a random
// string and stash it in the cache so the callback can validate it.
type OAuthState struct {
	State    string `json:"state"`
	Provider string `json:"provider"`
}

// externalIdentityKey returns the unique key for an external identity.
// Stored in `auth.bind_id` column as a string hash so we can look up
// users by their provider+uid without a separate table.
func externalIdentityKey(provider OAuthProvider, providerUID string) string {
	return utils.Sha2Encrypt(fmt.Sprintf("%s:%s", provider, providerUID))
}

// getOAuthConfig returns the client_id / client_secret for a provider.
func getOAuthConfig(provider OAuthProvider) (clientID, clientSecret string) {
	switch provider {
	case ProviderLinuxDo:
		return viper.GetString("system.oauth.linuxdo.client_id"),
			viper.GetString("system.oauth.linuxdo.client_secret")
	case ProviderGitHub:
		return viper.GetString("system.oauth.github.client_id"),
			viper.GetString("system.oauth.github.client_secret")
	}
	return "", ""
}

// isOAuthEnabled reports whether a provider is configured.
func isOAuthEnabled(provider OAuthProvider) bool {
	id, secret := getOAuthConfig(provider)
	return id != "" && secret != ""
}

// getRedirectURL returns the canonical callback URL for a provider.
// Uses system.general.backend if set, otherwise falls back to the
// request's Host header.
func getRedirectURL(c *gin.Context, provider OAuthProvider) string {
	backend := viper.GetString("system.general.backend")
	if backend == "" {
		// fall back to request origin
		scheme := "https"
		if c.Request.TLS == nil {
			if fwd := c.Request.Header.Get("X-Forwarded-Proto"); fwd != "" {
				scheme = fwd
			} else {
				scheme = "http"
			}
		}
		backend = fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	}
	backend = strings.TrimRight(backend, "/")
	prefix := "/api"
	if !viper.GetBool("serve_static") {
		prefix = ""
	}
	return fmt.Sprintf("%s%s/oauth/%s/callback", backend, prefix, provider)
}

// OAuthLoginAPI — GET /oauth/:provider/login
// Generates a state, stashes it in the cache, and redirects the user to
// the provider's authorize URL.
func OAuthLoginAPI(c *gin.Context) {
	providerStr := c.Param("provider")
	provider := OAuthProvider(providerStr)
	cfg, ok := providerConfigs[provider]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "unknown provider"})
		return
	}
	if !isOAuthEnabled(provider) {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": fmt.Sprintf("%s oauth is not configured", provider)})
		return
	}

	clientID, _ := getOAuthConfig(provider)
	state := utils.GenerateChar(32)
	redirectURI := getRedirectURL(c, provider)

	// stash state in cache for 10 minutes
	cache := utils.GetCacheFromContext(c)
	cache.Set(c.Request.Context(), fmt.Sprintf("oauth:state:%s", state), string(provider), 10*time.Minute)

	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("state", state)
	if cfg.Scope != "" {
		q.Set("scope", cfg.Scope)
	}

	authorizeURL := cfg.AuthorizeURL + "?" + q.Encode()
	c.Redirect(http.StatusFound, authorizeURL)
}

// OAuthCallbackAPI — GET /oauth/:provider/callback
// Validates the state, exchanges the code for an access token, fetches
// the user profile, and issues a JWT for the local user (creating one
// if necessary).
func OAuthCallbackAPI(c *gin.Context) {
	providerStr := c.Param("provider")
	provider := OAuthProvider(providerStr)
	cfg, ok := providerConfigs[provider]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "unknown provider"})
		return
	}
	if !isOAuthEnabled(provider) {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": fmt.Sprintf("%s oauth is not configured", provider)})
		return
	}

	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "missing code or state"})
		return
	}

	// validate state
	cache := utils.GetCacheFromContext(c)
	stashedProvider, err := cache.Get(c.Request.Context(), fmt.Sprintf("oauth:state:%s", state)).Result()
	if err != nil || stashedProvider != string(provider) {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "invalid or expired state"})
		return
	}
	cache.Del(c.Request.Context(), fmt.Sprintf("oauth:state:%s", state))

	// exchange code for access token
	accessToken, err := exchangeOAuthCode(provider, cfg, code, getRedirectURL(c, provider))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": fmt.Sprintf("token exchange failed: %s", err.Error())})
		return
	}

	// fetch user profile
	externalUID, username, email, err := fetchOAuthUserInfo(provider, cfg, accessToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": fmt.Sprintf("user info fetch failed: %s", err.Error())})
		return
	}

	db := utils.GetDBFromContext(c)
	token, err := loginOrCreateExternalUser(db, cache, provider, externalUID, username, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "error": err.Error()})
		return
	}

	// redirect to frontend with token in the URL fragment so the SPA
	// can pick it up. Use fragment (#) instead of query so the token
	// doesn't end up in server access logs.
	frontendURL := utils.GetStringWithDefault("system.oauth.frontend_url", "/")
	if !strings.HasPrefix(frontendURL, "http") {
		frontendURL = frontendURL
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("%s#access_token=%s", frontendURL, token))
}

// exchangeOAuthCode trades an authorization code for an access token.
func exchangeOAuthCode(provider OAuthProvider, cfg oauthProviderConfig, code, redirectURI string) (string, error) {
	clientID, clientSecret := getOAuthConfig(provider)

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	req, err := http.NewRequest("POST", cfg.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	if provider == ProviderGitHub {
		// GitHub accepts JSON but defaults to urlencoded response unless
		// you ask for JSON explicitly.
		req.Header.Set("Accept", "application/json")
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("token endpoint returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Most providers return JSON; GitHub also returns JSON if Accept header is set.
	var parsed struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("invalid token response: %s", err.Error())
	}
	if parsed.Error != "" {
		return "", fmt.Errorf("%s: %s", parsed.Error, parsed.ErrorDesc)
	}
	if parsed.AccessToken == "" {
		return "", errors.New("empty access_token in response")
	}
	return parsed.AccessToken, nil
}

// fetchOAuthUserInfo calls the provider's user-info endpoint and
// returns (provider_uid, username, email).
func fetchOAuthUserInfo(provider OAuthProvider, cfg oauthProviderConfig, accessToken string) (uid, username, email string, err error) {
	req, err := http.NewRequest("GET", cfg.UserInfoURL, nil)
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", "", "", fmt.Errorf("user info endpoint returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	switch provider {
	case ProviderLinuxDo:
		var info struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
			Name     string `json:"name"`
			Email    string `json:"email"`
		}
		if err := json.Unmarshal(body, &info); err != nil {
			return "", "", "", err
		}
		uid = fmt.Sprintf("%d", info.ID)
		username = info.Username
		if username == "" {
			username = info.Name
		}
		email = info.Email
	case ProviderGitHub:
		var info struct {
			ID    int64  `json:"id"`
			Login string `json:"login"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := json.Unmarshal(body, &info); err != nil {
			return "", "", "", err
		}
		uid = fmt.Sprintf("%d", info.ID)
		username = info.Login
		if username == "" {
			username = info.Name
		}
		email = info.Email
	}
	if uid == "" {
		return "", "", "", errors.New("provider returned empty user id")
	}
	return uid, username, email, nil
}

// loginOrCreateExternalUser looks up a local user bound to the given
// external identity. If none exists, creates one (auto-registration).
// Returns a JWT for the user.
//
// Username collision handling: if the OAuth provider's username is
// already taken locally, append a random suffix.
func loginOrCreateExternalUser(db *sql.DB, cache interface{}, provider OAuthProvider, externalUID, username, email string) (string, error) {
	identityKey := externalIdentityKey(provider, externalUID)

	// Look up by identity key (stored in `token` column for external
	// users — coai already uses this pattern for deeptrain).
	var existingID int64
	err := globals.QueryRowDb(db, `SELECT id FROM auth WHERE token = ?`, identityKey).Scan(&existingID)
	if err == nil {
		// existing user — issue a fresh JWT
		u := &User{ID: existingID}
		token, err := u.GenerateTokenSafe(db)
		if err != nil {
			return "", err
		}
		return token, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	// new external user — auto-register
	if username == "" {
		username = fmt.Sprintf("%s_%s", provider, externalUID)
	}
	// ensure username uniqueness
	if IsUserExist(db, username) {
		username = fmt.Sprintf("%s_%s", username, utils.GenerateChar(4))
	}
	if email != "" && IsEmailExist(db, email) {
		// don't fail just because the OAuth provider's email is already
		// used by a different local user — just clear the email field.
		email = ""
	}

	// generate a random password so password-based login is impossible
	randomPassword := utils.GenerateChar(64)
	hash := utils.Sha2Encrypt(randomPassword)

	// pick a non-conflicting bind_id
	bindID := getMaxBindId(db) + 1

	result, err := globals.ExecDb(db, `
		INSERT INTO auth (username, password, email, bind_id, token)
		VALUES (?, ?, ?, ?, ?)
	`, username, hash, email, bindID, identityKey)
	if err != nil {
		return "", err
	}
	newID, _ := result.LastInsertId()

	u := &User{ID: newID, Username: username, Password: hash}
	u.CreateInitialQuota(db)

	return u.GenerateToken()
}
