// Package handlers 提供测试辅助函数
//
// 本文件包含所有 handler 测试共用的辅助函数
//
// 作者: AhaVault Team
// 创建时间: 2026-02-04
package handlers

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建测试数据库
//
// 使用内存 SQLite 数据库，自动创建表结构
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)

	// 创建表结构
	err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			status TEXT NOT NULL DEFAULT 'active',
			storage_quota INTEGER NOT NULL DEFAULT 10737418240,
			storage_used INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			last_login_at DATETIME
		);

		CREATE TABLE file_blobs (
			hash TEXT PRIMARY KEY,
			store_path TEXT NOT NULL,
			encrypted_dek TEXT NOT NULL,
			size INTEGER NOT NULL,
			mime_type TEXT,
			ref_count INTEGER NOT NULL DEFAULT 1,
			is_banned INTEGER NOT NULL DEFAULT 0,
			ban_reason TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE files_metadata (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			file_blob_hash TEXT NOT NULL,
			filename TEXT NOT NULL,
			size INTEGER NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME,
			deleted_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (file_blob_hash) REFERENCES file_blobs(hash)
		);

		CREATE TABLE share_sessions (
			id TEXT PRIMARY KEY,
			pickup_code TEXT NOT NULL UNIQUE,
			creator_id TEXT NOT NULL,
			password_hash TEXT,
			max_downloads INTEGER NOT NULL DEFAULT 0,
			current_downloads INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME NOT NULL,
			stopped_at DATETIME,
			FOREIGN KEY (creator_id) REFERENCES users(id)
		);

		CREATE TABLE share_files (
			id TEXT PRIMARY KEY,
			share_id TEXT NOT NULL,
			file_id TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (share_id) REFERENCES share_sessions(id),
			FOREIGN KEY (file_id) REFERENCES files_metadata(id)
		);
	`).Error
	require.NoError(t, err)

	return db
}
