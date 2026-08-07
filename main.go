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
        var app *gin.RouterGroup
        if !viper.GetBool("serve_static") {
                app = engine.Group("")
        } else {
                app = engine.Group("/api")
        }

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
