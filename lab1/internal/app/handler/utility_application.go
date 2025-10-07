package handler

import (
	"fmt"
	"lab1/internal/app/ds"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GET "/utilities_application/:id"
// Показывает заявку и её позиции (услуги).
func (h *Handler) GetUtilitiesApplication(ctx *gin.Context) {
	idStr := strings.TrimSpace(ctx.Param("id"))
	var (
		app ds.UtilityApplication
		err error
	)

	if idStr == "" {
		app, err = h.Repository.CreateOrGetUtilityApplication(1)
	} else {
		id, convErr := strconv.Atoi(idStr)
		if convErr != nil {
			h.errorHandler(ctx, http.StatusBadRequest, convErr)
			return
		}
		app, err = h.Repository.GetUtilityApplicationByID(uint(id))
	}

	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("заявка не найдена"))
		return
	}

	ctx.HTML(http.StatusOK, "utilities_application.html", gin.H{
		"data": UtilitiesPageData{
			CartID:        app.ID,
			CartItems:     app.Services,
			CartItemCount: len(app.Services),
			Total:         app.TotalCost,
		},
	})
}

// POST "/utilities_application/add"
// Добавить услугу в текущую (черновую) заявку пользователя
// Ожидает form-data: service_id, quantity (необязательно). Редиректит на список услуг с сохранённым поиском.
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

	app, err := h.Repository.CreateOrGetUtilityApplication(1)
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

	logrus.Warn("delete application: ", idStr)

	if err := h.Repository.DeleteUtilityApplication(uint(appID)); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.Redirect(http.StatusMovedPermanently, "/utilities")
}
