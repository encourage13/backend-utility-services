package handler

import (
	"net/http"
	"strconv"
	"strings"

	"lab1/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repo *repository.Repository
}

func New(r *repository.Repository) *Handler {
	return &Handler{Repo: r}
}

type PageData struct {
	Services      []repository.UtilityService
	SearchQuery   string
	Service       repository.UtilityService
	CartID        int
	CartItems     []repository.SingleUtilityApplication
	CartItemCount int
	Total         float32
}

func (h *Handler) GetUtilities(c *gin.Context) {
	q := strings.TrimSpace(c.Query("searchUtilities"))

	var services []repository.UtilityService
	var err error
	if q == "" {
		services, err = h.Repo.GetUtilityServices()
	} else {
		services, err = h.Repo.SearchUtilityServices(q)
	}
	if err != nil {
		logrus.Error(err)
	}

	cart, _ := h.Repo.GetUtility__Application(1)

	c.HTML(http.StatusOK, "list_of_utilities.html", gin.H{
		"data": PageData{
			Services:      services,
			SearchQuery:   q,
			CartID:        1,
			CartItems:     cart.Services,
			CartItemCount: len(cart.Services),
			Total:         cart.TotalCost,
		},
	})
}

func (h *Handler) GetUtility(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid id")
		return
	}
	svc, err := h.Repo.GetUtilityServiceByID(id)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}

	cart, _ := h.Repo.GetUtility__Application(1)

	c.HTML(http.StatusOK, "utility.html", gin.H{
		"data": PageData{
			Service:       svc,
			CartID:        1,
			CartItems:     cart.Services,
			CartItemCount: len(cart.Services),
			Total:         cart.TotalCost,
		},
	})
}

func (h *Handler) GetUtilitiesApplication(c *gin.Context) {
	cartIDStr := strings.TrimSpace(c.Param("id"))
	if cartIDStr == "" {
		cartIDStr = "1"
	}
	cartID, err := strconv.Atoi(cartIDStr)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid application id")
		return
	}

	cart, err := h.Repo.GetUtility__Application(cartID)
	if err != nil {
		c.String(http.StatusNotFound, "application not found")
		return
	}

	c.HTML(http.StatusOK, "utilities_application.html", gin.H{
		"data": PageData{
			CartID:        cartID,
			CartItems:     cart.Services,
			CartItemCount: len(cart.Services),
			Total:         cart.TotalCost,
		},
	})
}
