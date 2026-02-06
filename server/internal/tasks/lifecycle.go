// Package tasks 提供后台任务服务
//
// 本文件实现分享生命周期检查任务：
//   - 定期检查过期的 share_sessions
//   - 检查下载次数是否达到上限
//   - 记录检查日志
//
// 作者: AhaVault Team
// 创建时间: 2026-02-06
package tasks

import (
	"log"
	"time"

	"ahavault/server/internal/models"
	"gorm.io/gorm"
)

// LifecycleChecker 生命周期检查器
type LifecycleChecker struct {
	db *gorm.DB
}

// LifecycleResult 生命周期检查结果
type LifecycleResult struct {
	ExpiredMarked      int // 标记为过期的分享数
	DownloadLimitHit   int // 达到下载上限的分享数
	ActiveSharesCount  int // 当前活跃分享数
	Duration           time.Duration
	Errors             []error
}

// NewLifecycleChecker 创建生命周期检查器
func NewLifecycleChecker(db *gorm.DB) *LifecycleChecker {
	return &LifecycleChecker{
		db: db,
	}
}

// Run 执行生命周期检查
//
// 执行内容：
//  1. 检查并标记过期的分享
//  2. 检查并标记达到下载上限的分享
//  3. 统计当前活跃分享数
func (lc *LifecycleChecker) Run() *LifecycleResult {
	startTime := time.Now()
	result := &LifecycleResult{
		Errors: make([]error, 0),
	}

	log.Println("[Lifecycle] Starting lifecycle check...")

	// 1. 检查过期的分享
	expiredCount, err := lc.checkExpiredShares()
	if err != nil {
		result.Errors = append(result.Errors, err)
		log.Printf("[Lifecycle] Error checking expired shares: %v", err)
	} else {
		result.ExpiredMarked = expiredCount
		if expiredCount > 0 {
			log.Printf("[Lifecycle] Marked %d shares as expired", expiredCount)
		}
	}

	// 2. 检查下载次数达到上限的分享
	limitHitCount, err := lc.checkDownloadLimits()
	if err != nil {
		result.Errors = append(result.Errors, err)
		log.Printf("[Lifecycle] Error checking download limits: %v", err)
	} else {
		result.DownloadLimitHit = limitHitCount
		if limitHitCount > 0 {
			log.Printf("[Lifecycle] Marked %d shares as download limit reached", limitHitCount)
		}
	}

	// 3. 统计活跃分享数
	activeCount, err := lc.countActiveShares()
	if err != nil {
		result.Errors = append(result.Errors, err)
		log.Printf("[Lifecycle] Error counting active shares: %v", err)
	} else {
		result.ActiveSharesCount = activeCount
	}

	result.Duration = time.Since(startTime)
	log.Printf("[Lifecycle] Lifecycle check completed in %v (active shares: %d)",
		result.Duration, result.ActiveSharesCount)

	return result
}

// checkExpiredShares 检查并标记过期的分享
func (lc *LifecycleChecker) checkExpiredShares() (int, error) {
	now := time.Now()

	// 更新过期但未停止的分享
	result := lc.db.Model(&models.ShareSession{}).
		Where("expires_at < ? AND stopped_at IS NULL", now).
		Update("stopped_at", now)

	if result.Error != nil {
		return 0, result.Error
	}

	return int(result.RowsAffected), nil
}

// checkDownloadLimits 检查下载次数达到上限的分享
func (lc *LifecycleChecker) checkDownloadLimits() (int, error) {
	now := time.Now()

	// 查找下载次数达到上限且未停止的分享
	// max_downloads > 0 表示有限制，current_downloads >= max_downloads 表示达到上限
	result := lc.db.Model(&models.ShareSession{}).
		Where("max_downloads > 0 AND current_downloads >= max_downloads AND stopped_at IS NULL").
		Update("stopped_at", now)

	if result.Error != nil {
		return 0, result.Error
	}

	return int(result.RowsAffected), nil
}

// countActiveShares 统计当前活跃分享数
func (lc *LifecycleChecker) countActiveShares() (int, error) {
	var count int64
	err := lc.db.Model(&models.ShareSession{}).
		Where("stopped_at IS NULL AND expires_at > ?", time.Now()).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return int(count), nil
}
