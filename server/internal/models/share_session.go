package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ShareSession 分享会话模型
type ShareSession struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PickupCode string    `gorm:"type:varchar(8);uniqueIndex;not null" json:"pickup_code"`
	CreatorID  uuid.UUID `gorm:"type:uuid;not null;index" json:"creator_id"`

	// 访问控制
	PasswordHash     string `gorm:"type:varchar(255)" json:"-"` // 密码哈希，不返回到前端
	MaxDownloads     int    `gorm:"type:int;not null;default:0" json:"max_downloads"`
	CurrentDownloads int    `gorm:"type:int;not null;default:0" json:"current_downloads"`

	// 生命周期
	CreatedAt time.Time  `gorm:"not null;default:now();index" json:"created_at"`
	ExpiresAt time.Time  `gorm:"not null;index" json:"expires_at"`
	StoppedAt *time.Time `gorm:"default:null" json:"stopped_at,omitempty"`

	// 关联关系
	Creator    User        `gorm:"foreignKey:CreatorID" json:"-"`
	ShareFiles []ShareFile `gorm:"foreignKey:ShareID" json:"-"`
}

// TableName 指定表名
func (ShareSession) TableName() string {
	return "share_sessions"
}

// BeforeCreate GORM 钩子：创建前
func (ss *ShareSession) BeforeCreate(tx *gorm.DB) error {
	if ss.ID == uuid.Nil {
		ss.ID = uuid.New()
	}
	return nil
}

// ShareStatus 分享状态类型
type ShareStatus string

const (
	ShareStatusActive    ShareStatus = "active"    // 活跃
	ShareStatusExpired   ShareStatus = "expired"   // 已过期
	ShareStatusExhausted ShareStatus = "exhausted" // 下载次数用尽
	ShareStatusStopped   ShareStatus = "stopped"   // 手动停止
)

// GetStatus 获取分享状态
func (ss *ShareSession) GetStatus() ShareStatus {
	// 优先检查手动停止
	if ss.StoppedAt != nil {
		return ShareStatusStopped
	}

	// 检查是否过期
	if ss.IsExpired() {
		return ShareStatusExpired
	}

	// 检查下载次数
	if ss.IsExhausted() {
		return ShareStatusExhausted
	}

	return ShareStatusActive
}

// IsExpired 检查是否已过期
func (ss *ShareSession) IsExpired() bool {
	return ss.ExpiresAt.Before(time.Now())
}

// IsExhausted 检查下载次数是否用尽
func (ss *ShareSession) IsExhausted() bool {
	if ss.MaxDownloads == 0 {
		return false // 0 表示不限次数
	}
	return ss.CurrentDownloads >= ss.MaxDownloads
}

// IsStopped 检查是否已手动停止
func (ss *ShareSession) IsStopped() bool {
	return ss.StoppedAt != nil
}

// IsActive 检查是否为活跃分享
func (ss *ShareSession) IsActive() bool {
	return ss.GetStatus() == ShareStatusActive
}

// HasPassword 检查是否设置了密码
func (ss *ShareSession) HasPassword() bool {
	return ss.PasswordHash != ""
}

// Stop 停止分享
func (ss *ShareSession) Stop(tx *gorm.DB) error {
	now := time.Now()
	ss.StoppedAt = &now
	return tx.Model(ss).Update("stopped_at", now).Error
}

// IncrementDownloadCount 增加下载次数
func (ss *ShareSession) IncrementDownloadCount(tx *gorm.DB) error {
	ss.CurrentDownloads++
	return tx.Model(ss).Update("current_downloads", gorm.Expr("current_downloads + ?", 1)).Error
}

// RemainingDownloads 获取剩余下载次数
func (ss *ShareSession) RemainingDownloads() int {
	if ss.MaxDownloads == 0 {
		return -1 // -1 表示无限制
	}
	remaining := ss.MaxDownloads - ss.CurrentDownloads
	if remaining < 0 {
		return 0
	}
	return remaining
}

// TimeUntilExpiry 获取距离过期的时间
func (ss *ShareSession) TimeUntilExpiry() time.Duration {
	return time.Until(ss.ExpiresAt)
}

// HoursUntilExpiry 获取距离过期的小时数
func (ss *ShareSession) HoursUntilExpiry() int {
	duration := ss.TimeUntilExpiry()
	hours := int(duration.Hours())
	if hours < 0 {
		return 0
	}
	return hours
}

// CanAccess 检查是否可以访问（综合检查）
func (ss *ShareSession) CanAccess() error {
	if ss.IsStopped() {
		return fmt.Errorf("share has been stopped by owner")
	}
	if ss.IsExpired() {
		return fmt.Errorf("share has expired")
	}
	if ss.IsExhausted() {
		return fmt.Errorf("download limit reached")
	}
	return nil
}

// DownloadProgress 获取下载进度百分比
func (ss *ShareSession) DownloadProgress() float64 {
	if ss.MaxDownloads == 0 {
		return 0 // 无限制返回 0
	}
	progress := float64(ss.CurrentDownloads) / float64(ss.MaxDownloads) * 100
	if progress > 100 {
		return 100
	}
	return progress
}
