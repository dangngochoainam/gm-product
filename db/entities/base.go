package entities

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	CreatedAt        time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;autoCreateTime"`
	CreatedUser      *string        `gorm:"column:created_user"`
	LastModifiedAt   time.Time      `gorm:"column:last_modified_at;not null;default:CURRENT_TIMESTAMP;autoUpdateTime"`
	LastModifiedUser *string        `gorm:"column:last_modified_user"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;index"`
}
