package middleware

import (
	"database/sql"
	"net/http"

	"neoai/connection"
	"neoai/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// health.go — lightweight health-check endpoints for split / k8s deployments.
//
//   GET /healthz       — liveness probe (always 200 if the process is up)
//   GET /ready         — readiness probe (200 only if DB + cache are reachable)
//
// Both endpoints are public (no auth middleware) so they can be hit by
// external load balancers without first obtaining a JWT.

func HealthzAPI(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "neoai",
	})
}

func ReadyAPI(c *gin.Context) {
	db := utils.GlobalDB
	cache := utils.GlobalCache

	status := gin.H{
		"db":     "down",
		"cache":  "down",
		"backend": connection.GetBackend(),
	}
	httpStatus := http.StatusOK

	if db != nil && db.Ping() == nil {
		status["db"] = "up"
	} else {
		httpStatus = http.StatusServiceUnavailable
	}

	if cache != nil && pingCache(cache) {
		status["cache"] = "up"
	} else {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, status)
}

func pingCache(cache *redis.Client) bool {
	if cache == nil {
		return false
	}
	if _, err := cache.Ping(cache.Context()).Result(); err != nil {
		return false
	}
	return true
}

// RegisterHealth wires the health endpoints onto the engine.
// Public paths — intentionally NOT behind the AuthMiddleware.
func RegisterHealth(engine *gin.Engine) {
	engine.GET("/healthz", HealthzAPI)
	engine.GET("/ready", ReadyAPI)
}

// silence unused import warning if sql isn't referenced elsewhere
var _ = sql.ErrConnDone
