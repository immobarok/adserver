package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"adserver/config"
	"adserver/db"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	_ = godotenv.Load()
	config.Load()

	rdb := redis.NewClient(&redis.Options{
		Addr:     config.App.RedisAddr,
		Username: config.App.RedisUsername,
		Password: config.App.RedisPassword,
		DB:       config.App.RedisDB,
	})
	db.RedisClient = rdb

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("❌ Redis PING failed: %v", err)
	}
	fmt.Printf("✅ Redis connected! PING → %s\n", pong)

	_ = rdb.Set(ctx, "adserver:test", "hello-adserver", 10*time.Second)
	val, _ := rdb.Get(ctx, "adserver:test").Result()
	fmt.Printf("✅ SET/GET smoke test OK: %q\n", val)
}
