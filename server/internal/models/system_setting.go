package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// SystemSetting 系统配置模型
type SystemSetting struct {
	Key         string    `gorm:"type:varchar(100);primary_key" json:"key"`
	Value       string    `gorm:"type:text;not null" json:"value"`
	Description string    `gorm:"type:text" json:"description"`
	UpdatedAt   time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 指定表名
func (SystemSetting) TableName() string {
	return "system_settings"
}

// 系统配置键常量
const (
	SettingRegistrationEnabled = "registration_enabled"
	SettingInviteCodeRequired  = "invite_code_required"
	SettingMaxFileSize         = "max_file_size"
	SettingStorageType         = "storage_type"
	SettingDefaultUserQuota    = "default_user_quota"
	SettingShareCodeLength     = "share_code_length"
	SettingGCRetentionDays     = "gc_retention_days"
)

// GetValue 获取配置值
func GetValue(tx *gorm.DB, key string) (string, error) {
	var setting SystemSetting
	if err := tx.Where("key = ?", key).First(&setting).Error; err != nil {
		return "", err
	}
	return setting.Value, nil
}

// SetValue 设置配置值
func SetValue(tx *gorm.DB, key, value string) error {
	return tx.Model(&SystemSetting{}).
		Where("key = ?", key).
		Updates(map[string]interface{}{
			"value":      value,
			"updated_at": time.Now(),
		}).Error
}

// GetBool 获取布尔值配置
func GetBool(tx *gorm.DB, key string) (bool, error) {
	value, err := GetValue(tx, key)
	if err != nil {
		return false, err
	}
	return value == "true", nil
}

// GetInt 获取整数配置
func GetInt(tx *gorm.DB, key string) (int, error) {
	value, err := GetValue(tx, key)
	if err != nil {
		return 0, err
	}
	var result int
	_, err = fmt.Sscanf(value, "%d", &result)
	return result, err
}

// GetInt64 获取 int64 配置
func GetInt64(tx *gorm.DB, key string) (int64, error) {
	value, err := GetValue(tx, key)
	if err != nil {
		return 0, err
	}
	var result int64
	_, err = fmt.Sscanf(value, "%d", &result)
	return result, err
}
