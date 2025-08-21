package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Log struct {
	gorm.Model
	ID     uuid.UUID      `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	Action string         `json:"action" gorm:"not null"`
	Data   datatypes.JSON `json:"data" gorm:"type:json"`
}

func (u *Log) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}
