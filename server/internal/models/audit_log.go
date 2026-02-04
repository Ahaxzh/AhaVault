package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AuditLog 审计日志模型
type AuditLog struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       *uuid.UUID     `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Action       string         `gorm:"type:varchar(100);not null;index" json:"action"`
	ResourceType string         `gorm:"type:varchar(50);index" json:"resource_type"`
	ResourceID   string         `gorm:"type:varchar(100);index" json:"resource_id"`
	IPAddress    string         `gorm:"type:inet" json:"ip_address"`
	UserAgent    string         `gorm:"type:text" json:"user_agent"`
	Details      datatypes.JSON `gorm:"type:jsonb" json:"details"`
	CreatedAt    time.Time      `gorm:"not null;default:now();index" json:"created_at"`

	// 关联关系
	User *User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "audit_logs"
}

// BeforeCreate GORM 钩子：创建前
func (al *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if al.ID == uuid.Nil {
		al.ID = uuid.New()
	}
	return nil
}

// 审计日志动作常量
const (
	ActionLogin          = "login"
	ActionLogout         = "logout"
	ActionRegister       = "register"
	ActionUploadFile     = "upload_file"
	ActionDownloadFile   = "download_file"
	ActionDeleteFile     = "delete_file"
	ActionRenameFile     = "rename_file"
	ActionCreateShare    = "create_share"
	ActionAccessShare    = "access_share"
	ActionStopShare      = "stop_share"
	ActionSaveToVault    = "save_to_vault"
	ActionBanFile        = "ban_file"
	ActionUnbanFile      = "unban_file"
	ActionDisableUser    = "disable_user"
	ActionEnableUser     = "enable_user"
	ActionUpdateSettings = "update_settings"
)

// 资源类型常量
const (
	ResourceTypeUser     = "user"
	ResourceTypeFile     = "file"
	ResourceTypeShare    = "share"
	ResourceTypeSettings = "settings"
)

// CreateLog 创建审计日志
func CreateLog(tx *gorm.DB, userID *uuid.UUID, action, resourceType, resourceID, ipAddress, userAgent string, details map[string]interface{}) error {
	var detailsJSON datatypes.JSON
	if details != nil {
		jsonBytes, err := json.Marshal(details)
		if err == nil {
			detailsJSON = datatypes.JSON(jsonBytes)
		}
	}

	log := &AuditLog{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Details:      detailsJSON,
	}

	return tx.Create(log).Error
}
