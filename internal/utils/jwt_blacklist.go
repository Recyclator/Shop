package utils

import (
	"context"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// blacklistEntry almacena un JTI con su expiración
type blacklistEntry struct {
	expiresAt time.Time
}

var (
	redisClient *redis.Client
	blacklist   = make(map[string]blacklistEntry)
	blackMu     sync.RWMutex
)

// SetBlacklistRedisClient configura el cliente Redis para la lista negra distribuida
func SetBlacklistRedisClient(client *redis.Client) {
	blackMu.Lock()
	defer blackMu.Unlock()
	redisClient = client
}

// init limpia entradas expiradas cada 60 minutos de la memoria local
func init() {
	go func() {
		ticker := time.NewTicker(60 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			cleanExpiredEntries()
		}
	}()
}

// RevokeToken revoca un token por su JTI (JWT ID) hasta su expiración en Redis y memoria local
func RevokeToken(jti string, expiresAt time.Time) {
	if jti == "" {
		return
	}
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return
	}

	blackMu.RLock()
	client := redisClient
	blackMu.RUnlock()

	if client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = client.Set(ctx, "nexora:blacklist:"+jti, "1", ttl).Err()
	}

	blackMu.Lock()
	defer blackMu.Unlock()
	blacklist[jti] = blacklistEntry{expiresAt: expiresAt}
}

// IsRevoked verifica si un token ha sido revocado consultando Redis (o memoria local)
func IsRevoked(jti string) bool {
	if jti == "" {
		return false
	}

	blackMu.RLock()
	client := redisClient
	blackMu.RUnlock()

	if client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		exists, err := client.Exists(ctx, "nexora:blacklist:"+jti).Result()
		if err == nil && exists > 0 {
			return true
		}
	}

	blackMu.RLock()
	defer blackMu.RUnlock()
	entry, exists := blacklist[jti]
	if !exists {
		return false
	}
	if time.Now().After(entry.expiresAt) {
		return false
	}
	return true
}

// RevokeTokenString revoca un token JWT completo (extrae JTI)
func RevokeTokenString(tokenString string) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return
	}
	jti, _ := claims["jti"].(string)
	if jti == "" {
		return
	}
	expFloat, _ := claims["exp"].(float64)
	var expiresAt time.Time
	if expFloat > 0 {
		expiresAt = time.Unix(int64(expFloat), 0)
	} else {
		expiresAt = time.Now().Add(time.Hour * 24 * 30) // 30 días por defecto
	}
	RevokeToken(jti, expiresAt)
}

// cleanExpiredEntries limpia entradas expiradas de la memoria local
func cleanExpiredEntries() {
	blackMu.Lock()
	defer blackMu.Unlock()
	now := time.Now()
	for jti, entry := range blacklist {
		if now.After(entry.expiresAt) {
			delete(blacklist, jti)
		}
	}
}
