package handler

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

func (h *Handler) GetUtilities(ctx *gin.Context) {
	q := strings.TrimSpace(ctx.Query("searchUtilities"))

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

	userID := h.GetCurrentUserID()
	app, err := h.Repository.CreateOrGetUtilityApplication(userID)
	if err == nil {
		app, err = h.Repository.GetUtilityApplicationByID(app.ID)
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.HTML(http.StatusOK, "list_of_utilities.html", gin.H{
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

func (h *Handler) GetUtility(ctx *gin.Context) {
	idStr := ctx.Param("id")
	raw, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || raw == 0 {

		ctx.Redirect(http.StatusFound, "/utilities")
		return
	}

	svc, err := h.Repository.GetUtilityServiceByID(uint32(raw))
	if err != nil {

		ctx.Redirect(http.StatusFound, "/utilities")
		return
	}

	userID := h.GetCurrentUserID()
	app, err := h.Repository.CreateOrGetUtilityApplication(userID)
	if err != nil {

		app = ds.UtilityApplication{}
	} else {
		app, err = h.Repository.GetUtilityApplicationByID(app.ID)
		if err != nil {
			app = ds.UtilityApplication{}
		}
	}

	ctx.HTML(http.StatusOK, "utility.html", gin.H{
		"data": UtilitiesPageData{
			Service:       svc,
			CartID:        app.ID,
			CartItems:     app.Services,
			CartItemCount: len(app.Services),
			Total:         app.TotalCost,
		},
	})
}

func (h *Handler) GetUtilityServicesAPI(ctx *gin.Context) {
	title := ctx.Query("title")

	services, total, err := h.Repository.GetUtilityServicesFiltered(title)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
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

	ctx.JSON(http.StatusOK, ds.PaginatedResponse{
		Items: serviceDTOs,
		Total: total,
	})
}

func (h *Handler) GetUtilityServiceAPI(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	service, err := h.Repository.GetUtilityServiceByID(uint32(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
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

	ctx.JSON(http.StatusOK, serviceDTO)
}

func (h *Handler) CreateUtilityService(ctx *gin.Context) {
	var req ds.UtilityServiceCreateRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
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
		h.errorHandler(ctx, http.StatusInternalServerError, err)
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

	ctx.JSON(http.StatusCreated, serviceDTO)
}

func (h *Handler) UpdateUtilityService(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var req ds.UtilityServiceUpdateRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	service, err := h.Repository.UpdateUtilityService(uint32(id), req)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
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

	ctx.JSON(http.StatusOK, serviceDTO)
}

func (h *Handler) DeleteUtilityService(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.DeleteUtilityService(uint32(id)); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusNoContent, gin.H{
		"message": "Услуга удалена",
	})
}

func (h *Handler) UploadUtilityServiceImage(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	imageURL, err := h.Repository.UploadUtilityServiceImage(uint32(id), file)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"image_url": imageURL})
}

func (h *Handler) AddServiceToDraft(ctx *gin.Context) {
	serviceID, err := strconv.Atoi(ctx.Param("service_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	userID := h.GetCurrentUserID()

	app, err := h.Repository.CreateOrGetUtilityApplication(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if err := h.Repository.AddServiceToApplication(app.ID, uint32(serviceID), 1); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message":        "Услуга добавлена в черновик заявки",
		"application_id": app.ID,
	})
}
