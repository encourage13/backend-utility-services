package handler

// @title BITOP
// @version 1.0
// @description Bmstu Open IT Platform
// @host localhost:8080
// @BasePath /
import (
	"lab1/internal/app/ds"
	"lab1/internal/app/role"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetUserData godoc
// @Summary Get user data by ID
// @Description Get user information by user ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} ds.UserDTO
// @Failure 400 {object} ds.ErrorResponse
// @Failure 404 {object} ds.ErrorResponse
// @Router /api/users/{id} [get]
func (h *Handler) GetUserData(ctx *gin.Context) {
	idStr := ctx.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.GetUserByID(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	userDTO := ds.UserDTO{
		Login:       user.Name,
		IsModerator: (user.Role == role.Manager || user.Role == role.Admin),
	}
	ctx.JSON(http.StatusOK, userDTO)
}

// UpdateUserData godoc
// @Summary Update user data
// @Description Update user information
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body ds.UserUpdateRequest true "User update data"
// @Success 204 {object} map[string]string
// @Failure 400 {object} ds.ErrorResponse
// @Failure 500 {object} ds.ErrorResponse
// @Router /api/users/{id} [put]
func (h *Handler) UpdateUserData(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var req ds.UserUpdateRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.UpdateUser(uint(id), req); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusNoContent, gin.H{
		"message": "Данные пользователя обновлены",
	})
}
