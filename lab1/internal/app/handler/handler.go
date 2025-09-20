package handler

import (
	"net/http"
	"strconv"
	"strings"

	"lab1/internal/app/repository"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Repo *repository.Repository
}

func New(repo *repository.Repository) *Handler {
	return &Handler{Repo: repo}
}

// Главная + поиск
func (h *Handler) Index(c *gin.Context) {
	search_service := strings.TrimSpace(c.Query("search_service"))

	// базовый список
	services := h.Repo.ListServices()

	if search_service != "" {
		qLower := strings.ToLower(search_service)
		filtered := make([]repository.Service, 0, len(services))
		for _, s := range services {
			if strings.Contains(strings.ToLower(s.Title), qLower) ||
				strings.Contains(strings.ToLower(s.Description), qLower) {
				filtered = append(filtered, s)
			}
		}
		services = filtered
	}

	// счётчик «заявки» из корзины
	count := h.Repo.CartCount()

	c.HTML(http.StatusOK, "index.html", gin.H{
		"cards":          services,
		"count":          count,
		"search_service": search_service,
	})
}

func (h *Handler) Service(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid id")
		return
	}

	s, ok := h.Repo.GetServiceByID(id)
	if !ok {
		c.String(http.StatusNotFound, "not found")
		return
	}

	c.HTML(http.StatusOK, "service.html", gin.H{
		"title":  s.Title,
		"desc":   s.Description,
		"img":    s.ImageURL,
		"count":  h.Repo.CartCount(),
		"tariff": s.Tariff,
	})
}

func (h *Handler) Cart(c *gin.Context) {
	services := h.Repo.ListCartServices()

	var grandTotal float32
	for _, s := range services {
		grandTotal += s.Total
	}

	c.HTML(http.StatusOK, "cart.html", gin.H{
		"services": services,
		"count":    h.Repo.CartCount(),
		"total":    grandTotal,
	})
}
