package handler

// @title BITOP
// @version 1.0
// @description Bmstu Open IT Platform
// @host localhost:8080
// @BasePath /
import (
	"lab1/internal/app/ds"
	"lab1/internal/app/role"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

const jwtPrefix = "Bearer "

func (h *Handler) WithAuthCheck(assignedRoles ...role.Role) func(ctx *gin.Context) {
	return func(gCtx *gin.Context) {
		jwtStr := gCtx.GetHeader("Authorization")

		// ЕСЛИ НЕТ ТОКЕНА - 401 UNAUTHORIZED
		if !strings.HasPrefix(jwtStr, jwtPrefix) {
			log.Println("No Authorization header or Bearer prefix")
			gCtx.AbortWithStatus(http.StatusUnauthorized) // ИЗМЕНИЛ НА 401
			return
		}

		// Отрезаем префикс
		jwtStr = jwtStr[len(jwtPrefix):]

		// ЕСЛИ ТОКЕН ПУСТОЙ - 401
		if jwtStr == "" {
			log.Println("Empty JWT token after Bearer prefix")
			gCtx.AbortWithStatus(http.StatusUnauthorized) // ИЗМЕНИЛ НА 401
			return
		}

		// ПРОВЕРКА В BLACKLIST
		err := h.Redis.CheckJWTInBlacklist(gCtx.Request.Context(), jwtStr)
		if err == nil {
			log.Println("Token found in blacklist")
			gCtx.AbortWithStatus(http.StatusForbidden) // 403 - токен в blacklist
			return
		}

		// ПАРСИНГ JWT
		token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.Config.JWT.Token), nil
		})
		if err != nil {
			log.Printf("JWT parsing error: %v", err)
			gCtx.AbortWithStatus(http.StatusUnauthorized) // 401 - невалидный токен
			return
		}

		myClaims := token.Claims.(*ds.JWTClaims)

		// ПРОВЕРКА РОЛЕЙ (ТОЛЬКО ЕСЛИ УКАЗАНЫ РОЛИ)
		if len(assignedRoles) > 0 {
			hasAccess := false
			for _, requiredRole := range assignedRoles {
				if myClaims.Role == requiredRole {
					hasAccess = true
					break
				}
			}

			if !hasAccess {
				log.Printf("Access denied: user role %v not in required roles %v", myClaims.Role, assignedRoles)
				gCtx.AbortWithStatus(http.StatusForbidden) // 403 - нет прав
				return
			}
		}

		// СОХРАНЯЕМ ДАННЫЕ ПОЛЬЗОВАТЕЛЯ
		gCtx.Set("userID", myClaims.UserID)
		gCtx.Set("userRole", myClaims.Role)

		gCtx.Next()
	}
}
