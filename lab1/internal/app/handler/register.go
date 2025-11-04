package handler

// @title BITOP
// @version 1.0
// @description Bmstu Open IT Platform
// @host localhost:8080
// @BasePath /
import (
	"encoding/json"
	"fmt"
	"lab1/internal/app/ds"
	"lab1/internal/app/role"
	"net/http"

	"github.com/gin-gonic/gin"
)

type registerReq struct {
	Name string `json:"name"`
	Pass string `json:"pass"`
}

type registerResp struct {
	Ok bool `json:"ok"`
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "User registration data"
// @Success 200 {object} RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/sign_up [post]

func (h *Handler) Register(gCtx *gin.Context) {
	req := &registerReq{}

	err := json.NewDecoder(gCtx.Request.Body).Decode(req)
	if err != nil {
		gCtx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if req.Pass == "" {
		gCtx.AbortWithError(http.StatusBadRequest, fmt.Errorf("pass is empty"))
		return
	}

	if req.Name == "" {
		gCtx.AbortWithError(http.StatusBadRequest, fmt.Errorf("name is empty"))
		return
	}

	err = h.Repository.Register(&ds.User{
		Role: role.Buyer,
		Name: req.Name,
		Pass: generateHashString(req.Pass),
	})
	if err != nil {
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	gCtx.JSON(http.StatusOK, &registerResp{
		Ok: true,
	})
}
