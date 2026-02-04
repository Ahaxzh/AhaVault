package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UploadStatus 上传状态类型
type UploadStatus string

const (
	UploadStatusUploading UploadStatus = "uploading"
	UploadStatusCompleted UploadStatus = "completed"
	UploadStatusFailed    UploadStatus = "failed"
)

// UploadSession Tus 协议上传会话模型
type UploadSession struct {
	ID     uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID uuid.UUID    `gorm:"type:uuid;not null;index" json:"user_id"`
	Status UploadStatus `gorm:"type:varchar(50);not null;default:'uploading';index" json:"status"`

	// 上传进度
	UploadOffset int64 `gorm:"type:bigint;not null;default:0" json:"upload_offset"` // 已上传字节数
	UploadLength int64 `gorm:"type:bigint;not null" json:"upload_length"`           // 文件总大小

	// 文件信息
	Filename string `gorm:"type:varchar(255);not null" json:"filename"`
	MimeType string `gorm:"type:varchar(128)" json:"mime_type"`
	Hash     string `gorm:"type:varchar(64)" json:"hash"` // 前端计算的哈希

	// 临时存储路径
	TempPath string `gorm:"type:varchar(255)" json:"temp_path"`

	// 时间戳
	CreatedAt   time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"not null;default:now();index" json:"updated_at"`
	CompletedAt *time.Time `gorm:"default:null" json:"completed_at,omitempty"`

	// 关联关系
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName 指定表名
func (UploadSession) TableName() string {
	return "upload_sessions"
}

// BeforeCreate GORM 钩子：创建前
func (us *UploadSession) BeforeCreate(tx *gorm.DB) error {
	if us.ID == uuid.Nil {
		us.ID = uuid.New()
	}
	return nil
}

// IsCompleted 检查是否已完成
func (us *UploadSession) IsCompleted() bool {
	return us.Status == UploadStatusCompleted
}

// IsFailed 检查是否失败
func (us *UploadSession) IsFailed() bool {
	return us.Status == UploadStatusFailed
}

// IsUploading 检查是否正在上传
func (us *UploadSession) IsUploading() bool {
	return us.Status == UploadStatusUploading
}

// Progress 获取上传进度百分比
func (us *UploadSession) Progress() float64 {
	if us.UploadLength == 0 {
		return 0
	}
	progress := float64(us.UploadOffset) / float64(us.UploadLength) * 100
	if progress > 100 {
		return 100
	}
	return progress
}

// RemainingBytes 获取剩余待上传字节数
func (us *UploadSession) RemainingBytes() int64 {
	remaining := us.UploadLength - us.UploadOffset
	if remaining < 0 {
		return 0
	}
	return remaining
}

// UpdateOffset 更新上传偏移量
func (us *UploadSession) UpdateOffset(tx *gorm.DB, offset int64) error {
	us.UploadOffset = offset
	return tx.Model(us).Update("upload_offset", offset).Error
}

// MarkCompleted 标记为已完成
func (us *UploadSession) MarkCompleted(tx *gorm.DB) error {
	now := time.Now()
	us.Status = UploadStatusCompleted
	us.CompletedAt = &now
	return tx.Model(us).Updates(map[string]interface{}{
		"status":       UploadStatusCompleted,
		"completed_at": now,
	}).Error
}

// MarkFailed 标记为失败
func (us *UploadSession) MarkFailed(tx *gorm.DB) error {
	us.Status = UploadStatusFailed
	return tx.Model(us).Update("status", UploadStatusFailed).Error
}

// IsStale 检查是否为陈旧会话（超过指定时间未更新）
func (us *UploadSession) IsStale(duration time.Duration) bool {
	return time.Since(us.UpdatedAt) > duration
}

// FormatUploadedSize 格式化已上传大小
func (us *UploadSession) FormatUploadedSize() string {
	return formatBytes(us.UploadOffset)
}

// FormatTotalSize 格式化总大小
func (us *UploadSession) FormatTotalSize() string {
	return formatBytes(us.UploadLength)
}

// UploadSpeed 计算上传速度（字节/秒）
func (us *UploadSession) UploadSpeed() float64 {
	duration := time.Since(us.CreatedAt).Seconds()
	if duration == 0 {
		return 0
	}
	return float64(us.UploadOffset) / duration
}

// EstimatedTimeRemaining 估算剩余上传时间
func (us *UploadSession) EstimatedTimeRemaining() time.Duration {
	speed := us.UploadSpeed()
	if speed == 0 {
		return 0
	}
	remaining := us.RemainingBytes()
	seconds := float64(remaining) / speed
	return time.Duration(seconds) * time.Second
}
