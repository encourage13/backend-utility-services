package handler

// @title BITOP
// @version 1.0
// @description Bmstu Open IT Platform
// @host localhost:8080
// @BasePath /
import (
	"lab1/internal/app/config"
	"lab1/internal/app/redis"
	"lab1/internal/app/repository"
	"lab1/internal/app/role"
	"path/filepath"

	_ "lab1/docs" // импортируйте docs папку (замените lab1 на имя вашего модуля)

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	Repository *repository.Repository
	Redis      *redis.Client
	Config     *config.Config
}

func NewHandler(r *repository.Repository, redisClient *redis.Client, cfg *config.Config) *Handler {
	return &Handler{
		Repository: r,
		Redis:      redisClient,
		Config:     cfg,
	}
}

// GetCurrentUserID получает ID пользователя из контекста gin
func (h *Handler) GetCurrentUserID(gCtx *gin.Context) uint {
	userID, exists := gCtx.Get("userID")
	if !exists {
		return 0
	}
	return userID.(uint)
}

// GetCurrentUserRole получает роль пользователя из контекста gin
func (h *Handler) GetCurrentUserRole(gCtx *gin.Context) role.Role {
	userRole, exists := gCtx.Get("userRole")
	if !exists {
		return role.Buyer
	}
	return userRole.(role.Role)
}

// GetCurrentUserUUID - временная заглушка
func (h *Handler) GetCurrentUserUUID() uuid.UUID {
	return uuid.Nil
}

func (h *Handler) RegisterHandler(r *gin.Engine) {
	// HTML роуты для фронтенда (публичные)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/", h.GetUtilities)
	r.GET("/utilities", h.GetUtilities)
	r.GET("/utilities/:id", h.GetUtility)
	r.GET("/utilities_application", h.GetUtilitiesApplication)
	r.GET("/utilities_application/:id", h.GetUtilitiesApplication)

	// API роуты
	api := r.Group("/api")
	{
		// Публичные эндпоинты (не требуют аутентизации)
		api.POST("/sign_up", h.Register)
		api.POST("/login", h.Login)

		// Защищенные эндпоинты (требуют JWT токен)
		protected := api.Group("")
		protected.Use(h.WithAuthCheck()) // применяем аутентификацию ко всем защищенным эндпоинтам
		{
			// Аутентификация
			protected.POST("/logout", h.Logout)

			// Коммунальные услуги (доступны всем аутентифицированным пользователям)
			protected.GET("/utilities", h.GetUtilityServicesAPI)
			protected.GET("/utilities/:id", h.GetUtilityServiceAPI)
			protected.POST("/applications/draft/services/:service_id", h.AddServiceToDraft)
			protected.GET("/applications/cart", h.GetCartBadge)

			// Заявки пользователя
			protected.GET("/applications", h.ListApplications)
			protected.GET("/applications/:id", h.GetApplication)
			protected.PUT("/applications/:id", h.UpdateApplication)
			protected.PUT("/applications/:id/form", h.FormApplication)
			protected.DELETE("/applications/:id", h.DeleteApplication)
			protected.DELETE("/applications/:id/services/:service_id", h.RemoveServiceFromApplication)
			protected.PUT("/applications/:id/services/:service_id", h.UpdateApplicationService)

			// Административные эндпоинты (только для модераторов и админов)
			admin := protected.Group("")
			admin.Use(h.WithAuthCheck(role.Manager, role.Admin))
			{
				admin.POST("/utilities", h.CreateUtilityService)
				admin.PUT("/utilities/:id", h.UpdateUtilityService)
				admin.DELETE("/utilities/:id", h.DeleteUtilityService)
				admin.POST("/utilities/:id/image", h.UploadUtilityServiceImage)
				admin.PUT("/applications/:id/resolve", h.ResolveApplication)
			}
		}
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
