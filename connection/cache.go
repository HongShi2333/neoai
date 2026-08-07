package connection

import (
        "context"
        "crypto/tls"
        "crypto/x509"
        "fmt"
        "os"
        "strings"

        "neoai/globals"

        "github.com/go-redis/redis/v8"
        "github.com/spf13/viper"
)

// CacheBackend identifies the type of cache backend in use.
// Both Valkey and Dragonfly are wire-protocol compatible with Redis,
// so the underlying client implementation is identical; the type only
// affects logging / diagnostics / healthcheck endpoints.
type CacheBackend string

const (
        BackendRedis    CacheBackend = "redis"
        BackendValkey   CacheBackend = "valkey"
        BackendDragonfly CacheBackend = "dragonfly"
)

var Cache *redis.Client
var ActiveBackend CacheBackend = BackendRedis

func InitRedisSafe() *redis.Client {
        ConnectRedis()

        // using Cache as a global variable to point to the latest redis connection
        RedisWorker(Cache)
        return Cache
}

// resolveBackend returns the configured cache backend, defaulting to redis.
// Recognised values (case-insensitive): "redis", "valkey", "dragonfly".
// An empty value keeps the historical default for backwards compatibility.
func resolveBackend() CacheBackend {
        raw := strings.ToLower(strings.TrimSpace(viper.GetString("cache.type")))
        if raw == "" {
                // Backwards-compat: existing deployments still use "redis" config block.
                raw = strings.ToLower(strings.TrimSpace(viper.GetString("redis.type")))
        }
        switch raw {
        case "valkey":
                return BackendValkey
        case "dragonfly":
                return BackendDragonfly
        case "redis":
                return BackendRedis
        default:
                // If `cache.type` is unspecified but `redis.host` is set, treat as redis.
                if raw == "" {
                        return BackendRedis
                }
                globals.Warn(fmt.Sprintf("[connection] unknown cache backend %q, falling back to redis", raw))
                return BackendRedis
        }
}

// readConnOptions reads the connection parameters from either the new
// `cache.*` config block or the legacy `redis.*` block, so existing
// config.yaml files keep working without modification.
func readConnOptions() (host string, port int, password string, db int) {
        if viper.IsSet("cache.host") || viper.IsSet("cache.port") {
                host = viper.GetString("cache.host")
                port = viper.GetInt("cache.port")
                password = viper.GetString("cache.password")
                db = viper.GetInt("cache.db")
                return
        }
        host = viper.GetString("redis.host")
        port = viper.GetInt("redis.port")
        password = viper.GetString("redis.password")
        db = viper.GetInt("redis.db")
        return
}

// readTLSOptions reads TLS-related config from either `cache.tls*` or
// `redis.tls*`. Returns nil if TLS is not enabled.
//
// Config keys:
//   cache.tls              (bool)   — enable TLS
//   cache.tls_skip_verify  (bool)   — skip cert verification (self-signed)
//   cache.tls_ca           (string) — path to CA cert file (PEM)
//
// Same keys under `redis.*` are also accepted for backwards compat.
func readTLSOptions() *tls.Config {
        tlsEnabled := viper.GetBool("cache.tls") || viper.GetBool("redis.tls")
        if !tlsEnabled {
                return nil
        }

        skipVerify := viper.GetBool("cache.tls_skip_verify") ||
                viper.GetBool("redis.tls_skip_verify")
        caPath := viper.GetString("cache.tls_ca")
        if caPath == "" {
                caPath = viper.GetString("redis.tls_ca")
        }

        cfg := &tls.Config{
                MinVersion:         tls.VersionTLS12,
                InsecureSkipVerify: skipVerify,
        }

        if caPath != "" {
                if caCert, err := os.ReadFile(caPath); err == nil {
                        caPool := x509.NewCertPool()
                        if caPool.AppendCertsFromPEM(caCert) {
                                cfg.RootCAs = caPool
                        } else {
                                globals.Warn(fmt.Sprintf("[connection] cache TLS CA file %s is not a valid PEM", caPath))
                        }
                } else {
                        globals.Warn(fmt.Sprintf("[connection] failed to read cache TLS CA file %s: %s", caPath, err.Error()))
                }
        }
        return cfg
}

func ConnectRedis() *redis.Client {
        backend := resolveBackend()
        ActiveBackend = backend

        host, port, password, db := readConnOptions()
        addr := fmt.Sprintf("%s:%d", host, port)

        // All three backends (Redis / Valkey / Dragonfly) speak RESP3/RESP2,
        // so a single redis.Client works for all of them. TLS is shared
        // across all three since the wire protocol is the same.
        opts := &redis.Options{
                Addr:     addr,
                Password: password,
                DB:       db,
        }
        if tlsCfg := readTLSOptions(); tlsCfg != nil {
                opts.TLSConfig = tlsCfg
        }
        Cache = redis.NewClient(opts)

        if err := pingRedis(Cache); err != nil {
                globals.Warn(
                        fmt.Sprintf(
                                "[connection] failed to connect to %s host: %s (message: %s), will retry in 5 seconds",
                                backend, addr, err.Error(),
                        ),
                )
        } else {
                globals.Debug(fmt.Sprintf("[connection] connected to %s (host: %s)", backend, addr))
        }

        if viper.GetBool("debug") {
                Cache.FlushAll(context.Background())
                globals.Debug(fmt.Sprintf("[connection] flush %s cache (host: %s)", backend, addr))
        }
        return Cache
}

// GetBackend returns the currently active cache backend identifier.
// Exposed so admin / health endpoints can report which store is in use.
func GetBackend() CacheBackend {
        return ActiveBackend
}
