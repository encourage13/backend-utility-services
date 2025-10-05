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

	r.GET("/", h.GetUtilities)

	r.GET("/utilities", h.GetUtilities)
	r.GET("/utilities/:id", h.GetUtility)
	r.GET("/utilities_application", h.GetUtilitiesApplication)
	r.GET("/utilities_application/:id", h.GetUtilitiesApplication)

	addr := ":8080"
	logrus.Infof("listening on %s", addr)
	if err := r.Run(addr); err != nil {
		logrus.Fatal(err)
	}
}
