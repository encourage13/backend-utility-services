package ds

import (
	"lab1/internal/app/role"
)

type User struct {
	ID   uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string    `gorm:"uniqueIndex" json:"name"` // ДОБАВЬТЕ ЭТО
	Role role.Role `sql:"type:string;" json:"role"`
	Pass string    `json:"-"`
}
