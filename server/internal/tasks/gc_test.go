// Package tasks 提供后台任务测试
//
// 本文件测试垃圾回收功能：
//   - 清理孤儿 blobs
//   - 清理过期分享
//   - 清理软删除文件
//
// 作者: AhaVault Team
// 创建时间: 2026-02-06
package tasks

import (
	"testing"
	"time"

	"ahavault/server/internal/models"
	"ahavault/server/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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
			password_hash TEXT NOT NULL,
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
	`).Error
	require.NoError(t, err)

	return db
}

func TestGarbageCollector_CleanOrphanBlobs(t *testing.T) {
	db := setupTestDB(t)
	storageEngine := storage.NewMemoryEngine()
	gc := NewGarbageCollector(db, storageEngine)

	// 创建孤儿 blob (ref_count = 0) - 使用有效的 64 字符哈希
	orphanHash := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	orphanBlob := &models.FileBlob{
		Hash:         orphanHash,
		StorePath:    "/data/storage/a1/b2/" + orphanHash,
		EncryptedDEK: "encrypted_dek",
		Size:         1024,
		RefCount:     0,
	}
	require.NoError(t, db.Create(orphanBlob).Error)
	// 显式更新 ref_count 为 0（SQLite 默认值可能覆盖）
	require.NoError(t, db.Model(orphanBlob).Update("ref_count", 0).Error)

	// 创建正常 blob (ref_count > 0)
	normalHash := "f1e2d3c4b5a6f1e2d3c4b5a6f1e2d3c4b5a6f1e2d3c4b5a6f1e2d3c4b5a6f1e2"
	normalBlob := &models.FileBlob{
		Hash:         normalHash,
		StorePath:    "/data/storage/f1/e2/" + normalHash,
		EncryptedDEK: "encrypted_dek",
		Size:         2048,
		RefCount:     1,
	}
	require.NoError(t, db.Create(normalBlob).Error)

	// 执行 GC
	result := gc.Run()

	// 验证结果
	assert.Equal(t, 1, result.OrphanBlobsDeleted)
	assert.Equal(t, int64(1024), result.SpaceReclaimed)
	assert.Empty(t, result.Errors)

	// 验证孤儿 blob 已被删除
	var count int64
	db.Model(&models.FileBlob{}).Where("hash = ?", orphanHash).Count(&count)
	assert.Equal(t, int64(0), count)

	// 验证正常 blob 仍存在
	db.Model(&models.FileBlob{}).Where("hash = ?", normalHash).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestGarbageCollector_CleanExpiredShares(t *testing.T) {
	db := setupTestDB(t)
	storageEngine := storage.NewMemoryEngine()
	gc := NewGarbageCollector(db, storageEngine)

	// 创建测试用户
	user := &models.User{
		Email:    "gc@test.com",
		Password: "hashed_password",
	}
	require.NoError(t, db.Create(user).Error)

	// 创建过期分享
	expiredShare := &models.ShareSession{
		ID:         uuid.New(),
		PickupCode: "EXPIRED1",
		CreatorID:  user.ID,
		ExpiresAt:  time.Now().Add(-1 * time.Hour),
	}
	require.NoError(t, db.Create(expiredShare).Error)

	// 创建有效分享
	validShare := &models.ShareSession{
		ID:         uuid.New(),
		PickupCode: "VALID123",
		CreatorID:  user.ID,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}
	require.NoError(t, db.Create(validShare).Error)

	// 执行 GC
	result := gc.Run()

	// 验证结果
	assert.Equal(t, 1, result.ExpiredSharesDeleted)
	assert.Empty(t, result.Errors)

	// 验证过期分享已被标记停止
	var share models.ShareSession
	db.First(&share, "pickup_code = ?", "EXPIRED1")
	assert.NotNil(t, share.StoppedAt)

	// 验证有效分享未被影响 - 重新查询
	var validShareCheck models.ShareSession
	db.Where("pickup_code = ?", "VALID123").First(&validShareCheck)
	assert.Nil(t, validShareCheck.StoppedAt)
}

func TestGarbageCollector_CleanSoftDeletedFiles(t *testing.T) {
	db := setupTestDB(t)
	storageEngine := storage.NewMemoryEngine()
	gc := NewGarbageCollector(db, storageEngine)

	// 创建测试用户
	user := &models.User{
		Email:    "softdel@test.com",
		Password: "hashed_password",
	}
	require.NoError(t, db.Create(user).Error)

	// 创建 blob - 使用有效的 64 字符哈希
	blobHash := "c1d2e3f4a5b6c1d2e3f4a5b6c1d2e3f4a5b6c1d2e3f4a5b6c1d2e3f4a5b6c1d2"
	blob := &models.FileBlob{
		Hash:         blobHash,
		StorePath:    "/data/storage/c1/d2/" + blobHash,
		EncryptedDEK: "encrypted_dek",
		Size:         512,
		RefCount:     1,
	}
	require.NoError(t, db.Create(blob).Error)

	// 创建软删除超过 7 天的文件
	deletedAt := time.Now().AddDate(0, 0, -10) // 10 天前删除
	oldDeletedFile := &models.FileMetadata{
		ID:           uuid.New(),
		UserID:       user.ID,
		FileBlobHash: blobHash,
		Filename:     "old_deleted.txt",
		Size:         512,
	}
	require.NoError(t, db.Create(oldDeletedFile).Error)
	db.Exec("UPDATE files_metadata SET deleted_at = ? WHERE id = ?", deletedAt, oldDeletedFile.ID)

	// 执行 GC
	result := gc.Run()

	// 验证结果
	assert.Equal(t, 1, result.SoftDeletedCleaned)

	// 验证文件已被永久删除
	var count int64
	db.Unscoped().Model(&models.FileMetadata{}).Where("id = ?", oldDeletedFile.ID).Count(&count)
	assert.Equal(t, int64(0), count)

	// 验证引用计数已减少
	var updatedBlob models.FileBlob
	db.First(&updatedBlob, "hash = ?", blobHash)
	assert.Equal(t, 0, updatedBlob.RefCount)
}
