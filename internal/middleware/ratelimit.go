package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/time/rate"
	"github.com/nexora/backend/internal/utils"
)

// clientLimiter mantiene un rate limiter por IP
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

// RateLimiterMiddleware crea un middleware de rate limiting basado en IP
// usando token bucket (golang.org/x/time/rate)
func RateLimiterMiddleware(maxRequests int, window time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()

		mu.Lock()
		client, exists := clients[ip]
		if !exists {
			// Configurar limiter: maxRequests por window
			// rate.Limit es maxRequests/second equivalente
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
			return utils.ErrorWithCode(c, fiber.StatusTooManyRequests, "RATE_LIMIT", "demasiadas solicitudes, intenta m\u00e1s tarde")
		}

		return c.Next()
	}
}
