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

func NewHandler(repo *repository.Repository) *Handler {
	return &Handler{Repo: repo}
}

// --------- GET / — список услуг + поиск (ORM) ---------
func (h *Handler) Index(c *gin.Context) {
	// В твоём проекте подставь реальный userID из сессии/мидлвари
	userID := uint(1)

	search := strings.TrimSpace(c.Query("search_service"))
	services, err := h.Repo.ListServices(search)
	if err != nil {
		c.String(http.StatusInternalServerError, "db error")
		return
	}

	count, _ := h.Repo.CountDraftItems(userID)
	_, lines, _, _ := h.Repo.GetDraftCart(userID) // just to know if есть черновик
	hasDraft := len(lines) > 0

	c.HTML(http.StatusOK, "index.html", gin.H{
		"cards":          services,
		"count":          count,
		"search_service": search,
		"hasDraft":       hasDraft, // если шаблон не использует — не страшно
	})
}

// --------- GET /services/:id — карточка услуги (ORM) ---------
func (h *Handler) Service(c *gin.Context) {
	userID := uint(1)

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid id")
		return
	}

	s, err := h.Repo.GetServiceByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}

	count, _ := h.Repo.CountDraftItems(userID)
	c.HTML(http.StatusOK, "service.html", gin.H{
		"title":  s.Title,
		"desc":   s.Description,
		"img":    s.ImageURL,
		"count":  count,
		"tariff": s.Tariff,
		"unit":   s.Unit,
		"id":     s.ID,
	})
}

// --------- GET /cart — текущая заявка (ORM) ---------
func (h *Handler) Cart(c *gin.Context) {
	userID := uint(1)

	reqID, lines, total, err := h.Repo.GetDraftCart(userID)
	if err != nil {
		c.String(http.StatusInternalServerError, "db error")
		return
	}

	count, _ := h.Repo.CountDraftItems(userID)
	if reqID == 0 {
		c.HTML(http.StatusOK, "cart.html", gin.H{
			"services": []any{},
			"count":    0,
			"total":    0,
			"reqID":    0,
		})
		return
	}

	c.HTML(http.StatusOK, "cart.html", gin.H{
		"services": lines,
		"count":    count,
		"total":    total,
		"reqID":    reqID,
	})
}

// --------- POST /cart/add — добавление в заявку (ORM) ---------
func (h *Handler) AddToCart(c *gin.Context) {
	userID := uint(1)

	sid := c.PostForm("service_id")
	qty := c.PostForm("quantity")

	svcID64, _ := strconv.ParseUint(sid, 10, 64)
	if svcID64 == 0 {
		c.String(http.StatusBadRequest, "service_id required")
		return
	}
	var q float64 = 1
	if qty != "" {
		if v, err := strconv.ParseFloat(qty, 64); err == nil && v > 0 {
			q = v
		}
	}

	_, err := h.Repo.AddServiceToDraft(userID, uint(svcID64), q)
	if err != nil {
		c.String(http.StatusInternalServerError, "db error")
		return
	}
	c.Redirect(http.StatusSeeOther, "/cart")
}

// --------- POST /cart/delete — логическое удаление заявки (RAW SQL) ---------
func (h *Handler) DeleteCart(c *gin.Context) {
	userID := uint(1)

	reqIDStr := c.PostForm("request_id")
	reqID64, _ := strconv.ParseUint(reqIDStr, 10, 64)
	if reqID64 == 0 {
		c.String(http.StatusBadRequest, "request_id required")
		return
	}

	_, err := h.Repo.SoftDeleteDraft(c, userID, uint(reqID64))
	if err != nil {
		c.String(http.StatusInternalServerError, "db error")
		return
	}
	c.Redirect(http.StatusSeeOther, "/")
}

// GET /cart/:id — просмотр конкретной АКТИВНОЙ заявки пользователя
func (h *Handler) CartByID(c *gin.Context) {
	userID := uint(1) // твой текущий «псевдо-логин»

	idStr := c.Param("id")
	reqID64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || reqID64 == 0 {
		c.String(http.StatusBadRequest, "invalid request id")
		return
	}

	reqID, lines, total, err := h.Repo.GetCartByID(userID, uint(reqID64))
	if err != nil {
		c.String(http.StatusInternalServerError, "db error")
		return
	}
	if reqID == 0 { // не найдена/не твоя/удалена
		c.String(http.StatusNotFound, "request not found")
		return
	}

	count, _ := h.Repo.CountDraftItems(userID)
	c.HTML(http.StatusOK, "cart.html", gin.H{
		"services": lines,
		"count":    count,
		"total":    total,
		"reqID":    reqID,
	})
}

// RegisterHandler настраивает все маршруты приложения
func (h *Handler) RegisterHandler(r *gin.Engine) {
	r.GET("/", h.Index)               // список услуг
	r.GET("/services/:id", h.Service) // карточка услуги
	r.GET("/cart", h.Cart)
	r.GET("/cart/:id", h.CartByID)
	r.POST("/cart/add", h.AddToCart)     // добавить в заявку
	r.POST("/cart/delete", h.DeleteCart) // удалить заявку (SQL)
}

// RegisterStatic подключает шаблоны и статические файлы
func (h *Handler) RegisterStatic(r *gin.Engine) {
	r.LoadHTMLGlob("../../templates/*.html")
	r.Static("/static", "../../recources")
	r.Static("/kartinki", "../../kartinki")
}
