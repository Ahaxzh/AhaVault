// Package services 提供业务逻辑服务层
//
// 本文件为 FileService 的单元测试，覆盖以下功能：
//   - 秒传检测（CheckInstantUpload）
//   - 文件元数据创建（CreateFileMetadata）
//   - 文件上传（UploadFile）
//   - 文件下载（DownloadFile）
//   - 文件删除（DeleteFile）
//   - 文件列表查询（ListFiles）
//
// 作者: AhaVault Team
// 创建时间: 2026-02-04
package services

import (
	"bytes"
	"io"
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

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		// 禁用外键约束，简化测试
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)

	// 手动创建简化的表结构（兼容 SQLite）
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

		CREATE INDEX idx_user_files ON files_metadata(user_id, deleted_at);
		CREATE INDEX idx_blob_hash ON files_metadata(file_blob_hash);
		CREATE INDEX idx_pickup_code ON share_sessions(pickup_code);
		CREATE INDEX idx_creator ON share_sessions(creator_id);
	`).Error
	require.NoError(t, err)

	return db
}

// createTestUser 创建测试用户
func createTestUser(t *testing.T, db *gorm.DB) *models.User {
	user := &models.User{
		Email:        "test@example.com",
		Password:     "hashed_password",
		Role:         models.RoleUser,
		Status:       models.StatusActive,
		StorageQuota: 10 * 1024 * 1024 * 1024, // 10GB
		StorageUsed:  0,
	}
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

// TestCheckInstantUpload 测试秒传检测功能
//
// 测试场景：
//  1. 文件不存在 - 返回 false
//  2. 文件存在且未被禁止 - 返回 true
//  3. 文件存在但被禁止 - 返回错误
//  4. 无效哈希值 - 返回错误
func TestCheckInstantUpload(t *testing.T) {
	db := setupTestDB(t)
	storageEngine := storage.NewMemoryEngine()
	kek := []byte("test-master-key-1234567890123456") // 32 bytes
	service := NewFileService(db, storageEngine, kek)
	user := createTestUser(t, db)

	// 准备测试数据：创建一个已存在的文件
	existingHash := "aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899"
	existingBlob := &models.FileBlob{
		Hash:         existingHash,
		StorePath:    "aa/bb/aabbccdd...",
		EncryptedDEK: "encrypted_dek_data",
		Size:         1024,
		RefCount:     1,
		IsBanned:     false,
	}
	require.NoError(t, db.Create(existingBlob).Error)

	// 创建一个被禁止的文件
	bannedHash := "1122334455667788990011223344556677889900112233445566778899001122"
	bannedBlob := &models.FileBlob{
		Hash:         bannedHash,
		StorePath:    "11/22/11223344...",
		EncryptedDEK: "encrypted_dek_data",
		Size:         2048,
		RefCount:     1,
		IsBanned:     true,
		BanReason:    "违规内容",
	}
	require.NoError(t, db.Create(bannedBlob).Error)

	tests := []struct {
		name        string
		hash        string
		wantExists  bool
		wantBlob    *models.FileBlob
		wantErr     bool
		errContains string
	}{
		{
			name:       "文件不存在 - 需要上传",
			hash:       "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00",
			wantExists: false,
			wantBlob:   nil,
			wantErr:    false,
		},
		{
			name:       "文件存在且未被禁止 - 可以秒传",
			hash:       existingHash,
			wantExists: true,
			wantBlob:   existingBlob,
			wantErr:    false,
		},
		{
			name:        "文件被禁止 - 返回错误",
			hash:        bannedHash,
			wantExists:  false,
			wantBlob:    nil,
			wantErr:     true,
			errContains: "banned",
		},
		{
			name:        "无效哈希值 - 长度不足",
			hash:        "invalid",
			wantExists:  false,
			wantBlob:    nil,
			wantErr:     true,
			errContains: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, blob, err := service.CheckInstantUpload(tt.hash, user.ID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantExists, exists)
			if tt.wantBlob != nil {
				require.NotNil(t, blob)
				assert.Equal(t, tt.wantBlob.Hash, blob.Hash)
			} else if !tt.wantExists {
				assert.Nil(t, blob)
			}
		})
	}
}

// TestCreateFileMetadata 测试文件元数据创建（秒传）
//
// 测试场景：
//  1. 正常创建 - 引用计数增加，存储使用量增加
//  2. 用户不存在 - 返回错误
//  3. 存储空间不足 - 返回错误
//  4. 文件 blob 不存在 - 返回错误
func TestCreateFileMetadata(t *testing.T) {
	db := setupTestDB(t)
	storageEngine := storage.NewMemoryEngine()
	kek := []byte("test-master-key-1234567890123456")
	service := NewFileService(db, storageEngine, kek)
	user := createTestUser(t, db)

	// 准备测试数据
	existingHash := "aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899"
	existingBlob := &models.FileBlob{
		Hash:         existingHash,
		StorePath:    "aa/bb/aabbccdd...",
		EncryptedDEK: "encrypted_dek_data",
		Size:         1024,
		RefCount:     1,
		IsBanned:     false,
	}
	require.NoError(t, db.Create(existingBlob).Error)

	tests := []struct {
		name        string
		userID      uuid.UUID
		hash        string
		filename    string
		size        int64
		wantErr     bool
		errContains string
		checkRefCount bool
		expectedRefCount int
	}{
		{
			name:             "正常秒传 - 引用计数增加",
			userID:           user.ID,
			hash:             existingHash,
			filename:         "test.pdf",
			size:             1024,
			wantErr:          false,
			checkRefCount:    true,
			expectedRefCount: 2,
		},
		{
			name:        "用户不存在",
			userID:      uuid.New(),
			hash:        existingHash,
			filename:    "test.pdf",
			size:        1024,
			wantErr:     true,
			errContains: "failed to get user",
		},
		{
			name:        "存储空间不足",
			userID:      user.ID,
			hash:        existingHash,
			filename:    "large.zip",
			size:        100 * 1024 * 1024 * 1024, // 100GB
			wantErr:     true,
			errContains: "insufficient storage",
		},
		{
			name:        "文件 blob 不存在",
			userID:      user.ID,
			hash:        "0000000000000000000000000000000000000000000000000000000000000000",
			filename:    "test.txt",
			size:        100,
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 记录初始状态
			var initialUser models.User
			db.First(&initialUser, user.ID)
			initialStorageUsed := initialUser.StorageUsed

			metadata, err := service.CreateFileMetadata(tt.userID, tt.hash, tt.filename, tt.size)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, metadata)
			assert.Equal(t, tt.filename, metadata.Filename)
			assert.Equal(t, tt.size, metadata.Size)
			assert.Equal(t, tt.hash, metadata.FileBlobHash)

			// 验证引用计数
			if tt.checkRefCount {
				var blob models.FileBlob
				db.First(&blob, "hash = ?", tt.hash)
				assert.Equal(t, tt.expectedRefCount, blob.RefCount)
			}

			// 验证存储使用量
			var updatedUser models.User
			db.First(&updatedUser, user.ID)
			assert.Equal(t, initialStorageUsed+tt.size, updatedUser.StorageUsed)
		})
	}
}

// TestUploadFile 测试文件上传功能
//
// 测试场景：
//  1. 上传新文件 - 创建 blob 和 metadata
//  2. 二次秒传检测 - 哈希计算后发现已存在
//  3. 存储空间不足
func TestUploadFile(t *testing.T) {
	db := setupTestDB(t)
	storageEngine := storage.NewMemoryEngine() // 使用内存存储引擎
	kek := []byte("test-master-key-1234567890123456")
	service := NewFileService(db, storageEngine, kek)
	user := createTestUser(t, db)

	tests := []struct {
		name        string
		filename    string
		content     []byte
		wantErr     bool
		errContains string
	}{
		{
			name:     "上传新文件",
			filename: "test.txt",
			content:  []byte("Hello, World!"),
			wantErr:  false,
		},
		{
			name:     "上传另一个文件",
			filename: "document.pdf",
			content:  bytes.Repeat([]byte("A"), 1024),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.content)
			size := int64(len(tt.content))

			metadata, err := service.UploadFile(user.ID, tt.filename, size, reader)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, metadata)
			assert.Equal(t, tt.filename, metadata.Filename)
			assert.Equal(t, size, metadata.Size)

			// 验证文件 blob 已创建
			var blob models.FileBlob
			err = db.First(&blob, "hash = ?", metadata.FileBlobHash).Error
			require.NoError(t, err)
			assert.Equal(t, 1, blob.RefCount)

			// 验证存储引擎中文件存在
			exists, err := storageEngine.Exists(blob.Hash)
			require.NoError(t, err)
			assert.True(t, exists)
		})
	}
}

// TestUploadFile_SecondaryInstantUpload 测试二次秒传检测
//
// 场景：用户上传文件时，哈希计算后发现已存在，自动转为秒传
func TestUploadFile_SecondaryInstantUpload(t *testing.T) {
	db := setupTestDB(t)
	storageEngine := storage.NewMemoryEngine()
	kek := []byte("test-master-key-1234567890123456")
	service := NewFileService(db, storageEngine, kek)
	user := createTestUser(t, db)

	content := []byte("Duplicate file content")

	// 第一次上传
	metadata1, err := service.UploadFile(user.ID, "file1.txt", int64(len(content)), bytes.NewReader(content))
	require.NoError(t, err)

	// 第二次上传相同内容（应该触发二次秒传）
	metadata2, err := service.UploadFile(user.ID, "file2.txt", int64(len(content)), bytes.NewReader(content))
	require.NoError(t, err)

	// 验证哈希相同
	assert.Equal(t, metadata1.FileBlobHash, metadata2.FileBlobHash)

	// 验证引用计数增加
	var blob models.FileBlob
	db.First(&blob, "hash = ?", metadata1.FileBlobHash)
	assert.Equal(t, 2, blob.RefCount)

	// 验证只有一份物理文件
	exists, err := storageEngine.Exists(blob.Hash)
	require.NoError(t, err)
	assert.True(t, exists)
}

// TestDownloadFile 测试文件下载功能
//
// 测试场景：
//  1. 正常下载 - 解密并返回内容
//  2. 文件不存在
//  3. 权限不足（不是文件所有者）
//  4. 文件已过期
func TestDownloadFile(t *testing.T) {
	db := setupTestDB(t)
	storageEngine := storage.NewMemoryEngine()
	kek := []byte("test-master-key-1234567890123456")
	service := NewFileService(db, storageEngine, kek)
	user := createTestUser(t, db)
	otherUser := &models.User{
		Email:        "other@example.com",
		Password:     "password",
		Role:         models.RoleUser,
		Status:       models.StatusActive,
		StorageQuota: 10 * 1024 * 1024 * 1024,
		StorageUsed:  0,
	}
	require.NoError(t, db.Create(otherUser).Error)

	// 上传测试文件
	testContent := []byte("Download test content")
	metadata, err := service.UploadFile(user.ID, "download-test.txt", int64(len(testContent)), bytes.NewReader(testContent))
	require.NoError(t, err)

	// 创建一个已过期的文件
	expiredTime := time.Now().Add(-1 * time.Hour)
	expiredMetadata := &models.FileMetadata{
		UserID:       user.ID,
		FileBlobHash: metadata.FileBlobHash,
		Filename:     "expired.txt",
		Size:         int64(len(testContent)),
		ExpiresAt:    &expiredTime,
	}
	require.NoError(t, db.Create(expiredMetadata).Error)

	tests := []struct {
		name        string
		fileID      uuid.UUID
		userID      uuid.UUID
		wantErr     bool
		errContains string
		checkContent bool
		expectedContent []byte
	}{
		{
			name:            "正常下载",
			fileID:          metadata.ID,
			userID:          user.ID,
			wantErr:         false,
			checkContent:    true,
			expectedContent: testContent,
		},
		{
			name:        "文件不存在",
			fileID:      uuid.New(),
			userID:      user.ID,
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:        "权限不足 - 其他用户",
			fileID:      metadata.ID,
			userID:      otherUser.ID,
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:        "文件已过期",
			fileID:      expiredMetadata.ID,
			userID:      user.ID,
			wantErr:     true,
			errContains: "expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, meta, err := service.DownloadFile(tt.fileID, tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, reader)
			require.NotNil(t, meta)
			defer reader.Close()

			// 验证内容
			if tt.checkContent {
				content, err := io.ReadAll(reader)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedContent, content)
			}
		})
	}
}

// TestDeleteFile 测试文件删除功能
//
// 测试场景：
//  1. 正常删除 - 软删除，引用计数减少，存储使用量减少
//  2. 文件不存在
//  3. 重复删除（已删除的文件）
//  4. 验证引用计数正确递减
func TestDeleteFile(t *testing.T) {
	db := setupTestDB(t)
	storageEngine := storage.NewMemoryEngine()
	kek := []byte("test-master-key-1234567890123456")
	service := NewFileService(db, storageEngine, kek)
	user := createTestUser(t, db)

	// 上传测试文件
	testContent := []byte("Delete test content")
	metadata, err := service.UploadFile(user.ID, "delete-test.txt", int64(len(testContent)), bytes.NewReader(testContent))
	require.NoError(t, err)

	// 创建第二个引用
	metadata2, err := service.CreateFileMetadata(user.ID, metadata.FileBlobHash, "copy.txt", int64(len(testContent)))
	require.NoError(t, err)

	// 验证初始引用计数
	var blob models.FileBlob
	db.First(&blob, "hash = ?", metadata.FileBlobHash)
	assert.Equal(t, 2, blob.RefCount)

	tests := []struct {
		name               string
		fileID             uuid.UUID
		userID             uuid.UUID
		wantErr            bool
		errContains        string
		expectedRefCount   int
		checkStorageUpdate bool
	}{
		{
			name:               "删除第一个文件 - 引用计数 2 -> 1",
			fileID:             metadata.ID,
			userID:             user.ID,
			wantErr:            false,
			expectedRefCount:   1,
			checkStorageUpdate: true,
		},
		{
			name:               "删除第二个文件 - 引用计数 1 -> 0",
			fileID:             metadata2.ID,
			userID:             user.ID,
			wantErr:            false,
			expectedRefCount:   0,
			checkStorageUpdate: true,
		},
		{
			name:        "文件不存在",
			fileID:      uuid.New(),
			userID:      user.ID,
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 记录初始存储使用量
			var initialUser models.User
			db.First(&initialUser, user.ID)
			initialStorageUsed := initialUser.StorageUsed

			err := service.DeleteFile(tt.fileID, tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)

			// 验证软删除
			var deletedMetadata models.FileMetadata
			db.Unscoped().First(&deletedMetadata, tt.fileID)
			assert.NotNil(t, deletedMetadata.DeletedAt)

			// 验证引用计数
			var updatedBlob models.FileBlob
			db.First(&updatedBlob, "hash = ?", metadata.FileBlobHash)
			assert.Equal(t, tt.expectedRefCount, updatedBlob.RefCount)

			// 验证存储使用量
			if tt.checkStorageUpdate {
				var updatedUser models.User
				db.First(&updatedUser, user.ID)
				assert.Equal(t, initialStorageUsed-int64(len(testContent)), updatedUser.StorageUsed)
			}
		})
	}
}

// TestListFiles 测试文件列表查询功能
//
// 测试场景：
//  1. 分页查询
//  2. 空列表
//  3. 排序验证（最新的在前）
//  4. 软删除文件不显示
func TestListFiles(t *testing.T) {
	db := setupTestDB(t)
	storageEngine := storage.NewMemoryEngine()
	kek := []byte("test-master-key-1234567890123456")
	service := NewFileService(db, storageEngine, kek)
	user := createTestUser(t, db)

	// 创建多个测试文件
	for i := 1; i <= 15; i++ {
		content := []byte("File content " + string(rune(i)))
		_, err := service.UploadFile(user.ID, "file"+string(rune('0'+i))+".txt", int64(len(content)), bytes.NewReader(content))
		require.NoError(t, err)
		time.Sleep(1 * time.Millisecond) // 确保时间戳不同
	}

	// 软删除一个文件
	var firstFile models.FileMetadata
	db.Where("user_id = ?", user.ID).First(&firstFile)
	err := service.DeleteFile(firstFile.ID, user.ID)
	require.NoError(t, err)

	tests := []struct {
		name          string
		page          int
		pageSize      int
		expectedCount int
		expectedTotal int64
	}{
		{
			name:          "第一页 - 10 条",
			page:          1,
			pageSize:      10,
			expectedCount: 10,
			expectedTotal: 14, // 15 - 1 (已删除)
		},
		{
			name:          "第二页 - 4 条",
			page:          2,
			pageSize:      10,
			expectedCount: 4,
			expectedTotal: 14,
		},
		{
			name:          "小分页 - 5 条",
			page:          1,
			pageSize:      5,
			expectedCount: 5,
			expectedTotal: 14,
		},
		{
			name:          "超出范围的页码",
			page:          100,
			pageSize:      10,
			expectedCount: 0,
			expectedTotal: 14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, total, err := service.ListFiles(user.ID, tt.page, tt.pageSize)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedTotal, total)
			assert.Equal(t, tt.expectedCount, len(files))

			// 验证排序（最新的在前）
			if len(files) > 1 {
				for i := 0; i < len(files)-1; i++ {
					assert.True(t, files[i].CreatedAt.After(files[i+1].CreatedAt) ||
						files[i].CreatedAt.Equal(files[i+1].CreatedAt))
				}
			}

			// 验证没有已删除的文件
			for _, file := range files {
				assert.Nil(t, file.DeletedAt)
			}
		})
	}
}

// TestListFiles_Empty 测试空文件列表
func TestListFiles_Empty(t *testing.T) {
	db := setupTestDB(t)
	storageEngine := storage.NewMemoryEngine()
	kek := []byte("test-master-key-1234567890123456")
	service := NewFileService(db, storageEngine, kek)
	user := createTestUser(t, db)

	files, total, err := service.ListFiles(user.ID, 1, 10)

	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, files)
}

// TestUploadFile_InsufficientStorage 测试存储空间不足场景
func TestUploadFile_InsufficientStorage(t *testing.T) {
	db := setupTestDB(t)
	storageEngine := storage.NewMemoryEngine()
	kek := []byte("test-master-key-1234567890123456")
	service := NewFileService(db, storageEngine, kek)

	// 创建存储空间不足的用户
	user := &models.User{
		Email:        "lowstorage@example.com",
		Password:     "password",
		Role:         models.RoleUser,
		Status:       models.StatusActive,
		StorageQuota: 100, // 仅 100 字节
		StorageUsed:  0,
	}
	require.NoError(t, db.Create(user).Error)

	content := bytes.Repeat([]byte("A"), 200) // 200 字节
	_, err := service.UploadFile(user.ID, "large.txt", int64(len(content)), bytes.NewReader(content))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient storage")
}
