package ds

import "time"

// Request соответствует таблице "requests".
type Request struct {
	ID            uint       `gorm:"primaryKey;column:id"`
	Status        int        `gorm:"column:status;type:smallint;not null"` // 1..5 (см. consts)
	DateCreated   time.Time  `gorm:"column:date_created;not null;default:now()"`
	CreatorID     uint       `gorm:"column:creator_id;not null"`
	DateFormed    *time.Time `gorm:"column:date_formed"`
	DateCompleted *time.Time `gorm:"column:date_completed"`
	ModeratorID   *uint      `gorm:"column:moderator_id"`

	// --- СВЯЗИ ---
	// Создатель заявки (обязательно)
	Creator User `gorm:"foreignKey:CreatorID;references:ID;constraint:OnDelete:RESTRICT,OnUpdate:RESTRICT"`
	// Модератор заявки (может быть NULL)
	Moderator *User `gorm:"foreignKey:ModeratorID;references:ID;constraint:OnDelete:RESTRICT,OnUpdate:RESTRICT"`

	// Строки заявки (m2m со своими полями)
	Lines []RequestService `gorm:"foreignKey:RequestID;references:ID"`
	// Не основная связь: список услуг через промежуточную таблицу (без доступа к дополнительным полям).
	Services []Service `gorm:"many2many:request_services;joinForeignKey:RequestID;joinReferences:ServiceID"`
}

func (Request) TableName() string { return "requests" }
