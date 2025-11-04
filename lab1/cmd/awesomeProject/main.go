package main

import (
	"context"
	"log"

	"lab1/internal/app/config"
	"lab1/internal/app/dsn"
	"lab1/internal/app/handler"
	"lab1/internal/app/redis"
	"lab1/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// @title BITOP
// @version 1.0
// @description Bmstu Open IT Platform

// @contact.name API Support
// @contact.url https://vk.com/bmstu_schedule
// @contact.email bitop@spatecon.ru

// @license.name AS IS (NO WARRANTY)

// @host localhost:8080
// @schemes https http
// @BasePath /

func main() {

	_ = godotenv.Load(".env")
	_ = godotenv.Load("../../.env")

	cfg, err := config.NewConfig()
	if err != nil {
		log.Println("config warning:", err)
	}

	d := dsn.FromEnv()
	log.Println("APP DSN =", d)
	if d == "" {
		log.Fatal("empty DSN: проверь .env и dsn.FromEnv()")
	}

	repo, err := repository.NewRepository(d)
	if err != nil {
		log.Fatal("error initializing repository:", err)
	}

	// Создаем Redis клиент с правильным названием функции
	redisClient, err := redis.New(context.Background(), cfg.Redis)
	if err != nil {
		log.Fatal("error initializing redis:", err)
	}
	defer redisClient.Close()

	r := gin.Default()
	// Передаем все необходимые аргументы
	h := handler.NewHandler(repo, redisClient, cfg)
	h.RegisterHandler(r)
	h.RegisterStatic(r, "./templates")

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
