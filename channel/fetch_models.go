package channel

import (
        "encoding/json"
        "fmt"
        "io"
        "net/http"
        "strings"
        "time"

        "neoai/utils"

        "github.com/gin-gonic/gin"
)

// fetch_models.go — admin endpoint to auto-discover available models
// from an upstream channel's `/v1/models` (OpenAI-compatible) endpoint.
//
// Many upstream providers (OpenAI itself, Azure, Anthropic, Moonshot,
// Groq, DeepSeek, Together, Anyscale, …) expose a GET /v1/models
// endpoint that returns a JSON list of available model IDs. Even for
// providers that don't, the call simply 404s and the admin can still
// paste model IDs manually — so this is always safe to call.

type fetchModelsForm struct {
        Type     string `json:"type"`
        Endpoint string `json:"endpoint"`
        Secret   string `json:"secret"`
}

type upstreamModelItem struct {
        ID      string `json:"id"`
        Object  string `json:"object"`
        OwnedBy string `json:"owned_by"`
}

type upstreamModelsResponse struct {
        Object string              `json:"object"`
        Data   []upstreamModelItem `json:"data"`
}

// FetchModelsAPI — POST /admin/channel/fetch-models
//
// Body:
//   { "type": "openai", "endpoint": "https://api.openai.com", "secret": "sk-..." }
//
// Response:
//   { "status": true, "data": ["gpt-4", "gpt-3.5-turbo", ...] }
//
// Only OpenAI-compatible channel types support this — for others the
// endpoint will return an empty list rather than failing.
func FetchModelsAPI(c *gin.Context) {
        // Auth check without importing the auth package (which would create
        // an import cycle, since auth already imports channel).
        if !utils.GetAdminFromContext(c) {
                c.JSON(http.StatusUnauthorized, gin.H{
                        "status": false,
                        "error":  "admin required",
                })
                return
        }

        var form fetchModelsForm
        if err := c.ShouldBindJSON(&form); err != nil {
                c.JSON(http.StatusOK, gin.H{
                        "status": false,
                        "error":  err.Error(),
                })
                return
        }

        if form.Endpoint == "" || form.Secret == "" {
                c.JSON(http.StatusOK, gin.H{
                        "status": false,
                        "error":  "endpoint and secret are required",
                })
                return
        }

        // Only OpenAI-compatible channel types support the /v1/models listing.
        // For other types we silently return an empty list — the admin UI then
        // falls back to the curated default list for that channel type.
        compatibleTypes := map[string]bool{
                "openai":   true,
                "azure":    false, // uses deployment-name style, not /v1/models
                "claude":   false,
                "moonshot": true,
                "groq":     true,
                "deepseek": true,
                "qwen":     false, // DashScope uses different listing API
                "chatglm":  true,  // BigModel is OpenAI-compatible
                "bailian":  true,
                "together": true,
        }
        if !compatibleTypes[form.Type] {
                c.JSON(http.StatusOK, gin.H{
                        "status": true,
                        "data":   []string{},
                        "note":   fmt.Sprintf("channel type %q does not expose /v1/models — please add models manually", form.Type),
                })
                return
        }

        // Normalize endpoint: strip trailing slash, ensure has scheme.
        endpoint := strings.TrimSpace(form.Endpoint)
        endpoint = strings.TrimRight(endpoint, "/")
        if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
                endpoint = "https://" + endpoint
        }
        url := endpoint + "/v1/models"

        client := &http.Client{Timeout: 15 * time.Second}
        req, err := http.NewRequest("GET", url, nil)
        if err != nil {
                c.JSON(http.StatusOK, gin.H{
                        "status": false,
                        "error":  err.Error(),
                })
                return
        }
        // Both `Authorization: Bearer sk-...` (OpenAI-style) and
        // `x-api-key: sk-...` (Anthropic-style) are tried — one of them
        // usually works for OpenAI-compatible providers.
        req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(form.Secret))
        req.Header.Set("x-api-key", strings.TrimSpace(form.Secret))
        req.Header.Set("Accept", "application/json")

        resp, err := client.Do(req)
        if err != nil {
                c.JSON(http.StatusOK, gin.H{
                        "status": false,
                        "error":  fmt.Sprintf("upstream request failed: %s", err.Error()),
                })
                return
        }
        defer resp.Body.Close()

        body, _ := io.ReadAll(resp.Body)
        if resp.StatusCode != 200 {
                c.JSON(http.StatusOK, gin.H{
                        "status": false,
                        "error":  fmt.Sprintf("upstream returned HTTP %d", resp.StatusCode),
                        "body":   string(body),
                })
                return
        }

        var parsed upstreamModelsResponse
        if err := json.Unmarshal(body, &parsed); err != nil {
                c.JSON(http.StatusOK, gin.H{
                        "status": false,
                        "error":  fmt.Sprintf("invalid JSON response: %s", err.Error()),
                        "body":   string(body),
                })
                return
        }

        models := make([]string, 0, len(parsed.Data))
        for _, item := range parsed.Data {
                if strings.TrimSpace(item.ID) != "" {
                        models = append(models, item.ID)
                }
        }

        c.JSON(http.StatusOK, gin.H{
                "status": true,
                "data":   models,
        })
}
