package models

import (
	"time"

	"github.com/google/uuid"
)

// ShareFile 分享文件关联模型（多对多关系）
type ShareFile struct {
	ShareID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"share_id"`
	FileID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"file_id"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`

	// 关联关系
	Share ShareSession `gorm:"foreignKey:ShareID" json:"-"`
	File  FileMetadata `gorm:"foreignKey:FileID" json:"-"`
}

// TableName 指定表名
func (ShareFile) TableName() string {
	return "share_files"
}
