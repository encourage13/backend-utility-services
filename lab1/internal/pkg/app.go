package pkg

import (
	"fmt"

	"lab1/internal/app/config"
	"lab1/internal/app/handler"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Application struct {
	Config  *config.Config
	Router  *gin.Engine
	Handler *handler.Handler
}

func NewApp(c *config.Config, r *gin.Engine, h *handler.Handler) *Application {
	return &Application{
		Config:  c,
		Router:  r,
		Handler: h,
	}
}

func (a *Application) RunApp() {
	logrus.Info("Server start up")

	// регистрируем все роуты и статику через методы Handler
	a.Handler.RegisterHandler(a.Router)
	a.Handler.RegisterStatic(a.Router)

	host := a.Config.ServiceHost
	if host == "" {
		host = "0.0.0.0"
	}
	port := a.Config.ServicePort
	if port == 0 {
		port = 9000
	}

	serverAddress := fmt.Sprintf("%s:%d", host, port)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Server down")
}
