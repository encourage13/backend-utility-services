package main

import (
	"fmt"

	"lab1/internal/app/config"
	"lab1/internal/app/dsn"
	"lab1/internal/app/handler"
	"lab1/internal/app/repository"
	"lab1/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	_ = godotenv.Load()             // грузит .env из корня (если cwd=корень)
	_ = godotenv.Load("../../.env") // на случай запуска из подкаталога IDE

	router := gin.Default()
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println("DSN ->", postgresString) // можно удалить после проверки

	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	hand := handler.NewHandler(rep)
	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
