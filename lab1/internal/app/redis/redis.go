package redis

import (
	"context"
	"fmt"
	"lab1/internal/app/config"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

const servicePrefix = "utility_service."

type Client struct {
	cfg    config.RedisConfig
	client *redis.Client
}

func New(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	client := &Client{cfg: cfg}

	redisClient := redis.NewClient(&redis.Options{
		Addr:        cfg.Host + ":" + strconv.Itoa(cfg.Port),
		Password:    cfg.Password,
		DB:          0,
		DialTimeout: cfg.DialTimeout,
		ReadTimeout: cfg.ReadTimeout,
	})

	client.client = redisClient

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("cant ping redis: %w", err)
	}

	return client, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

const jwtPrefix = "jwt."

func getJWTKey(token string) string {
	return servicePrefix + jwtPrefix + token
}

func (c *Client) WriteJWTToBlacklist(ctx context.Context, jwtStr string, ttl time.Duration) error {
	key := getJWTKey(jwtStr)
	log.Printf("Writing to blacklist: key=%s, ttl=%v", key, ttl)

	err := c.client.Set(ctx, key, "blacklisted", ttl).Err()
	if err != nil {
		log.Printf("Redis set error: %v", err)
		return err
	}

	// Проверяем что записалось
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		log.Printf("Redis get error: %v", err)
		return err
	}

	log.Printf("Blacklist verification: key=%s, value=%s", key, val)
	return nil
}

func (c *Client) CheckJWTInBlacklist(ctx context.Context, jwtStr string) error {
	key := getJWTKey(jwtStr)
	log.Printf("Checking blacklist: key=%s", key)

	_, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		log.Printf("Token NOT in blacklist: %s", key)
		return redis.Nil
	}
	if err != nil {
		log.Printf("Redis check error: %v", err)
		return err
	}

	log.Printf("Token FOUND in blacklist: %s", key)
	return nil
}
