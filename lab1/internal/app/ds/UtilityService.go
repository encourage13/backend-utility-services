package ds

type UtilityService struct {
	ID          uint32  `gorm:"primaryKey;autoIncrement"`
	Title       string  `gorm:"size:256;not null"`
	Description string  `gorm:"type:text"`
	ImageURL    string  `gorm:"size:512;default:null"`
	Unit        string  `gorm:"size:64;not null"`
	Tariff      float32 `gorm:"not null"`
}
