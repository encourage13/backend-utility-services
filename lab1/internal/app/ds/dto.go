package ds

import (
	"time"
)

// User DTOs
type UserDTO struct {
	ID          uint   `json:"id"` // Изменено с uint на uuid.UUID
	Login       string `json:"login"`
	IsModerator bool   `json:"is_moderator"`
}

type UserRegisterRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserLoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserUpdateRequest struct {
	Login       *string `json:"login"`
	Password    *string `json:"password"`
	IsModerator *bool   `json:"is_moderator"`
}

// UtilityService DTOs
type UtilityServiceDTO struct {
	ID          uint32  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	ImageURL    *string `json:"image_url"`
	Unit        string  `json:"unit"`
	Tariff      float32 `json:"tariff"`
}

type UtilityServiceCreateRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description" binding:"required"`
	ImageURL    *string `json:"image_url"`
	Unit        string  `json:"unit" binding:"required"`
	Tariff      float32 `json:"tariff" binding:"required"`
}

type UtilityServiceUpdateRequest struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	ImageURL    *string  `json:"image_url"`
	Unit        *string  `json:"unit"`
	Tariff      *float32 `json:"tariff"`
}

// UtilityApplication DTOs
type UtilityApplicationDTO struct {
	ID           uint       `json:"id"`
	UserID       uint       `json:"user_id"`
	Status       string     `json:"status"`
	TotalCost    float32    `json:"total_cost"`
	Address      *string    `json:"address"`
	DateCreated  time.Time  `json:"date_created"`
	DateFormed   *time.Time `json:"date_formed"`
	DateAccepted *time.Time `json:"date_accepted"`
	ModeratorID  *uint      `json:"moderator_id"`

	User      *UserDTO                       `json:"user,omitempty"`
	Moderator *UserDTO                       `json:"moderator,omitempty"`
	Services  []UtilityApplicationServiceDTO `json:"services,omitempty"`
}

type UtilityApplicationCreateRequest struct {
	UserID   uint                                   `json:"user_id" binding:"required"`
	Services []UtilityApplicationServiceItemRequest `json:"services" binding:"required"`
}

type UtilityApplicationUpdateRequest struct {
	Status      *string  `json:"status"`
	TotalCost   *float32 `json:"total_cost"`
	Address     *string  `json:"address"`
	ModeratorID *uint    `json:"moderator_id"`
}

// UtilityApplicationService DTOs
type UtilityApplicationServiceDTO struct {
	UtilityApplicationID uint     `json:"utility_application_id"`
	UtilityServiceID     uint32   `json:"utility_service_id"`
	Quantity             float32  `json:"quantity"`
	Total                float32  `json:"total"`
	CustomTariff         *float32 `json:"custom_tariff,omitempty"`

	Service *UtilityServiceDTO `json:"service,omitempty"`
}

// ds/dto.go
type UtilityApplicationServiceUpdateRequest struct {
	Quantity *float32 `json:"quantity"` // расход
	Tariff   *float32 `json:"tariff"`   // переопределенный тариф
}

type UtilityApplicationServiceItemRequest struct {
	UtilityServiceID uint32  `json:"utility_service_id" binding:"required"`
	Quantity         float32 `json:"quantity" binding:"required"`
}

// Cart DTOs
type CartBadgeDTO struct {
	ApplicationID *uint `json:"application_id"`
	Count         int   `json:"count"`
}

// Auth DTOs
type LoginResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

// Pagination DTO
type PaginatedResponse struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
}

// Application Status Update
type ApplicationStatusUpdateRequest struct {
	Status string `json:"status" binding:"required"` // "FORMED" | "REJECTED" | "COMPLETED"
}

// Stats DTO
type ApplicationStatsDTO struct {
	TotalApplications int     `json:"total_applications"`
	TotalCost         float64 `json:"total_cost"`
	DraftCount        int     `json:"draft_count"`
	FormedCount       int     `json:"formed_count"`
	CompletedCount    int     `json:"completed_count"`
}
type ErrorResponse struct {
	Error string `json:"error"`
}

// MessageResponse represents success message
type MessageResponse struct {
	Message string `json:"message"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// RegisterRequest represents registration data
type RegisterRequest struct {
	Name string `json:"name"`
	Pass string `json:"pass"`
}

// RegisterResponse represents registration response
type RegisterResponse struct {
	Ok bool `json:"ok"`
}
type LogoutRequest struct {
	Token string `json:"token"`
}
