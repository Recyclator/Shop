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
		ctx := context.Background()
		key := fmt.Sprintf("nexora:user:%d", userID)
		RDB.Del(ctx, key)
	}
}

