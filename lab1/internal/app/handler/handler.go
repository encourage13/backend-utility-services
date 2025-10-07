package handler

import (
	"lab1/internal/app/repository"
	"path/filepath"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{r}
}

func (h *Handler) RegisterHandler(r *gin.Engine) {

	r.GET("/", h.GetUtilities)
	r.GET("/utilities", h.GetUtilities)
	r.GET("/utilities/:id", h.GetUtility)

	r.GET("/utilities_application", h.GetUtilitiesApplication)
	r.GET("/utilities_application/:id", h.GetUtilitiesApplication)

	r.POST("/utilities_application/add", h.AddServiceInUtilityApplication)
	r.POST("/utilities_application/delete", h.DeleteUtilityApplication)
}

func (h *Handler) RegisterStatic(r *gin.Engine, templatesPath string) {
	r.LoadHTMLGlob(filepath.Join(templatesPath, "*.html"))
	r.Static("/static", "./recources")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorCode int, err error) {
	log.Error(err.Error())
	ctx.JSON(errorCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
