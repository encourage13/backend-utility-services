package ds

type UtilityApplicationService struct {
	UtilityApplicationID uint    `gorm:"primaryKey;autoIncrement:false"`
	UtilityServiceID     uint32  `gorm:"primaryKey;autoIncrement:false"`
	Quantity             float32 `gorm:"default:1;not null"`
	Total                float32 `gorm:"default:0;not null"`

	UtilityApplication UtilityApplication `gorm:"foreignKey:UtilityApplicationID"`
	Service            UtilityService     `gorm:"foreignKey:UtilityServiceID"`
}
