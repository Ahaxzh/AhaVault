// Package services 提供业务逻辑服务层
//
// 本文件为 ShareService 的单元测试，覆盖以下功能：
//   - 创建分享（CreateShare）
//   - 通过取件码获取分享（GetShareByCode）
//   - 增加下载次数（IncrementDownload）
//   - 停止分享（StopShare）
//   - 转存到文件柜（SaveToVault）
//   - 获取我的分享列表（ListMyShares）
//
// 作者: AhaVault Team
// 创建时间: 2026-02-04
package services

import (
	"bytes"
	"testing"
	"time"

	"ahavault/server/internal/models"
	"ahavault/server/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// setupShareTestEnv 创建分享测试环境
func setupShareTestEnv(t *testing.T) (*ShareService, *FileService, *models.User, *gorm.DB) {
	db := setupTestDB(t)
	storageEngine := storage.NewMemoryEngine()
	kek := []byte("test-master-key-1234567890123456")

	fileService := NewFileService(db, storageEngine, kek)
	shareService := NewShareService(db, fileService)

	user := createTestUser(t, db)

	return shareService, fileService, user, db
}

// TestCreateShare 测试创建分享功能
//
// 测试场景：
//  1. 正常创建分享 - 单个文件
//  2. 创建分享 - 多个文件
//  3. 创建分享 - 带密码
//  4. 创建分享 - 设置下载次数限制
//  5. 空文件列表 - 返回错误
//  6. 文件不存在 - 返回错误
//  7. 文件不属于用户 - 返回错误
func TestCreateShare(t *testing.T) {
	shareService, fileService, user, db := setupShareTestEnv(t)

	// 上传测试文件
	content1 := []byte("Test file 1")
	file1, err := fileService.UploadFile(user.ID, "test1.txt", int64(len(content1)), bytes.NewReader(content1))
	require.NoError(t, err)

	content2 := []byte("Test file 2")
	file2, err := fileService.UploadFile(user.ID, "test2.txt", int64(len(content2)), bytes.NewReader(content2))
	require.NoError(t, err)

	// 创建另一个用户的文件
	otherUser := &models.User{
		Email:        "other@example.com",
		Password:     "password",
		Role:         models.RoleUser,
		Status:       models.StatusActive,
		StorageQuota: 10 * 1024 * 1024 * 1024,
		StorageUsed:  0,
	}
	require.NoError(t, db.Create(otherUser).Error)

	content3 := []byte("Other user file")
	otherFile, err := fileService.UploadFile(otherUser.ID, "other.txt", int64(len(content3)), bytes.NewReader(content3))
	require.NoError(t, err)

	tests := []struct {
		name        string
		req         *CreateShareRequest
		wantErr     bool
		errContains string
		checkCode   bool
		checkFiles  int
	}{
		{
			name: "正常创建分享 - 单个文件",
			req: &CreateShareRequest{
				FileIDs:      []uuid.UUID{file1.ID},
				ExpiresIn:    24 * time.Hour,
				MaxDownloads: 10,
				Password:     "",
			},
			wantErr:    false,
			checkCode:  true,
			checkFiles: 1,
		},
		{
			name: "创建分享 - 多个文件",
			req: &CreateShareRequest{
				FileIDs:      []uuid.UUID{file1.ID, file2.ID},
				ExpiresIn:    7 * 24 * time.Hour,
				MaxDownloads: 0,
				Password:     "",
			},
			wantErr:    false,
			checkCode:  true,
			checkFiles: 2,
		},
		{
			name: "创建分享 - 带密码",
			req: &CreateShareRequest{
				FileIDs:      []uuid.UUID{file1.ID},
				ExpiresIn:    24 * time.Hour,
				MaxDownloads: 5,
				Password:     "secret123",
			},
			wantErr:    false,
			checkCode:  true,
			checkFiles: 1,
		},
		{
			name: "空文件列表",
			req: &CreateShareRequest{
				FileIDs:      []uuid.UUID{},
				ExpiresIn:    24 * time.Hour,
				MaxDownloads: 10,
				Password:     "",
			},
			wantErr:     true,
			errContains: "no files",
		},
		{
			name: "文件不存在",
			req: &CreateShareRequest{
				FileIDs:      []uuid.UUID{uuid.New()},
				ExpiresIn:    24 * time.Hour,
				MaxDownloads: 10,
				Password:     "",
			},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name: "文件不属于用户",
			req: &CreateShareRequest{
				FileIDs:      []uuid.UUID{otherFile.ID},
				ExpiresIn:    24 * time.Hour,
				MaxDownloads: 10,
				Password:     "",
			},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := shareService.CreateShare(user.ID, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, session)

			// 验证取件码
			if tt.checkCode {
				assert.Len(t, session.PickupCode, 8)
				assert.NoError(t, ValidatePickupCode(session.PickupCode, 8))
			}

			// 验证文件关联
			if tt.checkFiles > 0 {
				var shareFiles []models.ShareFile
				err := db.Where("share_id = ?", session.ID).Find(&shareFiles).Error
				require.NoError(t, err)
				assert.Len(t, shareFiles, tt.checkFiles)
			}

			// 验证密码
			if tt.req.Password != "" {
				assert.NotEmpty(t, session.PasswordHash)
				err := bcrypt.CompareHashAndPassword([]byte(session.PasswordHash), []byte(tt.req.Password))
				assert.NoError(t, err)
			}
		})
	}
}

// TestGetShareByCode 测试通过取件码获取分享
//
// 测试场景：
//  1. 正常获取 - 无密码
//  2. 正常获取 - 带正确密码
//  3. 密码错误
//  4. 未提供密码（需要密码）
//  5. 取件码不存在
//  6. 取件码格式无效
//  7. 分享已过期
//  8. 下载次数已达上限
func TestGetShareByCode(t *testing.T) {
	shareService, fileService, user, db := setupShareTestEnv(t)

	// 上传测试文件
	content := []byte("Shared file content")
	file, err := fileService.UploadFile(user.ID, "shared.txt", int64(len(content)), bytes.NewReader(content))
	require.NoError(t, err)

	// 创建普通分享
	normalShare, err := shareService.CreateShare(user.ID, &CreateShareRequest{
		FileIDs:      []uuid.UUID{file.ID},
		ExpiresIn:    24 * time.Hour,
		MaxDownloads: 10,
		Password:     "",
	})
	require.NoError(t, err)

	// 创建带密码的分享
	passwordShare, err := shareService.CreateShare(user.ID, &CreateShareRequest{
		FileIDs:      []uuid.UUID{file.ID},
		ExpiresIn:    24 * time.Hour,
		MaxDownloads: 10,
		Password:     "mypassword",
	})
	require.NoError(t, err)

	// 创建已过期的分享
	expiredShare := &models.ShareSession{
		PickupCode:       "EXPIRED2",
		CreatorID:        user.ID,
		MaxDownloads:     10,
		CurrentDownloads: 0,
		ExpiresAt:        time.Now().Add(-1 * time.Hour),
	}
	require.NoError(t, db.Create(expiredShare).Error)

	// 创建下载次数已达上限的分享
	exhaustedShare := &models.ShareSession{
		PickupCode:       "EXHAUST2",
		CreatorID:        user.ID,
		MaxDownloads:     5,
		CurrentDownloads: 5,
		ExpiresAt:        time.Now().Add(24 * time.Hour),
	}
	require.NoError(t, db.Create(exhaustedShare).Error)

	tests := []struct {
		name        string
		pickupCode  string
		password    string
		wantErr     bool
		errContains string
		checkFiles  bool
	}{
		{
			name:       "正常获取 - 无密码",
			pickupCode: normalShare.PickupCode,
			password:   "",
			wantErr:    false,
			checkFiles: true,
		},
		{
			name:       "正常获取 - 带正确密码",
			pickupCode: passwordShare.PickupCode,
			password:   "mypassword",
			wantErr:    false,
			checkFiles: true,
		},
		{
			name:        "密码错误",
			pickupCode:  passwordShare.PickupCode,
			password:    "wrongpassword",
			wantErr:     true,
			errContains: "invalid password",
		},
		{
			name:        "未提供密码（需要密码）",
			pickupCode:  passwordShare.PickupCode,
			password:    "",
			wantErr:     true,
			errContains: "password required",
		},
		{
			name:        "取件码不存在",
			pickupCode:  "ABCD2345",
			password:    "",
			wantErr:     true,
			errContains: "invalid pickup code",
		},
		{
			name:        "取件码格式无效",
			pickupCode:  "invalid",
			password:    "",
			wantErr:     true,
			errContains: "invalid",
		},
		{
			name:        "分享已过期",
			pickupCode:  expiredShare.PickupCode,
			password:    "",
			wantErr:     true,
			errContains: "expired",
		},
		{
			name:        "下载次数已达上限",
			pickupCode:  exhaustedShare.PickupCode,
			password:    "",
			wantErr:     true,
			errContains: "download limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, files, err := shareService.GetShareByCode(tt.pickupCode, tt.password)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, session)

			if tt.checkFiles {
				require.NotEmpty(t, files)
			}
		})
	}
}

// TestIncrementDownload 测试增加下载次数
func TestIncrementDownload(t *testing.T) {
	shareService, fileService, user, _ := setupShareTestEnv(t)

	// 上传测试文件
	content := []byte("Test file")
	file, err := fileService.UploadFile(user.ID, "test.txt", int64(len(content)), bytes.NewReader(content))
	require.NoError(t, err)

	// 创建分享
	session, err := shareService.CreateShare(user.ID, &CreateShareRequest{
		FileIDs:      []uuid.UUID{file.ID},
		ExpiresIn:    24 * time.Hour,
		MaxDownloads: 10,
		Password:     "",
	})
	require.NoError(t, err)

	// 初始下载次数应为 0
	assert.Equal(t, 0, session.CurrentDownloads)

	// 增加下载次数
	err = shareService.IncrementDownload(session.ID)
	require.NoError(t, err)

	// 验证下载次数增加
	var updated models.ShareSession
	err = shareService.db.First(&updated, session.ID).Error
	require.NoError(t, err)
	assert.Equal(t, 1, updated.CurrentDownloads)

	// 再次增加
	err = shareService.IncrementDownload(session.ID)
	require.NoError(t, err)

	err = shareService.db.First(&updated, session.ID).Error
	require.NoError(t, err)
	assert.Equal(t, 2, updated.CurrentDownloads)
}

// TestStopShare 测试停止分享
func TestStopShare(t *testing.T) {
	shareService, fileService, user, db := setupShareTestEnv(t)

	// 创建另一个用户
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
	content := []byte("Test file")
	file, err := fileService.UploadFile(user.ID, "test.txt", int64(len(content)), bytes.NewReader(content))
	require.NoError(t, err)

	// 创建分享
	session, err := shareService.CreateShare(user.ID, &CreateShareRequest{
		FileIDs:      []uuid.UUID{file.ID},
		ExpiresIn:    24 * time.Hour,
		MaxDownloads: 10,
		Password:     "",
	})
	require.NoError(t, err)

	tests := []struct {
		name        string
		shareID     uuid.UUID
		userID      uuid.UUID
		wantErr     bool
		errContains string
	}{
		{
			name:    "正常停止分享",
			shareID: session.ID,
			userID:  user.ID,
			wantErr: false,
		},
		{
			name:        "分享不存在",
			shareID:     uuid.New(),
			userID:      user.ID,
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:        "非创建者停止分享",
			shareID:     session.ID,
			userID:      otherUser.ID,
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := shareService.StopShare(tt.shareID, tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)

			// 验证分享已停止
			var stopped models.ShareSession
			err = db.First(&stopped, session.ID).Error
			require.NoError(t, err)
			assert.True(t, stopped.IsStopped())
		})
	}
}

// TestSaveToVault 测试转存到文件柜
func TestSaveToVault(t *testing.T) {
	shareService, fileService, user, db := setupShareTestEnv(t)

	// 创建另一个用户（转存者）
	receiver := &models.User{
		Email:        "receiver@example.com",
		Password:     "password",
		Role:         models.RoleUser,
		Status:       models.StatusActive,
		StorageQuota: 10 * 1024 * 1024 * 1024,
		StorageUsed:  0,
	}
	require.NoError(t, db.Create(receiver).Error)

	// 上传测试文件
	content := []byte("Shared file for vault")
	file, err := fileService.UploadFile(user.ID, "shared.txt", int64(len(content)), bytes.NewReader(content))
	require.NoError(t, err)

	// 创建分享
	session, err := shareService.CreateShare(user.ID, &CreateShareRequest{
		FileIDs:      []uuid.UUID{file.ID},
		ExpiresIn:    24 * time.Hour,
		MaxDownloads: 10,
		Password:     "vault123",
	})
	require.NoError(t, err)

	// 转存到文件柜
	savedIDs, err := shareService.SaveToVault(
		session.PickupCode,
		"vault123",
		[]uuid.UUID{file.ID},
		receiver.ID,
	)

	require.NoError(t, err)
	require.Len(t, savedIDs, 1)

	// 验证文件已转存
	var receiverFile models.FileMetadata
	err = db.First(&receiverFile, savedIDs[0]).Error
	require.NoError(t, err)
	assert.Equal(t, receiver.ID, receiverFile.UserID)
	assert.Equal(t, file.FileBlobHash, receiverFile.FileBlobHash)

	// 验证引用计数增加
	var blob models.FileBlob
	err = db.First(&blob, "hash = ?", file.FileBlobHash).Error
	require.NoError(t, err)
	assert.Equal(t, 2, blob.RefCount) // 原文件 + 转存文件

	// 验证下载次数增加
	var updated models.ShareSession
	err = db.First(&updated, session.ID).Error
	require.NoError(t, err)
	assert.Equal(t, 1, updated.CurrentDownloads)
}

// TestListMyShares 测试获取我的分享列表
func TestListMyShares(t *testing.T) {
	shareService, fileService, user, _ := setupShareTestEnv(t)

	// 上传测试文件
	content := []byte("Test file")
	file, err := fileService.UploadFile(user.ID, "test.txt", int64(len(content)), bytes.NewReader(content))
	require.NoError(t, err)

	// 创建多个分享
	for i := 0; i < 15; i++ {
		_, err := shareService.CreateShare(user.ID, &CreateShareRequest{
			FileIDs:      []uuid.UUID{file.ID},
			ExpiresIn:    24 * time.Hour,
			MaxDownloads: 10,
			Password:     "",
		})
		require.NoError(t, err)
		time.Sleep(1 * time.Millisecond) // 确保时间戳不同
	}

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
			expectedTotal: 15,
		},
		{
			name:          "第二页 - 5 条",
			page:          2,
			pageSize:      10,
			expectedCount: 5,
			expectedTotal: 15,
		},
		{
			name:          "小分页 - 5 条",
			page:          1,
			pageSize:      5,
			expectedCount: 5,
			expectedTotal: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shares, total, err := shareService.ListMyShares(user.ID, tt.page, tt.pageSize)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedTotal, total)
			assert.Equal(t, tt.expectedCount, len(shares))

			// 验证排序（最新的在前）
			if len(shares) > 1 {
				for i := 0; i < len(shares)-1; i++ {
					assert.True(t, shares[i].CreatedAt.After(shares[i+1].CreatedAt) ||
						shares[i].CreatedAt.Equal(shares[i+1].CreatedAt))
				}
			}
		})
	}
}

// TestListMyShares_Empty 测试空分享列表
func TestListMyShares_Empty(t *testing.T) {
	shareService, _, user, _ := setupShareTestEnv(t)

	shares, total, err := shareService.ListMyShares(user.ID, 1, 10)

	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, shares)
}
