package handler

import (
	"lab1/internal/app/repository"
	"path/filepath"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const HardcodedUserID = 1

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{r}
}

func (h *Handler) GetCurrentUserID() uint {
	return HardcodedUserID
}

func (h *Handler) RegisterHandler(r *gin.Engine) {
	// HTML роуты для фронтенда
	r.GET("/", h.GetUtilities)
	r.GET("/utilities", h.GetUtilities)
	r.GET("/utilities/:id", h.GetUtility)
	r.GET("/utilities_application", h.GetUtilitiesApplication)
	r.GET("/utilities_application/:id", h.GetUtilitiesApplication)
	r.POST("/utilities_application/add", h.AddServiceInUtilityApplication)
	r.POST("/utilities_application/delete", h.DeleteUtilityApplication)

	// API роуты
	api := r.Group("/api")
	{
		// Домен коммунальных услуг
		api.GET("/utilities", h.GetUtilityServicesAPI)
		api.GET("/utilities/:id", h.GetUtilityServiceAPI)
		api.POST("/utilities", h.CreateUtilityService)
		api.PUT("/utilities/:id", h.UpdateUtilityService)
		api.DELETE("/utilities/:id", h.DeleteUtilityService)
		api.POST("/utilities/:id/image", h.UploadUtilityServiceImage)
		api.POST("/applications/draft/services/:service_id", h.AddServiceToDraft)

		// Домен заявок
		api.GET("/applications/cart", h.GetCartBadge)
		api.GET("/applications", h.ListApplications)
		api.GET("/applications/:id", h.GetApplication)
		api.PUT("/applications/:id", h.UpdateApplication)
		api.PUT("/applications/:id/form", h.FormApplication)
		api.PUT("/applications/:id/resolve", h.ResolveApplication)
		api.DELETE("/applications/:id", h.DeleteApplication)

		// Домен связи заявок и услуг
		api.DELETE("/applications/:id/services/:service_id", h.RemoveServiceFromApplication)
		api.PUT("/applications/:id/services/:service_id", h.UpdateApplicationService)

		// Домен пользователей
		api.POST("/users", h.Register)
		api.GET("/users/:id", h.GetUserData)
		api.PUT("/users/:id", h.UpdateUserData)
		api.POST("/auth/login", h.Login)
		api.POST("/auth/logout", h.Logout)
	}
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
