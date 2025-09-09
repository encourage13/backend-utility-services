package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"lab1/internal/app/handler"
	"lab1/internal/app/repository"
)

func Start() {
	repo, _ := repository.NewRepository()
	h := handler.New(repo)

	r := gin.Default()
	r.LoadHTMLGlob("../../templates/*")
	r.Static("/static", "../../recources")

	r.GET("/", h.Index)
	r.GET("/service/:id", h.Service)
	r.GET("/cart", h.Cart)

	addr := ":8080"
	logrus.Infof("listening on %s", addr)
	if err := r.Run(addr); err != nil {
		logrus.Fatal(err)
	}
}
