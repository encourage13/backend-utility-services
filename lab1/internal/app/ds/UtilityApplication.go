package ds

import "time"

type UtilityApplication struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	UserID       uint      `gorm:"not null;index"`
	Status       string    `gorm:"size:64;default:DRAFT;not null"`
	TotalCost    float32   `gorm:"not null;default:0"`
	Address      string    `gorm:"size:512;default:null"`
	DateCreated  time.Time `gorm:"autoCreateTime;not null"`
	DateFormed   time.Time `gorm:"default:null"`
	DateAccepted time.Time `gorm:"default:null"`
	ModeratorID  *uint     `gorm:"index"`

	User      User `gorm:"foreignKey:UserID"`
	Moderator User `gorm:"foreignKey:ModeratorID"`

	Services []UtilityApplicationService `gorm:"foreignKey:UtilityApplicationID"`
}

type ApplicationStatus string

const (
	DRAFT     ApplicationStatus = "DRAFT"
	DELETED   ApplicationStatus = "DELETED"
	FORMED    ApplicationStatus = "FORMED"
	REJECTED  ApplicationStatus = "REJECTED"
	COMPLETED ApplicationStatus = "COMPLETED"
)
