package main

import (
        "neoai/adapter"
        "neoai/addition"
        "neoai/admin"
        "neoai/auth"
        "neoai/channel"
        "neoai/cli"
        "neoai/community"
        "neoai/globals"
        "neoai/manager"
        "neoai/manager/conversation"
        "neoai/middleware"
        "neoai/utils"
        "fmt"
        "github.com/gin-gonic/gin"
        "github.com/spf13/viper"
        "net/url"
)

func readCorsOrigins() {
        origins := viper.GetStringSlice("allow_origins")
        if len(origins) > 0 {
                globals.AllowedOrigins = utils.Each(origins, func(origin string) string {
                        // remove protocol and trailing slash
                        // e.g. https://neoai.net/ -> neoai.net

                        if host, err := url.Parse(origin); err == nil {
                                return host.Host
                        }

                        return origin
                })
        }
}

func registerApiRouter(engine *gin.Engine) {
        // NeoAI URL contract:
        //
        //   Frontend ALWAYS uses baseURL = "/api" (set in app/src/conf/env.ts
        //   as `VITE_BACKEND_ENDPOINT || "/api"`). When deployed cross-origin,
        //   the user sets VITE_BACKEND_ENDPOINT to the full backend URL
        //   (e.g. "https://api.example.com") — but the frontend still
        //   appends "/api" to it implicitly via the `getRestApi` helper.
        //
        //   Backend ALWAYS mounts every API route under `/api`. No routes at
        //   root, no dual-mounting. One convention, one place.
        //
        //   The only root-level endpoints are:
        //     - /healthz, /ready        (health probes, see middleware/health.go)
        //     - /v1/*, /mj/*, /attachments/* (legacy redirects, see utils/config.go)
        //     - static assets            (when serve_static is true)
        app := engine.Group("/api")
        {
                auth.Register(app)
                admin.Register(app)
                adapter.Register(app)
                manager.Register(app)
                addition.Register(app)
                conversation.Register(app)
                community.Register(app)
        }
}

func main() {
        utils.ReadConf()
        admin.InitInstance()
        channel.InitManager()

        if cli.Run() {
                return
        }

        app := utils.NewEngine()
        worker := middleware.RegisterMiddleware(app)
        defer worker()

        // Community feature — run DB migrations now that middleware has
        // initialised the global `connection.DB` handle.
        if db := utils.GetDBFromContextSafe(); db != nil {
                if err := community.Migrate(db); err != nil {
                        globals.Warn(fmt.Sprintf("[community] migration failed: %s", err.Error()))
                }
                // NeoAI: ensure the registration_code table exists. We can't
                // do this from the connection package (import cycle), so we
                // do it here right after middleware init.
                auth.EnsureRegistrationCodeTable(db)
        }

        // Health endpoints — must be registered BEFORE AuthMiddleware kicks in
        // so external probes can reach them without a JWT.
        middleware.RegisterHealth(app)

        utils.RegisterStaticRoute(app)
        registerApiRouter(app)
        readCorsOrigins()

        if err := app.Run(fmt.Sprintf(":%s", viper.GetString("server.port"))); err != nil {
                panic(err)
        }
}
