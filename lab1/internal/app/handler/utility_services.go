package handler

// @title BITOP
// @version 1.0
// @description Bmstu Open IT Platform
// @host 127.0.0.1:8080
// @BasePath /
import (
	"lab1/internal/app/ds"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type UtilitiesPageData struct {
	Services      []ds.UtilityService
	SearchQuery   string
	Service       ds.UtilityService
	CartID        uint
	CartItems     []ds.UtilityApplicationService
	CartItemCount int
	Total         float32
	Address       string
}

func (h *Handler) GetUtilities(gCtx *gin.Context) {
	q := strings.TrimSpace(gCtx.Query("searchUtilities"))

	var services []ds.UtilityService
	var err error
	if q == "" {
		services, err = h.Repository.GetUtilityServices()
	} else {
		services, err = h.Repository.SearchUtilityServices(q)
	}
	if err != nil {
		log.Error(err)
		services = []ds.UtilityService{}
	}
	userID := uint(1)
	//userID := h.GetCurrentUserID(gCtx)                             // ИЗМЕНЕНО
	app, err := h.Repository.CreateOrGetUtilityApplication(userID) // ИЗМЕНЕНО
	if err == nil {
		app, err = h.Repository.GetUtilityApplicationByID(app.ID)
	}
	if err != nil {
		h.errorHandler(gCtx, http.StatusInternalServerError, err)
		return
	}

	gCtx.HTML(http.StatusOK, "list_of_utilities.html", gin.H{
		"data": UtilitiesPageData{
			Services:      services,
			SearchQuery:   q,
			CartID:        app.ID,
			CartItems:     app.Services,
			CartItemCount: len(app.Services),
			Total:         app.TotalCost,
		},
	})
}

func (h *Handler) GetUtility(gCtx *gin.Context) {
	idStr := gCtx.Param("id")
	raw, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || raw == 0 {
		gCtx.Redirect(http.StatusFound, "/utilities")
		return
	}

	svc, err := h.Repository.GetUtilityServiceByID(uint32(raw))
	if err != nil {
		gCtx.Redirect(http.StatusFound, "/utilities")
		return
	}

	userID := uint(1)                                              // ИЗМЕНЕНО
	app, err := h.Repository.CreateOrGetUtilityApplication(userID) // ИЗМЕНЕНО
	if err != nil {
		app = ds.UtilityApplication{}
	} else {
		app, err = h.Repository.GetUtilityApplicationByID(app.ID)
		if err != nil {
			app = ds.UtilityApplication{}
		}
	}

	gCtx.HTML(http.StatusOK, "utility.html", gin.H{
		"data": UtilitiesPageData{
			Service:       svc,
			CartID:        app.ID,
			CartItems:     app.Services,
			CartItemCount: len(app.Services),
			Total:         app.TotalCost,
		},
	})
}

// GetUtilityServicesAPI godoc
// @Summary Get utility services
// @Description Get paginated list of utility services
// @Tags utilities
// @Accept json
// @Produce json
// @Param title query string false "Filter by title"
// @Success 200 {object} ds.PaginatedResponse
// @Failure 500 {object} ds.ErrorResponse
// @Router /api/utilities [get]
func (h *Handler) GetUtilityServicesAPI(gCtx *gin.Context) {
	title := gCtx.Query("title")

	services, total, err := h.Repository.GetUtilityServicesFiltered(title)
	if err != nil {
		h.errorHandler(gCtx, http.StatusInternalServerError, err)
		return
	}

	var serviceDTOs []ds.UtilityServiceDTO
	for _, service := range services {
		imageURL := &service.ImageURL
		if service.ImageURL == "" {
			imageURL = nil
		}
		serviceDTOs = append(serviceDTOs, ds.UtilityServiceDTO{
			ID:          service.ID,
			Title:       service.Title,
			Description: service.Description,
			ImageURL:    imageURL,
			Unit:        service.Unit,
			Tariff:      service.Tariff,
		})
	}

	gCtx.JSON(http.StatusOK, ds.PaginatedResponse{
		Items: serviceDTOs,
		Total: total,
	})
}

// GetUtilityServiceAPI godoc
// @Summary Get utility service by ID
// @Description Get detailed information about a utility service
// @Tags utilities
// @Accept json
// @Produce json
// @Param id path int true "Service ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ds.ErrorResponse
// @Failure 404 {object} ds.ErrorResponse
// @Router /api/utilities/{id} [get]
func (h *Handler) GetUtilityServiceAPI(gCtx *gin.Context) {
	id, err := strconv.Atoi(gCtx.Param("id"))
	if err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	service, err := h.Repository.GetUtilityServiceByID(uint32(id))
	if err != nil {
		h.errorHandler(gCtx, http.StatusNotFound, err)
		return
	}

	imageURL := &service.ImageURL
	if service.ImageURL == "" {
		imageURL = nil
	}

	serviceDTO := ds.UtilityServiceDTO{
		ID:          service.ID,
		Title:       service.Title,
		Description: service.Description,
		ImageURL:    imageURL,
		Unit:        service.Unit,
		Tariff:      service.Tariff,
	}

	gCtx.JSON(http.StatusOK, serviceDTO)
}

// CreateUtilityService godoc
// @Summary Create utility service
// @Description Create new utility service (Admin/Manager only)
// @Tags utilities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ds.UtilityServiceCreateRequest true "Service data"
// @Success 201 {object} map[string]string
// @Failure 400 {object} ds.ErrorResponse
// @Failure 500 {object} ds.ErrorResponse
// @Router /api/utilities [post]
func (h *Handler) CreateUtilityService(gCtx *gin.Context) {
	var req ds.UtilityServiceCreateRequest
	if err := gCtx.BindJSON(&req); err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	imageURL := ""
	if req.ImageURL != nil {
		imageURL = *req.ImageURL
	}

	service := ds.UtilityService{
		Title:       req.Title,
		Description: req.Description,
		ImageURL:    imageURL,
		Unit:        req.Unit,
		Tariff:      req.Tariff,
	}

	if err := h.Repository.CreateUtilityService(&service); err != nil {
		h.errorHandler(gCtx, http.StatusInternalServerError, err)
		return
	}

	imageURLPtr := &service.ImageURL
	if service.ImageURL == "" {
		imageURLPtr = nil
	}

	serviceDTO := ds.UtilityServiceDTO{
		ID:          service.ID,
		Title:       service.Title,
		Description: service.Description,
		ImageURL:    imageURLPtr,
		Unit:        service.Unit,
		Tariff:      service.Tariff,
	}

	gCtx.JSON(http.StatusCreated, serviceDTO)
}

// UpdateUtilityService godoc
// @Summary Update utility service
// @Description Update existing utility service (Admin/Manager only)
// @Tags utilities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Service ID"
// @Param request body ds.UtilityServiceUpdateRequest true "Service data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ds.ErrorResponse
// @Failure 500 {object} ds.ErrorResponse
// @Router /api/utilities/{id} [put]
func (h *Handler) UpdateUtilityService(gCtx *gin.Context) {
	id, err := strconv.Atoi(gCtx.Param("id"))
	if err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	var req ds.UtilityServiceUpdateRequest
	if err := gCtx.BindJSON(&req); err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	service, err := h.Repository.UpdateUtilityService(uint32(id), req)
	if err != nil {
		h.errorHandler(gCtx, http.StatusInternalServerError, err)
		return
	}

	imageURL := &service.ImageURL
	if service.ImageURL == "" {
		imageURL = nil
	}

	serviceDTO := ds.UtilityServiceDTO{
		ID:          service.ID,
		Title:       service.Title,
		Description: service.Description,
		ImageURL:    imageURL,
		Unit:        service.Unit,
		Tariff:      service.Tariff,
	}

	gCtx.JSON(http.StatusOK, serviceDTO)
}

// DeleteUtilityService godoc
// @Summary Delete utility service
// @Description Delete utility service (Admin/Manager only)
// @Tags utilities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Service ID"
// @Success 204 {object} map[string]string
// @Failure 400 {object} ds.ErrorResponse
// @Failure 500 {object} ds.ErrorResponse
// @Router /api/utilities/{id} [delete]
func (h *Handler) DeleteUtilityService(gCtx *gin.Context) {
	id, err := strconv.Atoi(gCtx.Param("id"))
	if err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.DeleteUtilityService(uint32(id)); err != nil {
		h.errorHandler(gCtx, http.StatusInternalServerError, err)
		return
	}

	gCtx.JSON(http.StatusNoContent, gin.H{
		"message": "Услуга удалена",
	})
}

// UploadUtilityServiceImage godoc
// @Summary Upload utility service image
// @Description Upload/replace utility service image using minio (Admin/Manager only)
// @Tags utilities
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "Service ID"
// @Param file formData file true "Image file"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ds.ErrorResponse
// @Failure 500 {object} ds.ErrorResponse
// @Router /api/utilities/{id}/image [post]
func (h *Handler) UploadUtilityServiceImage(gCtx *gin.Context) {
	id, err := strconv.Atoi(gCtx.Param("id"))
	if err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	file, err := gCtx.FormFile("file")
	if err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	imageURL, err := h.Repository.UploadUtilityServiceImage(uint32(id), file)
	if err != nil {
		h.errorHandler(gCtx, http.StatusInternalServerError, err)
		return
	}

	gCtx.JSON(http.StatusOK, gin.H{"image_url": imageURL})
}

// AddServiceToDraft godoc
// @Summary Add service to draft application
// @Description Add service to draft application (application is created automatically)
// @Tags applications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_id path int true "Service ID"
// @Success 201 {object} ds.MessageResponse
// @Failure 400 {object} ds.ErrorResponse
// @Failure 500 {object} ds.ErrorResponse
// @Router /api/applications/draft/services/{service_id} [post]
func (h *Handler) AddServiceToDraft(gCtx *gin.Context) {
	serviceID, err := strconv.Atoi(gCtx.Param("service_id"))
	if err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	userID := h.GetCurrentUserID(gCtx) // ИЗМЕНЕНО

	app, err := h.Repository.CreateOrGetUtilityApplication(userID) // ИЗМЕНЕНО
	if err != nil {
		h.errorHandler(gCtx, http.StatusInternalServerError, err)
		return
	}

	if err := h.Repository.AddServiceToApplication(app.ID, uint32(serviceID), 1); err != nil {
		h.errorHandler(gCtx, http.StatusInternalServerError, err)
		return
	}

	gCtx.JSON(http.StatusCreated, gin.H{
		"message":        "Услуга добавлена в черновик заявки",
		"application_id": app.ID,
	})
}
