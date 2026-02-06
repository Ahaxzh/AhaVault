package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email    string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password string    `gorm:"column:password_hash;type:varchar(255);not null" json:"-"` // JSON 中不返回密码
	Role     string    `gorm:"type:varchar(50);not null;default:'user'" json:"role"`
	Status   string    `gorm:"type:varchar(50);not null;default:'active'" json:"status"`

	// 存储配额
	StorageQuota int64 `gorm:"type:bigint;not null;default:10737418240" json:"storage_quota"` // 10GB
	StorageUsed  int64 `gorm:"type:bigint;not null;default:0" json:"storage_used"`

	// 时间戳
	CreatedAt   time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"not null;default:now()" json:"updated_at"`
	LastLoginAt *time.Time `gorm:"default:null" json:"last_login_at,omitempty"`

	// 关联关系
	Files          []FileMetadata  `gorm:"foreignKey:UserID" json:"-"`
	ShareSessions  []ShareSession  `gorm:"foreignKey:CreatorID" json:"-"`
	UploadSessions []UploadSession `gorm:"foreignKey:UserID" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate GORM 钩子：创建前
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// UserRole 用户角色常量
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// UserStatus 用户状态常量
const (
	StatusActive   = "active"
	StatusDisabled = "disabled"
)

// IsAdmin 检查是否为管理员
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsActive 检查账户是否激活
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// HasStorageSpace 检查是否有足够的存储空间
func (u *User) HasStorageSpace(requiredSize int64) bool {
	return u.StorageUsed+requiredSize <= u.StorageQuota
}

// AvailableStorage 获取可用存储空间
func (u *User) AvailableStorage() int64 {
	available := u.StorageQuota - u.StorageUsed
	if available < 0 {
		return 0
	}
	return available
}

// StorageUsagePercent 获取存储使用百分比
func (u *User) StorageUsagePercent() float64 {
	if u.StorageQuota == 0 {
		return 0
	}
	return float64(u.StorageUsed) / float64(u.StorageQuota) * 100
}

// UpdateLastLogin 更新最后登录时间
func (u *User) UpdateLastLogin(tx *gorm.DB) error {
	now := time.Now()
	u.LastLoginAt = &now
	return tx.Model(u).Update("last_login_at", now).Error
}
