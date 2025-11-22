package models

type User struct {
	ID        string `gorm:"type:char(26);primaryKey" json:"user_id"`
	Username  string `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Password  string `gorm:"size:128;not null" json:"-"`
	AvatarURL string `gorm:"size:256" json:"avatar_url"`
	Email     string `gorm:"size:128" json:"email"`
}
