package config

import (
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type JWTConfig struct {
	Token         string
	ExpiresIn     time.Duration
	SigningMethod *jwt.SigningMethodHMAC
}

type RedisConfig struct {
	Host        string
	Password    string
	Port        int
	User        string
	DialTimeout time.Duration
	ReadTimeout time.Duration
}

type Config struct {
	ServiceHost string
	ServicePort int
	JWT         JWTConfig
	Redis       RedisConfig
}

func NewConfig() (*Config, error) {
	var err error

	configName := "config"
	_ = godotenv.Load()
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("config")
	viper.AddConfigPath("../config")
	viper.AddConfigPath("../../config")
	viper.WatchConfig()

	if err = viper.ReadInConfig(); err != nil {
		log.Warnf("Config file not found, using environment variables: %v", err)
	}

	cfg := &Config{}

	// Базовые настройки
	cfg.ServiceHost = getEnv("SERVICE_HOST", "localhost")
	cfg.ServicePort, _ = strconv.Atoi(getEnv("SERVICE_PORT", "8080"))

	// JWT конфигурация (точно по методичке)
	cfg.JWT.Token = getEnv("JWT_SECRET", "test") // как в методичке "test"
	cfg.JWT.ExpiresIn = 3600 * time.Second       // как в методичке
	cfg.JWT.SigningMethod = jwt.SigningMethodHS256

	// Redis конфигурация
	cfg.Redis.Host = getEnv("REDIS_HOST", "localhost")
	cfg.Redis.Port, _ = strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "password") // как в методичке
	cfg.Redis.User = getEnv("REDIS_USER", "")
	cfg.Redis.DialTimeout = 10 * time.Second
	cfg.Redis.ReadTimeout = 10 * time.Second

	log.Info("config parsed successfully")

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
