package models

import "time"

type Message struct {
	ID        string    `gorm:"type:char(26);primaryKey" json:"message_id"`
	ChannelID string    `gorm:"type:char(26);index;not null" json:"channel_id"`
	SenderID  string    `gorm:"type:char(26);index;not null" json:"sender_id"`
	Content   string    `gorm:"size:1024;not null" json:"content"`
	SentAt    time.Time `gorm:"autoCreateTime" json:"sent_at"`
}
