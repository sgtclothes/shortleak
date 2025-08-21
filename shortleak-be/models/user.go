package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID       uuid.UUID      `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	FullName string         `json:"fullname" gorm:"column:fullname"`
	Email    string         `json:"email" gorm:"unique"`
	Password string         `json:"password"`
	Active   bool           `json:"active" gorm:"default:true"`
	Data     datatypes.JSON `json:"data" gorm:"type:json"`
	Link     []Link         `json:"links" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}
