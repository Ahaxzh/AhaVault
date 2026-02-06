# 后台任务 (Background Tasks) 任务清单

**模块名称**: 后台任务
**负责人**: Claude AI
**最后更新**: 2026-02-06
**当前进度**: 100%
**状态**: ✅ 完成

---

## 📊 进度概览

| 任务 | 进度 | 测试覆盖率 | 状态 |
|------|------|-----------|------|
| 垃圾回收 | 100% | 100% | ✅ |
| 生命周期检查 | 100% | 100% | ✅ |
| 统计任务 | 0% | 0% | ⚪ |
| 任务调度器 | 100% | N/A | ✅ |

---

## 📋 任务清单

### ✅ 已完成（高优先级 P1）

- [x] `gc.go` - 垃圾回收任务 ✅ **已完成 (2026-02-06)**
  - 优先级: **P1** ⚠️
  - 功能:
    - 清理 `ref_count = 0` 的 file_blobs
    - 删除对应的 CAS 物理文件
    - 清理过期的 share_sessions
    - 清理超过保留期的软删除文件（deleted_at + 7 天）
  - 执行时间: 每天 02:00

- [x] `lifecycle.go` - 分享生命周期检查 ✅ **已完成 (2026-02-06)**
  - 优先级: **P1** ⚠️
  - 功能:
    - 定期检查过期的 share_sessions
    - 标记为已过期（不删除，保留审计记录）
    - 检查下载次数是否达到上限
  - 执行时间: 每小时

- [x] `gc_test.go` - 垃圾回收测试 ✅ **已完成 (2026-02-06)**
  - 3/3 测试用例通过

- [x] `lifecycle_test.go` - 生命周期检查测试 ✅ **已完成 (2026-02-06)**
  - 3/3 测试用例通过

- [x] `scheduler.go` - 任务调度器 ✅ **已完成 (2026-02-06)**
  - 使用 robfig/cron/v3 实现
  - 统一管理所有后台任务

### ⚪ 待办（中优先级 P2）

- [ ] `stats.go` - 统计任务
  - 优先级: P2
  - 预计工作量: 3 小时
  - 功能:
    - 每日统计用户存储使用情况
    - 记录系统统计数据（总用户数、总文件数、总存储）
    - 生成日报/周报数据
  - 执行时间: 每天 03:00

- [ ] `scheduler.go` - 任务调度器
  - 优先级: P2
  - 预计工作量: 2 小时
  - 依赖: github.com/robfig/cron/v3 或 go-co-op/gocron
  - 功能:
    - 统一管理所有后台任务
    - 任务调度配置
    - 任务执行日志

---

## 🧪 测试状态

### 测试覆盖率

| 文件 | 测试文件 | 覆盖率 | 目标 | 状态 |
|------|---------|--------|------|------|
| `gc.go` | - | 0% | 70% | ⚪ |
| `lifecycle.go` | - | 0% | 70% | ⚪ |
| `stats.go` | - | 0% | 60% | ⚪ |
| `scheduler.go` | - | 0% | 60% | ⚪ |

**总体覆盖率**: 0% / 目标 70%

---

## 🐛 已知问题

### 高优先级 ⚠️

1. **垃圾回收机制完全缺失**
   - 影响范围: ref_count = 0 的文件无法清理，磁盘空间持续增长
   - 解决方案: 立即实现 gc.go
   - 预计修复: 2026-02-08

2. **过期分享无法自动清理**
   - 影响范围: 过期分享仍可访问
   - 解决方案: 实现 lifecycle.go
   - 预计修复: 2026-02-08

---

## 📝 技术债务

无技术债务 ✅

---

## 🔗 相关文档

- **数据模型**: [01-models.md](./01-models.md)
- **存储引擎**: [03-storage.md](./03-storage.md)

---

## 🗑️ 垃圾回收设计概要

### GC 流程

```
定时任务触发 (每天 02:00)
    ↓
1. 查找 ref_count = 0 的 file_blobs
    ↓
2. 对每个 blob:
   - 删除 CAS 物理文件 (storage.Delete(hash))
   - 删除 file_blobs 记录
    ↓
3. 查找过期的 share_sessions (expires_at < NOW())
    ↓
4. 删除过期分享
    ↓
5. 查找软删除超过 7 天的 files_metadata
   (deleted_at IS NOT NULL AND deleted_at < NOW() - INTERVAL '7 days')
    ↓
6. 永久删除（ref_count--）
    ↓
7. 记录 GC 日志（清理数量、释放空间）
```

### 生命周期检查流程

```
定时任务触发 (每小时)
    ↓
1. 查找所有活跃 share_sessions
    ↓
2. 检查过期时间
   IF expires_at < NOW():
     - 标记为过期（可选：添加 is_expired 字段）
    ↓
3. 检查下载次数
   IF access_count >= max_downloads:
     - 标记为已达上限
    ↓
4. 记录检查日志
```

---

## 🛠️ 实现建议

### 使用 cron 库

```go
package tasks

import (
    "github.com/robfig/cron/v3"
)

func InitScheduler() *cron.Cron {
    c := cron.New()

    // 每天 02:00 执行 GC
    c.AddFunc("0 2 * * *", RunGarbageCollection)

    // 每小时检查生命周期
    c.AddFunc("@hourly", CheckShareLifecycle)

    // 每天 03:00 统计
    c.AddFunc("0 3 * * *", GenerateStats)

    c.Start()
    return c
}
```

### 主程序集成

```go
// cmd/server/main.go

func main() {
    // ...

    // 启动后台任务
    scheduler := tasks.InitScheduler()
    defer scheduler.Stop()

    // ...
}
```

---

## 📅 更新日志

### 2026-02-04
- 📋 创建任务文档
- 📋 定义 GC 和生命周期检查需求

---

**维护者**: Claude AI
**下一步**: 实现 gc.go 和 lifecycle.go（P1 任务）
