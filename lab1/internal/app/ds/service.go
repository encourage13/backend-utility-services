package ds

// Service соответствует таблице "services".
type Service struct {
	ID          uint    `gorm:"primaryKey;column:id"`
	Title       string  `gorm:"column:title;size:100;not null"`
	Description string  `gorm:"column:description;type:text;not null"`
	IsActive    bool    `gorm:"column:is_active;not null;default:true"`
	ImageURL    *string `gorm:"column:image_url;size:255"`
	Unit        string  `gorm:"column:unit;size:20;not null"`
	Tariff      float64 `gorm:"column:tariff;type:numeric(12,2);not null"`

	// --- СВЯЗИ ---
	// "один-ко-многим": услуга встречается во многих строках заявок (m2m-таблица).
	RequestLines []RequestService `gorm:"foreignKey:ServiceID;references:ID"`
	// Не основная связь (удобная навигация): многие-ко-многим через request_services.
	// Учти, что дополнительные поля m2m (quantity, total_line и т.д.) здесь не видны.
	Requests []Request `gorm:"many2many:request_services;joinForeignKey:ServiceID;joinReferences:RequestID"`
}

func (Service) TableName() string { return "services" }
