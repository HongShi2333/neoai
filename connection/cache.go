package connection

import (
	"chat/globals"
	"context"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

var Cache *redis.Client

// Cache driver type constants.
// Valkey and Dragonfly are wire-compatible with the Redis protocol,
// so they all use the go-redis client; the driver is mainly used for
// logging and future feature gating.
const (
	RedisDriver     = "redis"
	ValkeyDriver    = "valkey"
	DragonflyDriver = "dragonfly"
)

// GetCacheDriver reads the configured cache driver (redis / valkey / dragonfly).
// Defaults to "redis" for backward compatibility.
func GetCacheDriver() string {
	driver := strings.ToLower(strings.TrimSpace(viper.GetString("redis.type")))
	if driver == "" {
		driver = strings.ToLower(strings.TrimSpace(viper.GetString("cache.type")))
	}
	switch driver {
	case ValkeyDriver, DragonflyDriver:
		return driver
	default:
		return RedisDriver
	}
}

func InitRedisSafe() *redis.Client {
	ConnectRedis()

	// using Cache as a global variable to point to the latest redis connection
	RedisWorker(Cache)
	return Cache
}

func ConnectRedis() *redis.Client {
	// connect to redis / valkey / dragonfly (all wire-compatible with redis protocol)
	Cache = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", viper.GetString("redis.host"), viper.GetInt("redis.port")),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	})

	driver := GetCacheDriver()

	if err := pingRedis(Cache); err != nil {
		globals.Warn(
			fmt.Sprintf(
				"[connection] failed to connect to %s host: %s (message: %s), will retry in 5 seconds",
				driver,
				viper.GetString("redis.host"),
				err.Error(),
			),
		)
	} else {
		globals.Debug(fmt.Sprintf("[connection] connected to %s (host: %s)", driver, viper.GetString("redis.host")))
	}

	if viper.GetBool("debug") {
		Cache.FlushAll(context.Background())
		globals.Debug(fmt.Sprintf("[connection] flush %s cache (host: %s)", driver, viper.GetString("redis.host")))
	}
	return Cache
}
