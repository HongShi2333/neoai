package middleware

import (
        "neoai/utils"
        "fmt"
        "github.com/gin-gonic/gin"
        "github.com/go-redis/redis/v8"
        "strings"
)

type Limiter struct {
        Duration int
        Count    int64
}

func (l *Limiter) RateLimit(client *redis.Client, ip string, path string) (bool, error) {
        key := fmt.Sprintf("rate:%s:%s", path, ip)
        rate, err := utils.IncrWithLimit(client, key, 1, l.Count, int64(l.Duration))
        return !rate, err
}

var limits = map[string]Limiter{
        "/login":        {Duration: 10, Count: 20},
        "/register":     {Duration: 120, Count: 10},
        "/verify":       {Duration: 120, Count: 10},
        "/reset":        {Duration: 120, Count: 10},
        "/apikey":       {Duration: 1, Count: 2},
        "/resetkey":     {Duration: 3600, Count: 3},
        "/package":      {Duration: 1, Count: 2},
        "/quota":        {Duration: 1, Count: 2},
        "/buy":          {Duration: 1, Count: 2},
        "/subscribe":    {Duration: 1, Count: 2},
        "/subscription": {Duration: 1, Count: 2},
        "/chat":         {Duration: 1, Count: 5},
        "/conversation": {Duration: 1, Count: 5},
        "/invite":       {Duration: 7200, Count: 20},
        "/redeem":       {Duration: 1200, Count: 60},
        "/dashboard":    {Duration: 1, Count: 5},
        "/card":         {Duration: 1, Count: 5},
        "/generation":   {Duration: 1, Count: 5},
        "/article":      {Duration: 1, Count: 5},
        "/broadcast":    {Duration: 1, Count: 2},
}

func GetPrefixMap[T comparable](s string, p map[string]T) *T {
        // NeoAI: routes are always mounted under /api — strip the prefix
        // so the rate-limit rules (which use bare paths like "/login")
        // match regardless of serve_static mode.
        s = strings.TrimPrefix(s, "/api")

        for k, v := range p {
                if strings.HasPrefix(s, k) {
                        return &v
                }
        }
        return nil
}

func ThrottleMiddleware() gin.HandlerFunc {
        return func(c *gin.Context) {
                ip := c.ClientIP()
                path := c.Request.URL.Path
                cache := utils.GetCacheFromContext(c)

                limiter := GetPrefixMap[Limiter](path, limits)
                if limiter != nil && cache != nil {
                        rate, err := limiter.RateLimit(cache, ip, path)

                        // NeoAI: if the cache backend is unreachable, we fail OPEN —
                        // i.e. let the request through rather than blocking every
                        // login / register attempt with a "connection refused" error.
                        // Rate limiting is a soft protection; users being unable to
                        // log in because Redis is down is far worse than skipping the
                        // rate limit for the duration of the outage.
                        if err != nil {
                                // log and continue
                                fmt.Printf("[throttle] cache error, skipping rate limit: %s\n", err.Error())
                                c.Next()
                                return
                        }

                        if rate {
                                c.JSON(200, gin.H{
                                        "status": false,
                                        "reason": "You have sent too many requests. Please try again later.",
                                        "error":  "request_throttled",
                                })
                                c.Abort()
                                return
                        }
                }
                c.Next()
        }
}
