// Package tasks 提供后台任务调度器
//
// 本文件实现任务调度器：
//   - 统一管理所有后台任务
//   - 支持 cron 表达式定时执行
//   - 任务执行日志
//
// 作者: AhaVault Team
// 创建时间: 2026-02-06
package tasks

import (
	"log"
	"sync"

	"ahavault/server/internal/storage"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// Scheduler 任务调度器
type Scheduler struct {
	cron      *cron.Cron
	db        *gorm.DB
	storage   storage.Engine
	gc        *GarbageCollector
	lifecycle *LifecycleChecker
	running   bool
	mu        sync.Mutex
}

// NewScheduler 创建任务调度器
func NewScheduler(db *gorm.DB, storageEngine storage.Engine) *Scheduler {
	return &Scheduler{
		cron:      cron.New(),
		db:        db,
		storage:   storageEngine,
		gc:        NewGarbageCollector(db, storageEngine),
		lifecycle: NewLifecycleChecker(db),
	}
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	log.Println("[Scheduler] Initializing background task scheduler...")

	// 每天凌晨 2:00 执行垃圾回收
	_, err := s.cron.AddFunc("0 2 * * *", func() {
		log.Println("[Scheduler] Running scheduled garbage collection...")
		result := s.gc.Run()
		log.Printf("[Scheduler] GC completed: blobs=%d, shares=%d, soft_deleted=%d, space=%d bytes, errors=%d",
			result.OrphanBlobsDeleted,
			result.ExpiredSharesDeleted,
			result.SoftDeletedCleaned,
			result.SpaceReclaimed,
			len(result.Errors))
	})
	if err != nil {
		return err
	}

	// 每小时执行生命周期检查
	_, err = s.cron.AddFunc("@hourly", func() {
		log.Println("[Scheduler] Running scheduled lifecycle check...")
		result := s.lifecycle.Run()
		log.Printf("[Scheduler] Lifecycle check completed: expired=%d, limit_hit=%d, active=%d, errors=%d",
			result.ExpiredMarked,
			result.DownloadLimitHit,
			result.ActiveSharesCount,
			len(result.Errors))
	})
	if err != nil {
		return err
	}

	s.cron.Start()
	s.running = true
	log.Println("[Scheduler] Background task scheduler started")

	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	ctx := s.cron.Stop()
	<-ctx.Done()
	s.running = false
	log.Println("[Scheduler] Background task scheduler stopped")
}

// RunGCNow 立即执行垃圾回收（用于手动触发或测试）
func (s *Scheduler) RunGCNow() *GCResult {
	return s.gc.Run()
}

// RunLifecycleCheckNow 立即执行生命周期检查（用于手动触发或测试）
func (s *Scheduler) RunLifecycleCheckNow() *LifecycleResult {
	return s.lifecycle.Run()
}

// IsRunning 检查调度器是否运行中
func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}
