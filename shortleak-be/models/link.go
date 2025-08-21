package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Link struct {
	gorm.Model
	ID         uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID     uuid.UUID
	URL        string `json:"url" gorm:"unique;not null"`
	ShortToken string `json:"short_token" gorm:"unique;not null"`
	Active     bool   `json:"active" gorm:"default:true"`
	User       User   `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (u *Link) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}
