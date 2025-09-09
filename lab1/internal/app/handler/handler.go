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

// Главная + ПОИСК
func (h *Handler) Index(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))

	// базовый список
	services := h.Repo.ListServices()

	// фильтрация по q (в названии или описании), регистр не учитываем
	if q != "" {
		qLower := strings.ToLower(q)
		filtered := make([]repository.Service, 0, len(services))
		for _, s := range services {
			if strings.Contains(strings.ToLower(s.Title), qLower) ||
				strings.Contains(strings.ToLower(s.Description), qLower) {
				filtered = append(filtered, s)
			}
		}
		services = filtered
	}

	// счётчик «заявки» (пока заглушка — можно заменить на реальный)
	count := h.Repo.CartCount()

	c.HTML(http.StatusOK, "index.html", gin.H{
		"cards": services,
		"count": count,
		"q":     q,
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
		"title": s.Title,
		"desc":  s.Description,
		"img":   s.ImageURL,
	})
}

func (h *Handler) Cart(c *gin.Context) {
	c.HTML(http.StatusOK, "cart.html", gin.H{
		"services": h.Repo.ListServices(),
	})
}
