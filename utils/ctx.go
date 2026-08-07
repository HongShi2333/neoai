package utils

import (
        "database/sql"

        "github.com/gin-gonic/gin"
        "github.com/go-redis/redis/v8"
)

func GetDBFromContext(c *gin.Context) *sql.DB {
        return c.MustGet("db").(*sql.DB)
}

func GetCacheFromContext(c *gin.Context) *redis.Client {
        return c.MustGet("cache").(*redis.Client)
}

// GetDBFromContextSafe returns the global DB instance without requiring a
// gin.Context. Used during boot (e.g. running migrations before the HTTP
// server starts listening).
//
// We use a function variable to break what would otherwise be an import
// cycle: utils -> connection -> utils. The connection package installs
// its DB pointer into `GlobalDB` after init.
var GlobalDB *sql.DB

// GetCacheFromContextSafe returns the global cache client without requiring
// a gin.Context.
var GlobalCache *redis.Client

func GetDBFromContextSafe() *sql.DB {
        return GlobalDB
}

func GetCacheFromContextSafe() *redis.Client {
        return GlobalCache
}

func GetUserFromContext(c *gin.Context) string {
        return c.MustGet("user").(string)
}

func GetAdminFromContext(c *gin.Context) bool {
        v, exists := c.Get("admin")
        if !exists {
                return false
        }
        return v.(bool)
}

func GetAgentFromContext(c *gin.Context) string {
        return c.MustGet("agent").(string)
}
