package admin

import (
	"net/http"

	"neoai/utils"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// oauth.go — admin endpoints for managing OAuth provider configuration.
//
//   GET  /admin/oauth/config          — returns the current oauth config
//   POST /admin/oauth/config          — saves the oauth config
//
// The config is persisted under `system.oauth.*` in config.yaml:
//
//   system:
//     oauth:
//       linuxdo:
//         client_id: ...
//         client_secret: ...
//       github:
//         client_id: ...
//         client_secret: ...
//       frontend_url: https://app.example.com   # where to redirect after login

type oauthProviderForm struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type oauthConfigForm struct {
	LinuxDo     oauthProviderForm `json:"linuxdo"`
	GitHub      oauthProviderForm `json:"github"`
	FrontendURL string            `json:"frontend_url"`
}

type oauthConfigResponse struct {
	Status      bool              `json:"status"`
	LinuxDo     oauthProviderForm `json:"linuxdo"`
	GitHub      oauthProviderForm `json:"github"`
	FrontendURL string            `json:"frontend_url"`
}

// GetOAuthConfigAPI — GET /admin/oauth/config
// Returns the saved OAuth config. Secrets are returned as-is so the
// admin form can pre-fill them — this endpoint is admin-only.
func GetOAuthConfigAPI(c *gin.Context) {
	c.JSON(http.StatusOK, oauthConfigResponse{
		Status: true,
		LinuxDo: oauthProviderForm{
			ClientID:     viper.GetString("system.oauth.linuxdo.client_id"),
			ClientSecret: viper.GetString("system.oauth.linuxdo.client_secret"),
		},
		GitHub: oauthProviderForm{
			ClientID:     viper.GetString("system.oauth.github.client_id"),
			ClientSecret: viper.GetString("system.oauth.github.client_secret"),
		},
		FrontendURL: viper.GetString("system.oauth.frontend_url"),
	})
}

// SetOAuthConfigAPI — POST /admin/oauth/config
// Persists the entire oauth config block. Empty client_id/secret disables
// the provider (the /info endpoint reports it as not-configured).
func SetOAuthConfigAPI(c *gin.Context) {
	var form oauthConfigForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	// Save each field separately so we don't clobber other system.oauth.*
	// keys that might have been set elsewhere.
	updates := map[string]interface{}{
		"system.oauth.linuxdo.client_id":     form.LinuxDo.ClientID,
		"system.oauth.linuxdo.client_secret": form.LinuxDo.ClientSecret,
		"system.oauth.github.client_id":      form.GitHub.ClientID,
		"system.oauth.github.client_secret":  form.GitHub.ClientSecret,
		"system.oauth.frontend_url":          form.FrontendURL,
	}
	for k, v := range updates {
		if err := utils.SaveConfig(k, v); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status":  false,
				"message": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}
