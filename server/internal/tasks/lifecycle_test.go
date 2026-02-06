// Package tasks 提供后台任务测试
//
// 本文件测试生命周期检查功能：
//   - 检查过期分享
//   - 检查下载次数上限
//   - 统计活跃分享
//
// 作者: AhaVault Team
// 创建时间: 2026-02-06
package tasks

import (
	"testing"
	"time"

	"ahavault/server/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLifecycleChecker_CheckExpiredShares(t *testing.T) {
	db := setupTestDB(t)
	lc := NewLifecycleChecker(db)

	// 创建测试用户
	user := &models.User{
		Email:    "lifecycle@test.com",
		Password: "hashed_password",
	}
	require.NoError(t, db.Create(user).Error)

	// 创建过期分享
	expiredShare := &models.ShareSession{
		ID:         uuid.New(),
		PickupCode: "EXPR1234",
		CreatorID:  user.ID,
		ExpiresAt:  time.Now().Add(-2 * time.Hour),
	}
	require.NoError(t, db.Create(expiredShare).Error)

	// 创建有效分享
	validShare := &models.ShareSession{
		ID:         uuid.New(),
		PickupCode: "VALD5678",
		CreatorID:  user.ID,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}
	require.NoError(t, db.Create(validShare).Error)

	// 执行生命周期检查
	result := lc.Run()

	// 验证结果
	assert.Equal(t, 1, result.ExpiredMarked)
	assert.Equal(t, 1, result.ActiveSharesCount)
	assert.Empty(t, result.Errors)

	// 验证过期分享已被标记停止
	var share models.ShareSession
	db.First(&share, "pickup_code = ?", "EXPR1234")
	assert.NotNil(t, share.StoppedAt)
}

func TestLifecycleChecker_CheckDownloadLimits(t *testing.T) {
	db := setupTestDB(t)
	lc := NewLifecycleChecker(db)

	// 创建测试用户
	user := &models.User{
		Email:    "limit@test.com",
		Password: "hashed_password",
	}
	require.NoError(t, db.Create(user).Error)

	// 创建达到下载上限的分享
	limitReachedShare := &models.ShareSession{
		ID:               uuid.New(),
		PickupCode:       "LIMIT123",
		CreatorID:        user.ID,
		MaxDownloads:     5,
		CurrentDownloads: 5,
		ExpiresAt:        time.Now().Add(24 * time.Hour),
	}
	require.NoError(t, db.Create(limitReachedShare).Error)

	// 创建未达到上限的分享
	underLimitShare := &models.ShareSession{
		ID:               uuid.New(),
		PickupCode:       "UNDER456",
		CreatorID:        user.ID,
		MaxDownloads:     10,
		CurrentDownloads: 3,
		ExpiresAt:        time.Now().Add(24 * time.Hour),
	}
	require.NoError(t, db.Create(underLimitShare).Error)

	// 创建无限制的分享
	unlimitedShare := &models.ShareSession{
		ID:               uuid.New(),
		PickupCode:       "UNLMT789",
		CreatorID:        user.ID,
		MaxDownloads:     0, // 0 表示无限制
		CurrentDownloads: 100,
		ExpiresAt:        time.Now().Add(24 * time.Hour),
	}
	require.NoError(t, db.Create(unlimitedShare).Error)

	// 执行生命周期检查
	result := lc.Run()

	// 验证结果
	assert.Equal(t, 1, result.DownloadLimitHit)
	assert.Equal(t, 2, result.ActiveSharesCount) // underLimit + unlimited
	assert.Empty(t, result.Errors)

	// 验证达到上限的分享已被标记停止
	var share models.ShareSession
	db.First(&share, "pickup_code = ?", "LIMIT123")
	assert.NotNil(t, share.StoppedAt)

	// 验证其他分享未被影响 - 重新查询
	var underLimitCheck models.ShareSession
	db.Where("pickup_code = ?", "UNDER456").First(&underLimitCheck)
	assert.Nil(t, underLimitCheck.StoppedAt)

	var unlimitedCheck models.ShareSession
	db.Where("pickup_code = ?", "UNLMT789").First(&unlimitedCheck)
	assert.Nil(t, unlimitedCheck.StoppedAt)
}

func TestLifecycleChecker_CountActiveShares(t *testing.T) {
	db := setupTestDB(t)
	lc := NewLifecycleChecker(db)

	// 创建测试用户
	user := &models.User{
		Email:    "count@test.com",
		Password: "hashed_password",
	}
	require.NoError(t, db.Create(user).Error)

	// 创建 3 个活跃分享
	for i := 0; i < 3; i++ {
		share := &models.ShareSession{
			ID:         uuid.New(),
			PickupCode: "ACTV" + string(rune('A'+i)) + "123",
			CreatorID:  user.ID,
			ExpiresAt:  time.Now().Add(time.Duration(i+1) * time.Hour),
		}
		require.NoError(t, db.Create(share).Error)
	}

	// 创建 1 个已停止的分享
	stoppedAt := time.Now()
	stoppedShare := &models.ShareSession{
		ID:         uuid.New(),
		PickupCode: "STOP1234",
		CreatorID:  user.ID,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		StoppedAt:  &stoppedAt,
	}
	require.NoError(t, db.Create(stoppedShare).Error)

	// 执行生命周期检查
	result := lc.Run()

	// 验证活跃分享数
	assert.Equal(t, 3, result.ActiveSharesCount)
	assert.Empty(t, result.Errors)
}
