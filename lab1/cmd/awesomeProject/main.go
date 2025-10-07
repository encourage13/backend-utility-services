package main

import (
	"log"

	"lab1/internal/app/config"
	"lab1/internal/app/dsn"
	"lab1/internal/app/handler"
	"lab1/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	_ = godotenv.Load(".env")
	_ = godotenv.Load("../../.env")

	if _, err := config.NewConfig(); err != nil {
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

	r := gin.Default()
	h := handler.NewHandler(repo)
	h.RegisterHandler(r)
	h.RegisterStatic(r, "./templates")

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
