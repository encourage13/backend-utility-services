package ds

// User соответствует таблице "users".
type User struct {
	ID          uint   `gorm:"primaryKey;column:id"`
	Login       string `gorm:"unique;column:login;size:25;not null"`
	Password    string `gorm:"column:password;size:100;not null"`
	IsModerator bool   `gorm:"column:is_moderator;not null;default:false"`

	// --- СВЯЗИ ---
	// "один-ко-многим": один пользователь может создать много заявок.
	RequestsCreated []Request `gorm:"foreignKey:CreatorID;references:ID"`
	// "один-ко-многим": один модератор может завершать много заявок.
	RequestsModerated []Request `gorm:"foreignKey:ModeratorID;references:ID"`
}

func (User) TableName() string { return "users" }
