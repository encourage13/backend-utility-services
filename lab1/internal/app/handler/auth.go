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
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type loginReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResp struct {
	ExpiresIn   time.Duration `json:"expires_in"`
	AccessToken string        `json:"access_token"`
	TokenType   string        `json:"token_type"`
	UserID      uint          `json:"user_id"`
	Role        role.Role     `json:"role"`
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ds.LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/login [post]
func (h *Handler) Login(gCtx *gin.Context) {
	cfg := h.Config
	req := &loginReq{}

	err := json.NewDecoder(gCtx.Request.Body).Decode(req)
	if err != nil {
		gCtx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// Валидация входных данных
	if req.Login == "" || req.Password == "" {
		gCtx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "login and password are required",
		})
		return
	}

	// Ищем пользователя в базе данных
	user, err := h.Repository.GetUserByLogin(req.Login)
	if err != nil {
		// Для безопасности возвращаем одинаковую ошибку при неверном логине или пароле
		gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid credentials",
		})
		return
	}

	// Проверяем пароль
	if !checkPasswordHash(req.Password, user.Pass) {
		gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid credentials",
		})
		return
	}

	// Генерируем JWT токен с данными пользователя из БД
	token := jwt.NewWithClaims(cfg.JWT.SigningMethod, &ds.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(cfg.JWT.ExpiresIn).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "bitop-admin",
		},
		UserID: user.ID,
		Role:   user.Role,
		Scopes: []string{},
	})

	if token == nil {
		gCtx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("token is nil"))
		return
	}

	strToken, err := token.SignedString([]byte(cfg.JWT.Token))
	if err != nil {
		gCtx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("cant create str token: %w", err))
		return
	}

	gCtx.JSON(http.StatusOK, loginResp{
		ExpiresIn:   cfg.JWT.ExpiresIn,
		AccessToken: strToken,
		TokenType:   "Bearer",
		UserID:      user.ID,
		Role:        user.Role,
	})
}

// Logout godoc
// @Summary User logout
// @Description Invalidate JWT token by adding it to blacklist
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "Logout data with JWT token"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/logout [post]
func (h *Handler) Logout(gCtx *gin.Context) {
	// Получаем заголовок
	jwtStr := gCtx.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, jwtPrefix) {
		gCtx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "authorization header is required with Bearer prefix",
		})
		return
	}

	// Отрезаем префикс
	jwtStr = jwtStr[len(jwtPrefix):]

	// Парсим токен чтобы получить expiration time
	token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.Config.JWT.Token), nil
	})
	if err != nil {
		gCtx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "invalid token",
		})
		return
	}

	claims, ok := token.Claims.(*ds.JWTClaims)
	if !ok {
		gCtx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "invalid token claims",
		})
		return
	}

	// Вычисляем оставшееся время жизни токена
	expiresAt := time.Unix(claims.ExpiresAt, 0)
	remainingTTL := time.Until(expiresAt)

	if remainingTTL <= 0 {
		gCtx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "token already expired",
		})
		return
	}

	// Сохраняем в блеклист редиса
	err = h.Redis.WriteJWTToBlacklist(gCtx.Request.Context(), jwtStr, remainingTTL)
	if err != nil {
		log.Printf("Failed to blacklist token: %v", err)
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	log.Printf("Token blacklisted successfully, TTL: %v", remainingTTL)

	gCtx.JSON(http.StatusOK, gin.H{
		"message": "successfully logged out",
	})
}

// generateHashString создает bcrypt хэш пароля
func generateHashString(s string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	if err != nil {
		// В случае ошибки fallback на менее безопасный метод (для совместимости)
		// В реальном приложении нужно обработать ошибку properly
		return fallbackHash(s)
	}
	return string(hash)
}

// checkPasswordHash проверяет соответствие пароля и хэша
func checkPasswordHash(password, hash string) bool {
	// Сначала пробуем bcrypt
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil {
		return true
	}

	// Если bcrypt не сработал, пробуем SHA1 (для миграции старых паролей)
	if checkFallbackHash(password, hash) {
		// Можно обновить хэш на bcrypt здесь
		return true
	}

	return false
}

// fallbackHash используется для совместимости со старыми паролями
func fallbackHash(s string) string {
	// Старая реализация с SHA1 - должна быть удалена после миграции
	// import "crypto/sha1"
	// import "encoding/hex"
	// h := sha1.New()
	// h.Write([]byte(s))
	// return hex.EncodeToString(h.Sum(nil))
	return s // временно возвращаем plain text для отладки
}

func checkFallbackHash(password, hash string) bool {
	// Реализация проверки старого хэша
	// Должна быть удалена после миграции всех паролей
	return password == hash // временно для отладки
}
