package models

type Channel struct {
	ID        string `gorm:"type:char(26);primaryKey" json:"channel_id"`
	Name      string `gorm:"uniqueIndex;size:64;not null" json:"channel_name"`
	CreatedBy string `gorm:"type:char(26);index;not null" json:"created_by"`
	Visible   bool   `gorm:"default:true" json:"-"`
}
