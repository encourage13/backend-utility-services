package main

import (
	"log"

	"lab1/internal/app/ds"
	"lab1/internal/app/dsn"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// грузим .env из текущей и из корневой директории
	_ = godotenv.Load()
	_ = godotenv.Load("../../.env")

	d := dsn.FromEnv() // паникует, если переменные неполные
	db, err := gorm.Open(postgres.Open(d), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Panicf("failed to connect database: %v", err)
	}

	// мигрируем нужные сущности; добавь остальные по месту
	if err := db.AutoMigrate(
		&ds.User{},
		&ds.Service{},
		&ds.Request{},
		&ds.RequestService{},
	); err != nil {
		log.Panic("cant migrate db: ", err)
	}
	log.Println("migrate: OK")
}
