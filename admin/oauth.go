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

        // NeoAI: save the entire oauth block in ONE SaveConfig call.
        // Calling SaveConfig multiple times in a row with dotted keys
        // causes viper to clobber nested map entries on each reload.
        // Building the full nested struct and saving once avoids that.
        oauthBlock := map[string]interface{}{
                "linuxdo": map[string]interface{}{
                        "client_id":     form.LinuxDo.ClientID,
                        "client_secret": form.LinuxDo.ClientSecret,
                },
                "github": map[string]interface{}{
                        "client_id":     form.GitHub.ClientID,
                        "client_secret": form.GitHub.ClientSecret,
                },
                "frontend_url": form.FrontendURL,
        }
        if err := utils.SaveConfig("system.oauth", oauthBlock); err != nil {
                c.JSON(http.StatusOK, gin.H{
                        "status":  false,
                        "message": err.Error(),
                })
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "status": true,
        })
}
