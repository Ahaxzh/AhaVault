# ⚡ 快速参考卡

**目的**: 快速查找常用命令、路径、概念
**适用对象**: 开发者、Claude AI

---

## 🎯 新 Claude 会话快速启动

**复制粘贴给新会话**：
```
请先阅读以下文档了解项目状态：
1. docs/knowledge/INDEX.md
2. docs/memory-bank/progress.md

然后告诉我当前进度和下一步优先级任务。
```

详细指南: [guides/new-session-guide.md](../guides/new-session-guide.md)

---

## 📂 关键文件路径

### 核心文档
```bash
docs/knowledge/INDEX.md              # Claude 快速入口 ⭐
docs/memory-bank/progress.md        # 当前进度 ⭐
docs/memory-bank/PRD.md              # 产品需求
docs/tasks/README.md                 # 任务索引
Claude.md                            # 协作规范
```

### 配置文件
```bash
.env                                 # 本地环境变量
deploy/.env.example                  # 生产环境模板
docker-compose.yml                   # 开发环境 Docker
deploy/docker-compose.yml            # 生产环境 Docker
```

### 关键代码
```bash
server/internal/crypto/envelope.go   # 信封加密
server/internal/storage/local.go     # CAS 存储
server/internal/api/routes.go        # 路由定义
server/migrations/001_init.sql       # 数据库初始化
```

---

## ⚙️ 常用命令

### 后端开发
```bash
# 启动开发环境数据库
docker-compose up -d

# 启动后端服务
cd server
go run cmd/server/main.go

# 运行测试
go test -v ./...

# 测试覆盖率
go test -cover ./...

# 格式化代码
go fmt ./...

# 静态检查
go vet ./...
```

### 前端开发
```bash
cd web
npm install                    # 安装依赖
npm run dev                    # 开发服务器 (http://localhost:5173)
npm run build                  # 生产构建
npm run test                   # 单元测试
npm run test:coverage          # 覆盖率报告
npm run test:e2e               # E2E 测试
npm run test:e2e:ui            # E2E 交互式 UI
```

### Docker 管理
```bash
# 开发环境
docker-compose up -d           # 启动
docker-compose down            # 停止
docker-compose logs -f         # 查看日志
docker-compose ps              # 查看状态

# 生产环境
cd deploy
docker-compose up -d
docker-compose logs -f postgres
```

### Git 提交
```bash
# 原子提交模板
git add <files>
git commit -m "type(scope): subject

detailed description

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## 🔑 核心概念速查

### 信封加密 (Envelope Encryption)
```
文件 → 随机 DEK (256-bit) → AES-256-GCM 加密
DEK → KEK 加密 → 存储到数据库 encrypted_dek 字段

解密: KEK 解密 DEK → DEK 解密文件
```

### CAS 存储 (Content-Addressable Storage)
```
文件 → SHA-256 哈希 → aa/bb/aabbccdd...
            ↓
      file_blobs.hash (ref_count)
            ↓
      files_metadata (用户引用)
```

### 引用计数
```
用户上传/转存 → ref_count++
用户删除     → ref_count--
后台 GC      → 删除 ref_count=0 的文件
```

### 取件码
```
格式: 8 位字符
字符集: 2-9, A-Z (排除 O, I, 0, 1)
示例: A2B3C4D5
```

---

## 📊 测试覆盖率目标

| 模块 | 目标覆盖率 |
|------|-----------|
| crypto | ≥80% |
| storage | ≥80% |
| services | ≥70% |
| handlers | ≥60% |
| components (前端) | ≥70% |
| utils (前端) | ≥80% |

---

## 🚦 优先级定义

- **P0** 🔥: 紧急，阻塞发布，必须立即处理
- **P1** ⚠️: 重要，影响核心功能，本周内完成
- **P2** 💡: 一般，可延后，有时间再做

---

## 📝 提交类型 (Conventional Commits)

```
feat     - 新功能
fix      - Bug 修复
docs     - 文档更新
style    - 代码格式（不影响逻辑）
refactor - 重构（不修改功能）
test     - 测试相关
chore    - 构建/工具配置
```

---

## 🐛 故障排查快速指南

### 后端启动失败
```bash
# 检查数据库是否运行
docker-compose ps

# 检查环境变量
cat .env

# 查看详细日志
go run cmd/server/main.go
```

### 前端无法连接后端
```bash
# 检查 VITE_API_URL
cat web/.env

# 应该是: VITE_API_URL=http://localhost:8080/api
```

### Docker 容器无法启动
```bash
# 查看日志
docker-compose logs postgres
docker-compose logs redis

# 重建容器
docker-compose down
docker-compose up -d --force-recreate
```

### 测试失败
```bash
# 清理缓存
go clean -testcache

# 重新运行
go test -v ./...
```

---

## 🔗 快速链接

| 需求 | 文档 |
|------|------|
| 新会话启动 | [guides/new-session-guide.md](../guides/new-session-guide.md) |
| 当前进度 | [memory-bank/progress.md](../memory-bank/progress.md) |
| 任务列表 | [tasks/README.md](../tasks/README.md) |
| 协作规范 | [Claude.md](../../Claude.md) |
| API 文档 | [api/API.md](../api/API.md) |
| 架构设计 | [architecture/](../architecture/) |
| 技术栈 | [memory-bank/tech-stack.md](../memory-bank/tech-stack.md) |

---

## 🎯 工作流程速查

### 接到新任务
```
1. 查看任务文档 (docs/tasks/xxx.md)
2. 理解需求和验收标准
3. 查看相关代码
4. 开始编码 (遵循 Claude.md 规范)
5. 编写测试
6. 更新文档 (progress.md)
7. 提交代码
```

### 完成任务后
```
1. ✅ 更新 docs/memory-bank/progress.md
2. ✅ 更新对应任务文档 ([ ] → [x])
3. ✅ 运行测试并更新覆盖率
4. ✅ 提交文档更新
```

---

## 🔐 安全相关

### 环境变量
```bash
APP_MASTER_KEY    # 64 字符 HEX (32 字节)
JWT_SECRET        # Base64，足够长
POSTGRES_PASSWORD # 强密码
REDIS_PASSWORD    # 强密码
```

### 生成密钥
```bash
openssl rand -hex 32        # 生成 APP_MASTER_KEY
openssl rand -base64 64     # 生成 JWT_SECRET
openssl rand -base64 32     # 生成数据库密码
```

---

## 📈 性能指标

| 指标 | 目标值 |
|------|--------|
| API 响应时间 (P99) | ≤1s |
| 秒传检测 | ≤500ms |
| 文件上传速度 | ≥10MB/s |
| 并发用户 | ≥100 |

---

<div align="center">

**⚡ 常用操作一键可查！**

维护者：开发团队
最后更新：2026-02-04

</div>
