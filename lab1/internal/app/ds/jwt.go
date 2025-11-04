package ds

import (
	"lab1/internal/app/role"

	"github.com/golang-jwt/jwt"
)

type JWTClaims struct {
	jwt.StandardClaims
	UserID uint      `json:"user_id"`
	Scopes []string  `json:"scopes"`
	Role   role.Role `json:"role"`
}
