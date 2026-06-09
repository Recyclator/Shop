package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/time/rate"
	"github.com/nexora/backend/internal/database"
	"github.com/nexora/backend/internal/utils"
)

// clientLimiter mantiene un rate limiter por IP (para fallback en memoria)
type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	clients = make(map[string]*clientLimiter)
	mu      sync.RWMutex
)

// Cleanup antiguos limiters cada 5 minutos
func init() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 10*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

// RateLimiterMiddleware crea un middleware de rate limiting basado en IP.
// Utiliza Redis distribuido de forma preferente, con fallback en memoria local.
func RateLimiterMiddleware(maxRequests int, window time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()

		// 1. Intentar con Redis (Distribuido)
		if database.RDB != nil {
			ctx := c.Context()
			key := fmt.Sprintf("nexora:ratelimit:%s:%s", ip, c.Path())

			val, err := database.RDB.Incr(ctx, key).Result()
			if err == nil {
				if val == 1 {
					database.RDB.Expire(ctx, key, window)
				}
				if val > int64(maxRequests) {
					return utils.ErrorWithCode(c, fiber.StatusTooManyRequests, "RATE_LIMIT", "demasiadas solicitudes, intenta más tarde")
				}
				return c.Next()
			}
			// Si falla la comunicación con Redis, continúa silenciosamente al fallback en memoria
		}

		// 2. Fallback en Memoria Local (Token Bucket)
		mu.Lock()
		client, exists := clients[ip]
		if !exists {
			limiter := rate.NewLimiter(rate.Limit(float64(maxRequests)/window.Seconds()), maxRequests)
			client = &clientLimiter{
				limiter:  limiter,
				lastSeen: time.Now(),
			}
			clients[ip] = client
		}
		client.lastSeen = time.Now()
		mu.Unlock()

		if !client.limiter.Allow() {
			return utils.ErrorWithCode(c, fiber.StatusTooManyRequests, "RATE_LIMIT", "demasiadas solicitudes, intenta más tarde")
		}

		return c.Next()
	}
}
