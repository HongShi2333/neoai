package middleware

import (
        "neoai/connection"
        "neoai/utils"

        "github.com/gin-gonic/gin"
)

func RegisterMiddleware(app *gin.Engine) func() {
        db := connection.InitMySQLSafe()
        cache := connection.InitRedisSafe()

        // Expose the global DB / cache handles to packages that can't import
        // `connection` directly (e.g. `utils`, which `connection` itself
        // imports).
        utils.GlobalDB = db
        utils.GlobalCache = cache

        app.Use(CORSMiddleware())
        app.Use(BuiltinMiddleWare(db, cache))
        app.Use(ThrottleMiddleware())
        app.Use(AuthMiddleware())

        return func() {
                db.Close()
                cache.Close()
        }
}
