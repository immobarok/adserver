package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"adserver/config"
)

var (
	DB          *sql.DB
	RedisClient *redis.Client
)

// ConnectPostgres opens and validates the PostgreSQL connection pool.
func ConnectPostgres() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.App.DBHost,
		config.App.DBPort,
		config.App.DBUser,
		config.App.DBPassword,
		config.App.DBName,
		config.App.DBSSLMode,
	)

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("❌ Failed to open DB: %v", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(10)
	DB.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = DB.PingContext(ctx); err != nil {
		log.Fatalf("❌ Cannot reach PostgreSQL: %v", err)
	}

	log.Println("✅ PostgreSQL connected")
}

// ConnectRedis opens and validates the Redis connection.
func ConnectRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.App.RedisAddr,
		Password: config.App.RedisPassword,
		DB:       config.App.RedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := RedisClient.Ping(ctx).Result(); err != nil {
		log.Printf("⚠️  Redis not available (%v) – caching disabled", err)
		RedisClient = nil
		return
	}

	log.Println("✅ Redis connected")
}
