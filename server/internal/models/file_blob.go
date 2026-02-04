package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// FileBlob 物理文件模型（CAS 内容寻址存储）
type FileBlob struct {
	Hash         string `gorm:"type:varchar(64);primary_key" json:"hash"`
	StorePath    string `gorm:"type:varchar(255);not null" json:"store_path"`
	EncryptedDEK string `gorm:"type:text;not null" json:"-"` // 加密的 DEK，不返回到前端
	Size         int64  `gorm:"type:bigint;not null" json:"size"`
	MimeType     string `gorm:"type:varchar(128)" json:"mime_type"`

	// 引用计数（CAS 核心字段）
	RefCount int `gorm:"type:int;not null;default:1;index" json:"ref_count"`

	// 管理字段
	IsBanned  bool   `gorm:"type:boolean;not null;default:false;index" json:"is_banned"`
	BanReason string `gorm:"type:text" json:"ban_reason,omitempty"`

	// 时间戳
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`

	// 关联关系
	Files []FileMetadata `gorm:"foreignKey:FileBlobHash;references:Hash" json:"-"`
}

// TableName 指定表名
func (FileBlob) TableName() string {
	return "file_blobs"
}

// IncrementRefCount 原子地增加引用计数
func (fb *FileBlob) IncrementRefCount(tx *gorm.DB) error {
	return tx.Model(fb).Update("ref_count", gorm.Expr("ref_count + ?", 1)).Error
}

// DecrementRefCount 原子地减少引用计数
func (fb *FileBlob) DecrementRefCount(tx *gorm.DB) error {
	return tx.Model(fb).Update("ref_count", gorm.Expr("ref_count - ?", 1)).Error
}

// IsOrphan 检查是否为孤儿文件（引用计数为 0）
func (fb *FileBlob) IsOrphan() bool {
	return fb.RefCount <= 0
}

// CanShare 检查是否可以分享（未被禁止）
func (fb *FileBlob) CanShare() bool {
	return !fb.IsBanned
}

// Ban 禁止文件
func (fb *FileBlob) Ban(tx *gorm.DB, reason string) error {
	fb.IsBanned = true
	fb.BanReason = reason
	return tx.Model(fb).Updates(map[string]interface{}{
		"is_banned":  true,
		"ban_reason": reason,
	}).Error
}

// Unban 解除禁止
func (fb *FileBlob) Unban(tx *gorm.DB) error {
	fb.IsBanned = false
	fb.BanReason = ""
	return tx.Model(fb).Updates(map[string]interface{}{
		"is_banned":  false,
		"ban_reason": "",
	}).Error
}

// FormatSize 格式化文件大小（返回人类可读格式）
func (fb *FileBlob) FormatSize() string {
	return formatBytes(fb.Size)
}

// formatBytes 格式化字节数
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
