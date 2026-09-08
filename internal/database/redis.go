package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/nexora/backend/internal/config"
)

var RDB *redis.Client

func ConnectRedis(cfg *config.Config) error {
	addr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Test connection with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("error al conectar a Redis: %w", err)
	}

	log.Println("✅ Conexión a Redis establecida")
	return nil
}

func CloseRedis() error {
	if RDB != nil {
		return RDB.Close()
	}
	return nil
}

func InvalidateUserCache(userID uint) {
	if RDB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		key := fmt.Sprintf("nexora:user:%d", userID)
		RDB.Del(ctx, key)
	}
}

// InvalidateCategoryCache invalida la caché del árbol/listado de categorías y detalles individuales
func InvalidateCategoryCache(categoryIDs ...uint) {
	if RDB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		keys := []string{"nexora:catalog:categories:all"}
		for _, cid := range categoryIDs {
			if cid > 0 {
				keys = append(keys, fmt.Sprintf("nexora:catalog:category:%d", cid))
			}
		}
		RDB.Del(ctx, keys...)
	}
}

// InvalidateProductCache invalida la caché de uno o varios productos
func InvalidateProductCache(productIDs ...uint) {
	if RDB != nil && len(productIDs) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		keys := make([]string, 0, len(productIDs))
		for _, pid := range productIDs {
			if pid > 0 {
				keys = append(keys, fmt.Sprintf("nexora:catalog:product:%d", pid))
			}
		}
		if len(keys) > 0 {
			RDB.Del(ctx, keys...)
		}
	}
}

