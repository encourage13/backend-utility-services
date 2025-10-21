package handler

import (
	"lab1/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Register(ctx *gin.Context) {
	var req ds.UserRegisterRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	user := ds.User{
		Login:          req.Login,
		HashedPassword: string(hashedPassword),
		IsModerator:    false,
	}

	if err := h.Repository.CreateUser(&user); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	userDTO := ds.UserDTO{
		ID:          user.ID,
		Login:       user.Login,
		IsModerator: user.IsModerator,
	}

	ctx.JSON(http.StatusCreated, userDTO)
}

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
		ID:          user.ID,
		Login:       user.Login,
		IsModerator: user.IsModerator,
	}
	ctx.JSON(http.StatusOK, userDTO)
}

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

func (h *Handler) Login(ctx *gin.Context) {
	var req ds.UserLoginRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.GetUserByUsername(req.Login)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password)); err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	response := ds.LoginResponse{
		Token: "the_really_good_token",
		User: ds.UserDTO{
			ID:          user.ID,
			Login:       user.Login,
			IsModerator: user.IsModerator,
		},
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *Handler) Logout(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Деавторизация прошла успешно",
	})
}
