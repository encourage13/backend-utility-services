package handler

import (
	"fmt"
	"lab1/internal/app/ds"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetUtilitiesApplication(ctx *gin.Context) {
	idStr := strings.TrimSpace(ctx.Param("id"))
	var (
		app ds.UtilityApplication
		err error
	)

	userID := h.GetCurrentUserID()

	if idStr == "" {

		app, err = h.Repository.CreateOrGetUtilityApplication(userID)
		if err != nil {

			ctx.Redirect(http.StatusFound, "/utilities")
			return
		}
	} else {
		id, convErr := strconv.Atoi(idStr)
		if convErr != nil {

			ctx.Redirect(http.StatusFound, "/utilities")
			return
		}

		app, err = h.Repository.GetUtilityApplicationByID(uint(id))
		if err != nil {

			ctx.Redirect(http.StatusFound, "/utilities")
			return
		}

		if app.Status == string(ds.DELETED) {
			ctx.Redirect(http.StatusFound, "/utilities")
			return
		}
		if app.TotalCost == 0 {
			ctx.Redirect(http.StatusFound, "/utilities")
			return
		}
		if len(app.Services) == 0 {
			ctx.Redirect(http.StatusFound, "/utilities")
			return
		}
	}

	ctx.HTML(http.StatusOK, "utilities_application.html", gin.H{
		"data": UtilitiesPageData{
			CartID:        app.ID,
			CartItems:     app.Services,
			CartItemCount: len(app.Services),
			Total:         app.TotalCost,
			Address:       app.Address,
		},
	})
}

func (h *Handler) AddServiceInUtilityApplication(ctx *gin.Context) {
	rawID := ctx.PostForm("service_id")
	serviceID, err := strconv.Atoi(rawID)
	if err != nil || serviceID <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	qty := float32(1)
	if rawQty := strings.TrimSpace(ctx.PostForm("quantity")); rawQty != "" {
		if f, convErr := strconv.ParseFloat(rawQty, 32); convErr == nil && f > 0 {
			qty = float32(f)
		}
	}

	userID := h.GetCurrentUserID()
	app, err := h.Repository.CreateOrGetUtilityApplication(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if err := h.Repository.AddServiceToApplication(app.ID, uint32(serviceID), qty); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusMovedPermanently, "/utilities?searchUtilities="+ctx.PostForm("searchUtilities"))
}

func (h *Handler) DeleteUtilityApplication(ctx *gin.Context) {
	idStr := ctx.PostForm("application_id")
	appID, err := strconv.Atoi(idStr)
	if err != nil || appID <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.DeleteUtilityApplication(uint(appID)); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.Redirect(http.StatusMovedPermanently, "/utilities")
}

func (h *Handler) GetCartBadge(ctx *gin.Context) {
	userID := h.GetCurrentUserID()
	draft, err := h.Repository.GetDraftApplication(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, ds.CartBadgeDTO{
			ApplicationID: nil,
			Count:         0,
		})
		return
	}

	fullApplication, err := h.Repository.GetApplicationWithServices(draft.ID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, ds.CartBadgeDTO{
		ApplicationID: &fullApplication.ID,
		Count:         len(fullApplication.Services),
	})
}

func (h *Handler) ListApplications(ctx *gin.Context) {
	status := ctx.Query("status")
	from := ctx.Query("from")
	to := ctx.Query("to")

	applications, err := h.Repository.GetApplicationsFiltered(status, from, to)
	if err != nil {

		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	var applicationDTOs []ds.UtilityApplicationDTO
	for _, app := range applications {
		dto := ds.UtilityApplicationDTO{
			ID:           app.ID,
			UserID:       app.UserID,
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

		if app.User.ID != 0 {
			dto.User = &ds.UserDTO{
				ID:          app.User.ID,
				Login:       app.User.Login,
				IsModerator: app.User.IsModerator,
			}

		}
		if app.Moderator.ID != 0 {
			dto.Moderator = &ds.UserDTO{
				ID:          app.Moderator.ID,
				Login:       app.Moderator.Login,
				IsModerator: app.Moderator.IsModerator,
			}

		}

		applicationDTOs = append(applicationDTOs, dto)
	}

	ctx.JSON(http.StatusOK, applicationDTOs)
}

func (h *Handler) GetApplication(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	application, err := h.Repository.GetApplicationWithServices(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
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
		UserID:       application.UserID,
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

	if application.User.ID != 0 {
		applicationDTO.User = &ds.UserDTO{
			ID:          application.User.ID,
			Login:       application.User.Login,
			IsModerator: application.User.IsModerator,
		}
	}
	if application.Moderator.ID != 0 {
		applicationDTO.Moderator = &ds.UserDTO{
			ID:          application.Moderator.ID,
			Login:       application.Moderator.Login,
			IsModerator: application.Moderator.IsModerator,
		}
	}

	ctx.JSON(http.StatusOK, applicationDTO)
}

func (h *Handler) UpdateApplication(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var req ds.UtilityApplicationUpdateRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.UpdateApplicationUserFields(uint(id), req); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusNoContent, gin.H{
		"message": "Данные заявки обновлены",
	})
}

func (h *Handler) FormApplication(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	userID := h.GetCurrentUserID()
	if err := h.Repository.FormApplication(uint(id), userID); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusNoContent, gin.H{
		"message": "Заявка сформирована",
	})
}

func (h *Handler) ResolveApplication(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var req ds.ApplicationStatusUpdateRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	moderatorID := h.GetCurrentUserID()
	if err := h.Repository.ResolveApplication(uint(id), moderatorID, req.Status); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusNoContent, gin.H{
		"message": "Заявка обработана модератором",
	})
}

func (h *Handler) DeleteApplication(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	userID := h.GetCurrentUserID()

	app, err := h.Repository.GetUtilityApplicationByID(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	if app.UserID != userID {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("only creator can delete application"))
		return
	}

	if err := h.Repository.LogicallyDeleteApplication(uint(id)); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusNoContent, gin.H{
		"message": "Заявка удалена",
	})
}

func (h *Handler) RemoveServiceFromApplication(ctx *gin.Context) {
	applicationID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	serviceID, err := strconv.Atoi(ctx.Param("service_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.RemoveServiceFromApplication(uint(applicationID), uint32(serviceID)); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusNoContent, gin.H{
		"message": "Услуга удалена из заявки",
	})
}

func (h *Handler) UpdateApplicationService(ctx *gin.Context) {
	applicationID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid application ID"))
		return
	}

	serviceID, err := strconv.Atoi(ctx.Param("service_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid service ID"))
		return
	}

	var req ds.UtilityApplicationServiceUpdateRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid request data: %v", err))
		return
	}

	if req.Quantity == nil && req.Tariff == nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("no fields to update"))
		return
	}

	if err := h.Repository.UpdateApplicationService(uint(applicationID), uint32(serviceID), req); err != nil {
		log.Printf("Error updating application service: %v", err)
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("failed to update service in application: %v", err))
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

	ctx.JSON(http.StatusOK, response)
}
