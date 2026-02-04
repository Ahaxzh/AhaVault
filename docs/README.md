# 📚 AhaVault 文档中心

**欢迎来到 AhaVault 文档中心！**

---

## 🚀 快速开始

### 我是新用户
- 📖 [产品介绍](./memory-bank/PRD.md) - 了解 AhaVault 是什么
- ⚡ [快速安装](./guides/development.md#快速开始) - 5 分钟搭建开发环境

### 我是开发者
- 🔍 **从这里开始**: [knowledge/INDEX.md](./knowledge/INDEX.md) ⭐
- 📊 [当前进度](./memory-bank/progress.md) - 了解项目状态
- 📋 [任务列表](./tasks/README.md) - 查看待办任务
- 🤖 [Claude 协作指南](../Claude.md) - 与 AI 协作规范

### 我是 Claude AI
- 🎯 **第一步**: 阅读 [knowledge/INDEX.md](./knowledge/INDEX.md)
- 🎯 **第二步**: 阅读 [memory-bank/progress.md](./memory-bank/progress.md)
- 🎯 **第三步**: 查看 [tasks/README.md](./tasks/README.md) 找到下一步任务

---

## 📂 文档结构

### 🧠 记忆库 (Memory Bank)
> Vibecode 风格的项目总览，Claude 快速了解项目的核心文档

- [progress.md](./memory-bank/progress.md) ⭐ **核心** - 当前进度与下一步计划
- [PRD.md](./memory-bank/PRD.md) - 产品需求文档
- [tech-stack.md](./memory-bank/tech-stack.md) - 技术栈选型
- [architecture.md](./memory-bank/architecture.md) - 架构总览
- [implementation-plan.md](./memory-bank/implementation-plan.md) - 实施计划
- [future-ideas.md](./memory-bank/future-ideas.md) - 未来想法

### 📋 任务管理 (Tasks)
> 细粒度的任务分解，按模块组织

- [README.md](./tasks/README.md) - 任务索引
- [backend/](./tasks/backend/) - 后端任务（6 个模块）
- [frontend/](./tasks/frontend/) - 前端任务（4 个模块）
- [testing/](./tasks/testing/) - 测试补充任务
- [templates/](./tasks/templates/) - 任务文档模板

### 📊 进度追踪 (Progress)
> 自动化生成的进度报告

- [coverage.md](./progress/coverage.md) - 测试覆盖率（脚本生成）
- [weekly/](./progress/weekly/) - 每周进度快照

### 🤔 架构决策 (Decisions)
> 重要技术决策的记录（ADR）

- [README.md](./decisions/README.md) - ADR 索引
- [000X-decision-title.md](./decisions/) - 具体决策记录

### 🏗️ 架构设计 (Architecture)
> 详细的技术设计文档

- [encryption.md](./architecture/encryption.md) - 信封加密设计
- [storage.md](./architecture/storage.md) - CAS 存储设计
- [security.md](./architecture/security.md) - 安全策略

### 🔌 API 文档 (API)
> RESTful API 接口规范

- [API.md](./api/API.md) - 完整 API 文档

### 📖 开发指南 (Guides)
> 实用的开发、测试、部署指南

- [development.md](./guides/development.md) - 本地开发环境搭建
- [deployment.md](./guides/deployment.md) - 生产环境部署（见 deploy/README.md）

### 🔍 知识库 (Knowledge)
> Claude AI 的快速检索索引

- [INDEX.md](./knowledge/INDEX.md) ⭐ **Claude 入口** - 知识总索引
- [quick-reference.md](./knowledge/quick-reference.md) - 快速参考卡
- [troubleshooting.md](./knowledge/troubleshooting.md) - 故障排查手册

---

## 🔄 文档维护

### 更新原则
- ✅ 每次完成任务必须更新文档
- ✅ 保持 `memory-bank/progress.md` 与 `tasks/` 同步
- ✅ 遵循 [Claude.md](../Claude.md#任务追踪与文档更新规范) 中的更新流程

### 更新检查清单
- [ ] 是否更新了 `memory-bank/progress.md`？
- [ ] 是否更新了对应的任务文档？
- [ ] 是否运行了测试并更新覆盖率？
- [ ] 是否在"更新日志"添加记录？

---

## 🎯 常用链接

| 需求 | 文档 |
|------|------|
| 了解项目进度 | [memory-bank/progress.md](./memory-bank/progress.md) |
| 查看下一步任务 | [tasks/README.md](./tasks/README.md) |
| 了解技术选型 | [memory-bank/tech-stack.md](./memory-bank/tech-stack.md) |
| 查看架构设计 | [architecture/](./architecture/) |
| API 接口查询 | [api/API.md](./api/API.md) |
| 本地开发环境 | [guides/development.md](./guides/development.md) |
| Claude 协作规范 | [../Claude.md](../Claude.md) |

---

## 📞 获取帮助

- 💬 提出问题：创建 GitHub Issue
- 📧 联系维护者：见项目 README
- 📚 查看常见问题：[guides/development.md](./guides/development.md)

---

<div align="center">

**🎯 保持文档更新，确保知识传承！**

维护者：开发团队
最后更新：2026-02-04

</div>
