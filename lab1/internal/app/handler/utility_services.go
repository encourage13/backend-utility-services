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
}

// GET "/" и "/utilities" — список услуг
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

	app, err := h.Repository.CreateOrGetUtilityApplication(1)
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

// GET "/utilities/:id" — карточка одной услуги
func (h *Handler) GetUtility(ctx *gin.Context) {
	idStr := ctx.Param("id")
	raw, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || raw == 0 {
		ctx.String(http.StatusBadRequest, "invalid id")
		return
	}

	svc, err := h.Repository.GetUtilityServiceByID(uint32(raw))
	if err != nil {
		ctx.String(http.StatusNotFound, "not found")
		return
	}

	app, err := h.Repository.CreateOrGetUtilityApplication(1)
	if err == nil {
		app, err = h.Repository.GetUtilityApplicationByID(app.ID)
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
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
