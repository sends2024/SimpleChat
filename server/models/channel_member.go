package models

type ChannelMember struct {
	ChannelID string `gorm:"type:char(26);primaryKey" json:"channel_id"`
	UserID    string `gorm:"type:char(26);primaryKey" json:"user_id"`
}
