package utils

import (
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// blacklistEntry almacena un JTI con su expiración
type blacklistEntry struct {
	expiresAt time.Time
}

var (
	blacklist = make(map[string]blacklistEntry)
	blackMu   sync.RWMutex
)

// init limpia entradas expiradas cada 60 minutos
func init() {
	go func() {
		ticker := time.NewTicker(60 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			cleanExpiredEntries()
		}
	}()
}

// RevokeToken revoca un token por su JTI (JWT ID) hasta su expiración
func RevokeToken(jti string, expiresAt time.Time) {
	blackMu.Lock()
	defer blackMu.Unlock()
	blacklist[jti] = blacklistEntry{expiresAt: expiresAt}
}

// IsRevoked verifica si un token ha sido revocado
func IsRevoked(jti string) bool {
	blackMu.RLock()
	defer blackMu.RUnlock()
	entry, exists := blacklist[jti]
	if !exists {
		return false
	}
	// Si ya expiró, podemos eliminarlo de la blacklist
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
		jti = uuid.New().String() // fallback
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

// cleanExpiredEntries limpia entradas expiradas de la blacklist
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
