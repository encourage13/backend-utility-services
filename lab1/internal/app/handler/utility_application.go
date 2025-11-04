package handler

// @title BITOP
// @version 1.0
// @description Bmstu Open IT Platform
// @host localhost:8080
// @BasePath /
import (
	"fmt"
	"lab1/internal/app/ds"
	"lab1/internal/app/role"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetUtilitiesApplication godoc
// @Summary Get utility application
// @Description Get utility application by ID or create new draft
// @Tags applications
// @Produce html
// @Param id path string false "Application ID"
// @Success 200
// @Router /utilities_application/{id} [get]
func (h *Handler) GetUtilitiesApplication(ggCtx *gin.Context) {
	idStr := strings.TrimSpace(ggCtx.Param("id"))
	var (
		app ds.UtilityApplication
		err error
	)

	userID := uint(1) // ИЗМЕНЕНО: используем ID вместо UUID

	if idStr == "" {
		app, err = h.Repository.CreateOrGetUtilityApplication(userID) // ИЗМЕНЕНО
		if err != nil {
			ggCtx.Redirect(http.StatusFound, "/utilities")
			return
		}
	} else {
		id, convErr := strconv.Atoi(idStr)
		if convErr != nil {
			ggCtx.Redirect(http.StatusFound, "/utilities")
			return
		}

		app, err = h.Repository.GetUtilityApplicationByID(uint(id))
		if err != nil {
			ggCtx.Redirect(http.StatusFound, "/utilities")
			return
		}

		if app.Status == string(ds.DELETED) {
			ggCtx.Redirect(http.StatusFound, "/utilities")
			return
		}
		if app.TotalCost == 0 {
			ggCtx.Redirect(http.StatusFound, "/utilities")
			return
		}
		if len(app.Services) == 0 {
			ggCtx.Redirect(http.StatusFound, "/utilities")
			return
		}
	}

	ggCtx.HTML(http.StatusOK, "utilities_application.html", gin.H{
		"data": UtilitiesPageData{
			CartID:        app.ID,
			CartItems:     app.Services,
			CartItemCount: len(app.Services),
			Total:         app.TotalCost,
			Address:       app.Address,
		},
	})
}

func (h *Handler) AddServiceInUtilityApplication(gCtx *gin.Context) {
	rawID := gCtx.PostForm("service_id")
	serviceID, err := strconv.Atoi(rawID)
	if err != nil || serviceID <= 0 {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	qty := float32(1)
	if rawQty := strings.TrimSpace(gCtx.PostForm("quantity")); rawQty != "" {
		if f, convErr := strconv.ParseFloat(rawQty, 32); convErr == nil && f > 0 {
			qty = float32(f)
		}
	}

	userID := uint(1)                                              // ИЗМЕНЕНО
	app, err := h.Repository.CreateOrGetUtilityApplication(userID) // ИЗМЕНЕНО
	if err != nil {
		h.errorHandler(gCtx, http.StatusInternalServerError, err)
		return
	}

	if err := h.Repository.AddServiceToApplication(app.ID, uint32(serviceID), qty); err != nil {
		h.errorHandler(gCtx, http.StatusInternalServerError, err)
		return
	}

	gCtx.Redirect(http.StatusMovedPermanently, "/utilities?searchUtilities="+gCtx.PostForm("searchUtilities"))
}

func (h *Handler) DeleteUtilityApplication(gCtx *gin.Context) {
	idStr := gCtx.PostForm("application_id")
	appID, err := strconv.Atoi(idStr)
	if err != nil || appID <= 0 {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.DeleteUtilityApplication(uint(appID)); err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	gCtx.Redirect(http.StatusMovedPermanently, "/utilities")
}

// GetCartBadge godoc
// @Summary Get cart badge
// @Description Get draft application ID and service count for current user
// @Tags applications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/applications/cart [get]
func (h *Handler) GetCartBadge(gCtx *gin.Context) {
	userID := h.GetCurrentUserID(gCtx)                     // ИЗМЕНЕНО
	draft, err := h.Repository.GetDraftApplication(userID) // ИЗМЕНЕНО
	if err != nil {
		gCtx.JSON(http.StatusOK, ds.CartBadgeDTO{
			ApplicationID: nil,
			Count:         0,
		})
		return
	}

	fullApplication, err := h.Repository.GetApplicationWithServices(draft.ID)
	if err != nil {
		h.errorHandler(gCtx, http.StatusInternalServerError, err)
		return
	}

	gCtx.JSON(http.StatusOK, ds.CartBadgeDTO{
		ApplicationID: &fullApplication.ID,
		Count:         len(fullApplication.Services),
	})
}

// ListApplications godoc
// @Summary Get filtered list of utility applications
// @Description Get filtered list of applications (excluding deleted and drafts) with date range and status filters. Requires JWT token.
// @Tags applications
// @Accept json
// @Produce json
// @Param authorization header string true "JWT Token in format: Bearer {token}"
// @Param status query string false "Status filter: DRAFT, FORMED, COMPLETED, REJECTED"
// @Param from query string false "Date from filter (YYYY-MM-DD)"
// @Param to query string false "Date to filter (YYYY-MM-DD)"
// @Success 200 {array} ds.UtilityApplicationDTO
// @Failure 401 {object} ds.ErrorResponse
// @Failure 403 {object} ds.ErrorResponse
// @Failure 500 {object} ds.ErrorResponse
// @Router /api/applications [get]
func (h *Handler) ListApplications(gCtx *gin.Context) {
	status := gCtx.Query("status")
	from := gCtx.Query("from")
	to := gCtx.Query("to")

	applications, err := h.Repository.GetApplicationsFiltered(status, from, to)
	if err != nil {
		h.errorHandler(gCtx, http.StatusInternalServerError, err)
		return
	}

	var applicationDTOs []ds.UtilityApplicationDTO
	for _, app := range applications {
		dto := ds.UtilityApplicationDTO{
			ID:           app.ID,
			UserID:       app.UserID, // ИЗМЕНЕНО
			Status:       app.Status,
			TotalCost:    app.TotalCost,
			DateCreated:  app.DateCreated,
			DateFormed:   &app.DateFormed,
			DateAccepted: &app.DateAccepted,
			ModeratorID:  app.ModeratorID,
		}

		if app.DateFormed.IsZero() {
			dto.DateFormed = nil
		}
		if app.DateAccepted.IsZero() {
			dto.DateAccepted = nil
		}

		// Используем правильные поля из структуры User
		if app.User.ID != 0 { // Проверяем что User не пустой
			isModerator := app.User.Role == role.Manager || app.User.Role == role.Admin
			dto.User = &ds.UserDTO{
				ID:          app.User.ID, // ИЗМЕНЕНО
				Login:       app.User.Name,
				IsModerator: isModerator,
			}
		}

		if app.ModeratorID != nil && app.Moderator.ID != 0 { // Проверяем что Moderator не пустой
			isModerator := app.Moderator.Role == role.Manager || app.Moderator.Role == role.Admin
			dto.Moderator = &ds.UserDTO{
				ID:          app.Moderator.ID, // ИЗМЕНЕНО
				Login:       app.Moderator.Name,
				IsModerator: isModerator,
			}
		}

		applicationDTOs = append(applicationDTOs, dto)
	}

	gCtx.JSON(http.StatusOK, applicationDTOs)
}

// GetApplication godoc
// @Summary Get application by ID
// @Description Get detailed application information with services
// @Tags applications
// @Accept json
// @Produce json
// @Param id path int true "Application ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ds.ErrorResponse
// @Failure 404 {object} ds.ErrorResponse
// @Router /api/applications/{id} [get]
func (h *Handler) GetApplication(gCtx *gin.Context) {
	id, err := strconv.Atoi(gCtx.Param("id"))
	if err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	application, err := h.Repository.GetApplicationWithServices(uint(id))
	if err != nil {
		h.errorHandler(gCtx, http.StatusNotFound, err)
		return
	}

	var services []ds.UtilityApplicationServiceDTO
	for _, service := range application.Services {
		imageURL := &service.Service.ImageURL
		if service.Service.ImageURL == "" {
			imageURL = nil
		}

		services = append(services, ds.UtilityApplicationServiceDTO{
			UtilityApplicationID: service.UtilityApplicationID,
			UtilityServiceID:     service.UtilityServiceID,
			Quantity:             service.Quantity,
			Total:                service.Total,
			CustomTariff:         service.CustomTariff,
			Service: &ds.UtilityServiceDTO{
				ID:          service.Service.ID,
				Title:       service.Service.Title,
				Description: service.Service.Description,
				ImageURL:    imageURL,
				Unit:        service.Service.Unit,
				Tariff:      service.Service.Tariff,
			},
		})
	}

	applicationDTO := ds.UtilityApplicationDTO{
		ID:           application.ID,
		UserID:       application.UserID, // ИЗМЕНЕНО
		Status:       application.Status,
		TotalCost:    application.TotalCost,
		DateCreated:  application.DateCreated,
		DateFormed:   &application.DateFormed,
		DateAccepted: &application.DateAccepted,
		ModeratorID:  application.ModeratorID,
		Services:     services,
	}

	if application.DateFormed.IsZero() {
		applicationDTO.DateFormed = nil
	}
	if application.DateAccepted.IsZero() {
		applicationDTO.DateAccepted = nil
	}

	// Используем правильные поля из структуры User
	if application.User.ID != 0 {
		isModerator := application.User.Role == role.Manager || application.User.Role == role.Admin
		applicationDTO.User = &ds.UserDTO{
			ID:          application.User.ID, // ИЗМЕНЕНО
			Login:       application.User.Name,
			IsModerator: isModerator,
		}
	}

	if application.ModeratorID != nil && application.Moderator.ID != 0 {
		isModerator := application.Moderator.Role == role.Manager || application.Moderator.Role == role.Admin
		applicationDTO.Moderator = &ds.UserDTO{
			ID:          application.Moderator.ID, // ИЗМЕНЕНО
			Login:       application.Moderator.Name,
			IsModerator: isModerator,
		}
	}

	gCtx.JSON(http.StatusOK, applicationDTO)
}

// UpdateApplication godoc
// @Summary Update application
// @Description Update application data
// @Tags applications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Application ID"
// @Param request body ds.UtilityApplicationUpdateRequest true "Application data"
// @Success 204 {object} map[string]string
// @Failure 400 {object} ds.ErrorResponse
// @Failure 500 {object} ds.ErrorResponse
// @Router /api/applications/{id} [put]
func (h *Handler) UpdateApplication(gCtx *gin.Context) {
	id, err := strconv.Atoi(gCtx.Param("id"))
	if err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	var req ds.UtilityApplicationUpdateRequest
	if err := gCtx.BindJSON(&req); err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.UpdateApplicationUserFields(uint(id), req); err != nil {
		h.errorHandler(gCtx, http.StatusInternalServerError, err)
		return
	}

	gCtx.JSON(http.StatusNoContent, gin.H{
		"message": "Данные заявки обновлены",
	})
}

// FormApplication godoc
// @Summary Form application
// @Description Form draft application to make it ready for processing
// @Tags applications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Application ID"
// @Success 204 {object} map[string]string
// @Failure 400 {object} ds.ErrorResponse
// @Failure 500 {object} ds.ErrorResponse
// @Router /api/applications/{id}/form [put]
func (h *Handler) FormApplication(gCtx *gin.Context) {
	id, err := strconv.Atoi(gCtx.Param("id"))
	if err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	userID := h.GetCurrentUserID(gCtx) // ИЗМЕНЕНО
	if err := h.Repository.FormApplication(uint(id), userID); err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	gCtx.JSON(http.StatusNoContent, gin.H{
		"message": "Заявка сформирована",
	})
}

// ResolveApplication godoc
// @Summary Resolve application
// @Description Resolve application by moderator (Admin/Manager only)
// @Tags applications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Application ID"
// @Param request body ds.ApplicationStatusUpdateRequest true "Resolution data"
// @Success 204 {object} map[string]string
// @Failure 400 {object} ds.ErrorResponse
// @Failure 500 {object} ds.ErrorResponse
// @Router /api/applications/{id}/resolve [put]
func (h *Handler) ResolveApplication(gCtx *gin.Context) {
	id, err := strconv.Atoi(gCtx.Param("id"))
	if err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	var req ds.ApplicationStatusUpdateRequest
	if err := gCtx.BindJSON(&req); err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	moderatorID := h.GetCurrentUserID(gCtx)
	userRole := h.GetCurrentUserRole(gCtx)

	// ДОБАВЬ ПРОВЕРКУ РОЛИ
	if userRole != role.Manager && userRole != role.Admin {
		h.errorHandler(gCtx, http.StatusForbidden, fmt.Errorf("only moderators and admins can resolve applications"))
		return
	}

	if err := h.Repository.ResolveApplication(uint(id), moderatorID, req.Status); err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	gCtx.JSON(http.StatusNoContent, gin.H{
		"message": "Заявка обработана модератором",
	})
}

// DeleteApplication godoc
// @Summary Delete application
// @Description Delete application (only by creator)
// @Tags applications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Application ID"
// @Success 204 {object} map[string]string
// @Failure 400 {object} ds.ErrorResponse
// @Failure 500 {object} ds.ErrorResponse
// @Router /api/applications/{id} [delete]
func (h *Handler) DeleteApplication(gCtx *gin.Context) {
	id, err := strconv.Atoi(gCtx.Param("id"))
	if err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	userID := h.GetCurrentUserID(gCtx) // ИЗМЕНЕНО

	app, err := h.Repository.GetUtilityApplicationByID(uint(id))
	if err != nil {
		h.errorHandler(gCtx, http.StatusNotFound, err)
		return
	}

	if app.UserID != userID { // ИЗМЕНЕНО
		h.errorHandler(gCtx, http.StatusForbidden, fmt.Errorf("only creator can delete application"))
		return
	}

	if err := h.Repository.LogicallyDeleteApplication(uint(id)); err != nil {
		h.errorHandler(gCtx, http.StatusInternalServerError, err)
		return
	}

	gCtx.JSON(http.StatusNoContent, gin.H{
		"message": "Заявка удалена",
	})
}

// RemoveServiceFromApplication godoc
// @Summary Remove service from application
// @Description Remove service from application (without M-M PK)
// @Tags applications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Application ID"
// @Param service_id path int true "Service ID"
// @Success 204 {object} ds.MessageResponse
// @Failure 400 {object} ds.ErrorResponse
// @Failure 500 {object} ds.ErrorResponse
// @Router /api/applications/{id}/services/{service_id} [delete]
func (h *Handler) RemoveServiceFromApplication(gCtx *gin.Context) {
	applicationID, err := strconv.Atoi(gCtx.Param("id"))
	if err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	serviceID, err := strconv.Atoi(gCtx.Param("service_id"))
	if err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.RemoveServiceFromApplication(uint(applicationID), uint32(serviceID)); err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, err)
		return
	}

	gCtx.JSON(http.StatusNoContent, gin.H{
		"message": "Услуга удалена из заявки",
	})
}

// UpdateApplicationService godoc
// @Summary Update application service
// @Description Update service quantity/tariff in application (without M-M PK)
// @Tags applications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Application ID"
// @Param service_id path int true "Service ID"
// @Param request body ds.UtilityApplicationServiceUpdateRequest true "Service update data"
// @Success 200 {object} ds.MessageResponse
// @Failure 400 {object} ds.ErrorResponse
// @Failure 500 {object} ds.ErrorResponse
// @Router /api/applications/{id}/services/{service_id} [put]
func (h *Handler) UpdateApplicationService(gCtx *gin.Context) {
	applicationID, err := strconv.Atoi(gCtx.Param("id"))
	if err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, fmt.Errorf("invalid application ID"))
		return
	}

	serviceID, err := strconv.Atoi(gCtx.Param("service_id"))
	if err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, fmt.Errorf("invalid service ID"))
		return
	}

	var req ds.UtilityApplicationServiceUpdateRequest
	if err := gCtx.BindJSON(&req); err != nil {
		h.errorHandler(gCtx, http.StatusBadRequest, fmt.Errorf("invalid request data: %v", err))
		return
	}

	if req.Quantity == nil && req.Tariff == nil {
		h.errorHandler(gCtx, http.StatusBadRequest, fmt.Errorf("no fields to update"))
		return
	}

	if err := h.Repository.UpdateApplicationService(uint(applicationID), uint32(serviceID), req); err != nil {
		log.Printf("Error updating application service: %v", err)
		h.errorHandler(gCtx, http.StatusBadRequest, fmt.Errorf("failed to update service in application: %v", err))
		return
	}

	response := gin.H{
		"message":        "Услуга в заявке успешно обновлена",
		"application_id": applicationID,
		"service_id":     serviceID,
	}

	if req.Quantity != nil {
		response["quantity"] = *req.Quantity
	}
	if req.Tariff != nil {
		response["tariff"] = *req.Tariff
	}

	gCtx.JSON(http.StatusOK, response)
}
