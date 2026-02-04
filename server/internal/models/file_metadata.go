package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FileMetadata 逻辑文件元数据模型
type FileMetadata struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index:idx_user_files" json:"user_id"`
	FileBlobHash string    `gorm:"type:varchar(64);not null;index" json:"file_blob_hash"`

	// 用户自定义字段
	Filename string `gorm:"type:varchar(255);not null" json:"filename"`
	Size     int64  `gorm:"type:bigint;not null" json:"size"`

	// 生命周期管理
	CreatedAt time.Time  `gorm:"not null;default:now();index" json:"created_at"`
	ExpiresAt *time.Time `gorm:"default:null;index" json:"expires_at,omitempty"`
	DeletedAt *time.Time `gorm:"default:null;index:idx_user_files" json:"deleted_at,omitempty"` // 软删除

	// 关联关系
	User       User        `gorm:"foreignKey:UserID" json:"-"`
	FileBlob   FileBlob    `gorm:"foreignKey:FileBlobHash;references:Hash" json:"-"`
	ShareFiles []ShareFile `gorm:"foreignKey:FileID" json:"-"`
}

// TableName 指定表名
func (FileMetadata) TableName() string {
	return "files_metadata"
}

// BeforeCreate GORM 钩子：创建前
func (fm *FileMetadata) BeforeCreate(tx *gorm.DB) error {
	if fm.ID == uuid.Nil {
		fm.ID = uuid.New()
	}
	return nil
}

// IsDeleted 检查是否已软删除
func (fm *FileMetadata) IsDeleted() bool {
	return fm.DeletedAt != nil
}

// IsExpired 检查是否已过期
func (fm *FileMetadata) IsExpired() bool {
	if fm.ExpiresAt == nil {
		return false
	}
	return fm.ExpiresAt.Before(time.Now())
}

// IsActive 检查是否为活跃文件（未删除且未过期）
func (fm *FileMetadata) IsActive() bool {
	return !fm.IsDeleted() && !fm.IsExpired()
}

// SoftDelete 软删除文件
func (fm *FileMetadata) SoftDelete(tx *gorm.DB) error {
	now := time.Now()
	fm.DeletedAt = &now
	return tx.Model(fm).Update("deleted_at", now).Error
}

// Restore 恢复软删除的文件
func (fm *FileMetadata) Restore(tx *gorm.DB) error {
	fm.DeletedAt = nil
	return tx.Model(fm).Update("deleted_at", nil).Error
}

// SetExpiry 设置过期时间
func (fm *FileMetadata) SetExpiry(tx *gorm.DB, expiresAt time.Time) error {
	fm.ExpiresAt = &expiresAt
	return tx.Model(fm).Update("expires_at", expiresAt).Error
}

// Rename 重命名文件
func (fm *FileMetadata) Rename(tx *gorm.DB, newFilename string) error {
	fm.Filename = newFilename
	return tx.Model(fm).Update("filename", newFilename).Error
}

// FormatSize 格式化文件大小
func (fm *FileMetadata) FormatSize() string {
	return formatBytes(fm.Size)
}

// GetExtension 获取文件扩展名
func (fm *FileMetadata) GetExtension() string {
	for i := len(fm.Filename) - 1; i >= 0; i-- {
		if fm.Filename[i] == '.' {
			return fm.Filename[i:]
		}
	}
	return ""
}

// DaysUntilExpiry 获取距离过期的天数
func (fm *FileMetadata) DaysUntilExpiry() int {
	if fm.ExpiresAt == nil {
		return -1 // 永不过期
	}
	duration := time.Until(*fm.ExpiresAt)
	return int(duration.Hours() / 24)
}

// DaysSinceDeleted 获取删除后的天数
func (fm *FileMetadata) DaysSinceDeleted() int {
	if fm.DeletedAt == nil {
		return 0
	}
	duration := time.Since(*fm.DeletedAt)
	return int(duration.Hours() / 24)
}
