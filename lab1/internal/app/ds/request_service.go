package ds

// RequestService соответствует таблице "request_services" (m2m с доп. полями).
type RequestService struct {
	RequestID      uint    `gorm:"primaryKey;column:request_id"`
	ServiceID      uint    `gorm:"primaryKey;column:service_id"`
	Quantity       float64 `gorm:"column:quantity;type:numeric(12,3);not null;default:1"`
	IsPrimary      bool    `gorm:"column:is_primary;not null;default:false"`
	Position       int     `gorm:"column:position;not null;default:0"`
	TariffSnapshot float64 `gorm:"column:tariff_snapshot;type:numeric(12,2);not null"`
	TotalLine      float64 `gorm:"column:total_line;type:numeric(14,2);not null"`

	// --- СВЯЗИ ---
	// Каждая строка принадлежит конкретной заявке и конкретной услуге.
	Request Request `gorm:"foreignKey:RequestID;references:ID;constraint:OnDelete:RESTRICT,OnUpdate:RESTRICT"`
	Service Service `gorm:"foreignKey:ServiceID;references:ID;constraint:OnDelete:RESTRICT,OnUpdate:RESTRICT"`
}

func (RequestService) TableName() string { return "request_services" }
