// Package tasks 提供后台任务服务
//
// 本文件实现垃圾回收 (GC) 任务：
//   - 清理 ref_count = 0 的 file_blobs
//   - 删除对应的 CAS 物理文件
//   - 清理过期的 share_sessions
//   - 清理软删除超过 7 天的 files_metadata
//
// 作者: AhaVault Team
// 创建时间: 2026-02-06
package tasks

import (
	"log"
	"time"

	"ahavault/server/internal/models"
	"ahavault/server/internal/storage"
	"gorm.io/gorm"
)

// GarbageCollector 垃圾回收器
type GarbageCollector struct {
	db      *gorm.DB
	storage storage.Engine
}

// GCResult 垃圾回收结果
type GCResult struct {
	OrphanBlobsDeleted   int   // 删除的孤儿文件数
	ExpiredSharesDeleted int   // 删除的过期分享数
	SoftDeletedCleaned   int   // 清理的软删除文件数
	SpaceReclaimed       int64 // 释放的存储空间 (bytes)
	Duration             time.Duration
	Errors               []error
}

// NewGarbageCollector 创建垃圾回收器
func NewGarbageCollector(db *gorm.DB, storageEngine storage.Engine) *GarbageCollector {
	return &GarbageCollector{
		db:      db,
		storage: storageEngine,
	}
}

// Run 执行垃圾回收
//
// 执行顺序：
//  1. 清理软删除超过 7 天的 files_metadata（触发引用计数减少）
//  2. 清理 ref_count = 0 的 file_blobs 和物理文件
//  3. 清理过期的 share_sessions
func (gc *GarbageCollector) Run() *GCResult {
	startTime := time.Now()
	result := &GCResult{
		Errors: make([]error, 0),
	}

	log.Println("[GC] Starting garbage collection...")

	// 1. 清理软删除超过 7 天的 files_metadata
	softDeletedCount, err := gc.cleanSoftDeletedFiles()
	if err != nil {
		result.Errors = append(result.Errors, err)
		log.Printf("[GC] Error cleaning soft-deleted files: %v", err)
	} else {
		result.SoftDeletedCleaned = softDeletedCount
		log.Printf("[GC] Cleaned %d soft-deleted files", softDeletedCount)
	}

	// 2. 清理孤儿 blobs (ref_count = 0)
	orphanCount, spaceReclaimed, err := gc.cleanOrphanBlobs()
	if err != nil {
		result.Errors = append(result.Errors, err)
		log.Printf("[GC] Error cleaning orphan blobs: %v", err)
	} else {
		result.OrphanBlobsDeleted = orphanCount
		result.SpaceReclaimed = spaceReclaimed
		log.Printf("[GC] Cleaned %d orphan blobs, reclaimed %d bytes", orphanCount, spaceReclaimed)
	}

	// 3. 清理过期的 share_sessions
	expiredCount, err := gc.cleanExpiredShares()
	if err != nil {
		result.Errors = append(result.Errors, err)
		log.Printf("[GC] Error cleaning expired shares: %v", err)
	} else {
		result.ExpiredSharesDeleted = expiredCount
		log.Printf("[GC] Cleaned %d expired shares", expiredCount)
	}

	result.Duration = time.Since(startTime)
	log.Printf("[GC] Garbage collection completed in %v", result.Duration)

	return result
}

// cleanSoftDeletedFiles 清理软删除超过 7 天的文件
func (gc *GarbageCollector) cleanSoftDeletedFiles() (int, error) {
	threshold := time.Now().AddDate(0, 0, -7) // 7 天前

	// 查找需要清理的文件
	var files []models.FileMetadata
	err := gc.db.Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at < ?", threshold).
		Find(&files).Error
	if err != nil {
		return 0, err
	}

	if len(files) == 0 {
		return 0, nil
	}

	count := 0
	for _, file := range files {
		// 使用事务处理引用计数和删除
		err := gc.db.Transaction(func(tx *gorm.DB) error {
			// 减少引用计数
			err := tx.Model(&models.FileBlob{}).
				Where("hash = ?", file.FileBlobHash).
				Update("ref_count", gorm.Expr("ref_count - 1")).Error
			if err != nil {
				return err
			}

			// 永久删除元数据
			return tx.Unscoped().Delete(&file).Error
		})

		if err != nil {
			log.Printf("[GC] Failed to clean file %s: %v", file.ID, err)
			continue
		}
		count++
	}

	return count, nil
}

// cleanOrphanBlobs 清理孤儿 blobs (ref_count = 0)
func (gc *GarbageCollector) cleanOrphanBlobs() (int, int64, error) {
	var blobs []models.FileBlob
	err := gc.db.Where("ref_count <= 0").Find(&blobs).Error
	if err != nil {
		return 0, 0, err
	}

	if len(blobs) == 0 {
		return 0, 0, nil
	}

	count := 0
	var spaceReclaimed int64

	for _, blob := range blobs {
		// 删除物理文件
		err := gc.storage.Delete(blob.Hash)
		if err != nil {
			log.Printf("[GC] Failed to delete physical file %s: %v", blob.Hash, err)
			// 继续处理其他文件
		}

		// 删除数据库记录
		err = gc.db.Delete(&blob).Error
		if err != nil {
			log.Printf("[GC] Failed to delete blob record %s: %v", blob.Hash, err)
			continue
		}

		count++
		spaceReclaimed += blob.Size
	}

	return count, spaceReclaimed, nil
}

// cleanExpiredShares 清理过期的分享
func (gc *GarbageCollector) cleanExpiredShares() (int, error) {
	now := time.Now()

	// 查找过期的分享
	var shares []models.ShareSession
	err := gc.db.Where("expires_at < ? AND stopped_at IS NULL", now).Find(&shares).Error
	if err != nil {
		return 0, err
	}

	if len(shares) == 0 {
		return 0, nil
	}

	count := 0
	for _, share := range shares {
		// 标记为已停止
		err := gc.db.Model(&share).Update("stopped_at", now).Error
		if err != nil {
			log.Printf("[GC] Failed to stop expired share %s: %v", share.ID, err)
			continue
		}
		count++
	}

	return count, nil
}
