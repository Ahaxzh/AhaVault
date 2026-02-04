# 本地开发环境指南

**版本**: v1.0
**最后更新**: 2026-02-04
**负责人**: Claude AI

---

## 📋 目录

- [1. 环境要求](#1-环境要求)
- [2. 快速开始](#2-快速开始)
- [3. 前端开发](#3-前端开发)
- [4. 后端开发](#4-后端开发)
- [5. 数据库管理](#5-数据库管理)
- [6. 调试技巧](#6-调试技巧)
- [7. Claude Code 会话管理](#7-claude-code-会话管理)
- [8. 常见问题](#8-常见问题)

---

## 1. 环境要求

### 1.1 必需软件

| 软件 | 版本要求 | 说明 |
|------|---------|------|
| **Docker** | 20.10+ | 用于运行 PostgreSQL 和 Redis |
| **Docker Compose** | 2.0+ | 容器编排 |
| **Go** | 1.21+ | 后端开发 |
| **Node.js** | 20+ | 前端开发 |
| **Git** | 2.0+ | 版本控制 |

### 1.2 推荐工具

| 工具 | 用途 |
|------|------|
| **VS Code** | 主要 IDE（推荐安装 Go、React、ESLint 插件） |
| **Postman** / **Insomnia** | API 测试 |
| **DBeaver** / **pgAdmin** | PostgreSQL 数据库管理 |
| **RedisInsight** | Redis 可视化工具 |

### 1.3 系统要求

- **操作系统**: macOS / Linux / Windows (WSL2)
- **内存**: 至少 8GB（推荐 16GB）
- **磁盘**: 至少 10GB 可用空间

---

## 2. 快速开始

### 2.1 克隆项目

```bash
# 克隆仓库
git clone https://github.com/yourusername/AhaVault.git
cd AhaVault

# 查看项目结构
tree -L 2 -I 'node_modules|.git'
```

### 2.2 配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env

# 生成 Master Key (KEK)
openssl rand -hex 32

# 编辑 .env 文件
vim .env
```

**必填配置项**:
```bash
# 核心配置（必填）
APP_MASTER_KEY=<刚才生成的 64 字符 HEX 字符串>
APP_INVITE_CODE=AHAVAULT2026

# 数据库配置（使用 Docker 默认值即可）
POSTGRES_PASSWORD=ahavault_dev_2026
REDIS_PASSWORD=redis_dev_2026

# 存储配置（本地开发使用 local）
STORAGE_TYPE=local
STORAGE_PATH=/data/storage

# 开发模式
GIN_MODE=debug
LOG_LEVEL=debug
```

### 2.3 启动基础服务（Docker）

```bash
# 启动 PostgreSQL 和 Redis
docker-compose up -d postgres redis

# 验证服务状态
docker-compose ps

# 查看日志
docker-compose logs -f postgres redis
```

**预期输出**:
```
NAME                    STATUS        PORTS
ahavault-postgres       Up            0.0.0.0:5432->5432/tcp
ahavault-redis          Up            0.0.0.0:6379->6379/tcp
```

### 2.4 初始化数据库

```bash
# 连接到 PostgreSQL
docker exec -it ahavault-postgres psql -U ahavault -d ahavault

# 执行初始化 SQL（在 psql 中）
\i /docker-entrypoint-initdb.d/001_init.sql

# 验证表是否创建成功
\dt

# 退出
\q
```

---

## 3. 前端开发

### 3.1 安装依赖

```bash
cd web

# 使用 npm
npm install

# 或使用 pnpm（推荐，速度更快）
pnpm install
```

### 3.2 启动开发服务器

```bash
# 启动 Vite 开发服务器
npm run dev

# 或指定端口
npm run dev -- --port 5173
```

**预期输出**:
```
  VITE v5.0.0  ready in 500 ms

  ➜  Local:   http://localhost:5173/
  ➜  Network: http://192.168.1.100:5173/
  ➜  press h to show help
```

### 3.3 前端项目结构

```
web/
├── src/
│   ├── components/       # React 组件
│   │   ├── common/       # 通用组件
│   │   ├── upload/       # 上传相关
│   │   └── share/        # 分享相关
│   ├── pages/            # 页面组件
│   ├── services/         # API 服务层
│   ├── hooks/            # 自定义 Hooks
│   ├── utils/            # 工具函数
│   ├── workers/          # Web Workers
│   ├── types/            # TypeScript 类型
│   ├── assets/           # 静态资源
│   ├── styles/           # 全局样式
│   ├── App.tsx           # 应用入口
│   └── main.tsx          # Vite 入口
├── public/               # 公共资源
├── index.html
├── vite.config.ts        # Vite 配置
├── tsconfig.json         # TypeScript 配置
├── tailwind.config.js    # TailwindCSS 配置
└── package.json
```

### 3.4 前端开发命令

```bash
# 启动开发服务器
npm run dev

# 类型检查
npm run type-check

# 代码格式化
npm run format

# 代码检查
npm run lint

# 修复 ESLint 错误
npm run lint:fix

# 构建生产版本
npm run build

# 预览构建产物
npm run preview
```

### 3.5 环境变量配置

创建 `web/.env.local`:
```bash
# 后端 API 地址
VITE_API_BASE_URL=http://localhost:8080/api

# Turnstile Site Key（测试环境）
VITE_TURNSTILE_SITE_KEY=1x00000000000000000000AA

# 调试模式
VITE_DEBUG=true
```

### 3.6 Hot Module Replacement (HMR)

Vite 自带 HMR，修改代码后会自动刷新：

```tsx
// src/App.tsx
import { useState } from 'react';

function App() {
  const [count, setCount] = useState(0);

  return (
    <div>
      <h1>Count: {count}</h1>
      <button onClick={() => setCount(c => c + 1)}>
        Increment
      </button>
    </div>
  );
}

export default App;

// 保存文件后，浏览器会立即更新，状态保持不变
```

---

## 4. 后端开发

### 4.1 安装依赖

```bash
cd server

# 下载 Go 模块
go mod download

# 验证依赖
go mod verify
```

### 4.2 初始化 Go 模块（首次）

```bash
# 如果 go.mod 不存在，初始化
go mod init github.com/yourusername/AhaVault

# 安装核心依赖
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/go-redis/redis/v8
go get github.com/aws/aws-sdk-go
```

### 4.3 启动后端服务

```bash
cd server

# 方式 1: 直接运行
go run cmd/server/main.go

# 方式 2: 使用 air 实现热重载（推荐）
# 安装 air
go install github.com/cosmtrek/air@latest

# 启动 air
air

# 方式 3: 构建后运行
go build -o ahavault cmd/server/main.go
./ahavault
```

**预期输出**:
```
[GIN-debug] [WARNING] Running in "debug" mode.
[GIN-debug] GET    /api/health           --> main.healthCheck (3 handlers)
[GIN-debug] POST   /api/auth/login       --> handlers.Login (4 handlers)
[GIN-debug] Listening and serving HTTP on :8080
```

### 4.4 后端项目结构

```
server/
├── cmd/
│   └── server/
│       └── main.go           # 主程序入口
├── internal/
│   ├── api/
│   │   ├── routes.go         # 路由定义
│   │   └── handlers/         # HTTP 处理器
│   ├── models/               # 数据模型
│   ├── services/             # 业务逻辑
│   ├── storage/              # 存储引擎
│   ├── crypto/               # 加密模块
│   ├── middleware/           # 中间件
│   ├── config/               # 配置管理
│   └── tasks/                # 后台任务
├── pkg/                      # 公共库
├── migrations/               # 数据库迁移
│   └── 001_init.sql
├── go.mod
└── go.sum
```

### 4.5 后端开发命令

```bash
# 运行服务
go run cmd/server/main.go

# 运行测试
go test ./...

# 运行测试（详细输出）
go test -v ./...

# 运行测试（覆盖率）
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 代码格式化
go fmt ./...

# 代码检查
golangci-lint run

# 构建
go build -o ahavault cmd/server/main.go

# 构建（启用优化）
CGO_ENABLED=0 go build -ldflags="-s -w" -o ahavault cmd/server/main.go
```

### 4.6 Air 配置（热重载）

创建 `server/.air.toml`:
```toml
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/main cmd/server/main.go"
  bin = "tmp/main"
  full_bin = "APP_ENV=dev ./tmp/main"
  include_ext = ["go", "tpl", "tmpl", "html"]
  exclude_dir = ["assets", "tmp", "vendor"]
  include_dir = []
  exclude_file = []
  delay = 1000
  stop_on_error = true
  log = "air.log"

[color]
  main = "magenta"
  watcher = "cyan"
  build = "yellow"
  runner = "green"

[misc]
  clean_on_exit = true
```

---

## 5. 数据库管理

### 5.1 连接 PostgreSQL

**方式 1: psql 命令行**
```bash
# 通过 Docker
docker exec -it ahavault-postgres psql -U ahavault -d ahavault

# 本地 psql（如果安装了）
psql -h localhost -p 5432 -U ahavault -d ahavault
```

**方式 2: DBeaver / pgAdmin**
```
Host: localhost
Port: 5432
Database: ahavault
Username: ahavault
Password: <POSTGRES_PASSWORD>
```

### 5.2 常用 SQL 命令

```sql
-- 查看所有表
\dt

-- 查看表结构
\d users
\d files_metadata
\d file_blobs

-- 查询用户数量
SELECT COUNT(*) FROM users;

-- 查询文件统计
SELECT
    COUNT(*) as total_files,
    SUM(size) as total_size,
    AVG(ref_count) as avg_ref_count
FROM file_blobs;

-- 查看引用计数分布
SELECT ref_count, COUNT(*) as count
FROM file_blobs
GROUP BY ref_count
ORDER BY ref_count;

-- 清空测试数据（慎用！）
TRUNCATE TABLE files_metadata CASCADE;
TRUNCATE TABLE file_blobs CASCADE;
TRUNCATE TABLE users CASCADE;
```

### 5.3 连接 Redis

**方式 1: redis-cli**
```bash
# 通过 Docker
docker exec -it ahavault-redis redis-cli -a <REDIS_PASSWORD>

# 本地 redis-cli
redis-cli -h localhost -p 6379 -a <REDIS_PASSWORD>
```

**常用 Redis 命令**:
```bash
# 查看所有 key
KEYS *

# 查看 IP 限流记录
KEYS ratelimit:*

# 查看 Session
KEYS session:*

# 查看 key 的值
GET session:550e8400-e29b-41d4-a716-446655440000

# 删除 key
DEL ratelimit:192.168.1.100

# 清空所有数据（慎用！）
FLUSHALL
```

### 5.4 数据库迁移

创建迁移文件 `server/migrations/001_init.sql`:
```sql
-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user',
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 创建物理文件表
CREATE TABLE IF NOT EXISTS file_blobs (
    hash VARCHAR(64) PRIMARY KEY,
    store_path VARCHAR(255) NOT NULL,
    encrypted_dek TEXT NOT NULL,
    size BIGINT NOT NULL,
    mime_type VARCHAR(128),
    ref_count INT NOT NULL DEFAULT 1,
    is_banned BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 创建逻辑文件表
CREATE TABLE IF NOT EXISTS files_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_blob_hash VARCHAR(64) NOT NULL REFERENCES file_blobs(hash) ON DELETE RESTRICT,
    filename VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建分享会话表
CREATE TABLE IF NOT EXISTS share_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pickup_code VARCHAR(8) UNIQUE NOT NULL,
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    max_downloads INT DEFAULT 0,
    current_downloads INT DEFAULT 0,
    password_hash VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL
);

-- 创建分享-文件关联表
CREATE TABLE IF NOT EXISTS share_files (
    share_id UUID NOT NULL REFERENCES share_sessions(id) ON DELETE CASCADE,
    file_id UUID NOT NULL REFERENCES files_metadata(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (share_id, file_id)
);

-- 创建索引
CREATE INDEX idx_user_files ON files_metadata(user_id, deleted_at);
CREATE INDEX idx_blob_hash ON files_metadata(file_blob_hash);
CREATE INDEX idx_ref_count ON file_blobs(ref_count);
CREATE INDEX idx_pickup_code ON share_sessions(pickup_code);
CREATE INDEX idx_deleted_at ON files_metadata(deleted_at) WHERE deleted_at IS NOT NULL;
```

**执行迁移**:
```bash
# 方式 1: 通过 psql 执行
docker exec -i ahavault-postgres psql -U ahavault -d ahavault < server/migrations/001_init.sql

# 方式 2: 在 psql 中执行
\i /path/to/001_init.sql
```

---

## 6. 调试技巧

### 6.1 前端调试

#### 6.1.1 浏览器 DevTools

```javascript
// 在代码中添加断点
debugger;

// 使用 console
console.log('User data:', userData);
console.table(fileList);
console.error('Upload failed:', error);
console.trace('Call stack');
```

#### 6.1.2 React DevTools

安装 Chrome 插件：[React Developer Tools](https://chrome.google.com/webstore/detail/react-developer-tools/fmkadmapgofadopljbjfkapdkoienihi)

**功能**:
- 查看组件树
- 检查 Props 和 State
- 性能分析

#### 6.1.3 网络请求调试

```typescript
// src/services/api.ts
import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
});

// 请求拦截器（添加日志）
api.interceptors.request.use(
  (config) => {
    console.log('[API Request]', config.method?.toUpperCase(), config.url);
    return config;
  },
  (error) => {
    console.error('[API Request Error]', error);
    return Promise.reject(error);
  }
);

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    console.log('[API Response]', response.status, response.config.url);
    return response;
  },
  (error) => {
    console.error('[API Response Error]', error.response?.status, error.response?.data);
    return Promise.reject(error);
  }
);
```

### 6.2 后端调试

#### 6.2.1 日志输出

```go
// server/internal/api/handlers/file.go
import "github.com/sirupsen/logrus"

func (h *FileHandler) Upload(c *gin.Context) {
    log := logrus.WithFields(logrus.Fields{
        "user_id": c.GetString("user_id"),
        "ip":      c.ClientIP(),
    })

    log.Info("Starting file upload")

    // 业务逻辑...

    log.WithField("hash", hash).Info("File upload completed")
}
```

#### 6.2.2 使用 Delve 调试器

```bash
# 安装 Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 启动调试
dlv debug cmd/server/main.go

# 在 Delve 中设置断点
(dlv) break handlers.Upload
(dlv) continue

# 查看变量
(dlv) print hash
(dlv) print user
```

#### 6.2.3 VS Code 调试配置

创建 `.vscode/launch.json`:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Server",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/server/cmd/server",
      "env": {
        "APP_MASTER_KEY": "your-key-here",
        "GIN_MODE": "debug"
      },
      "args": []
    }
  ]
}
```

### 6.3 API 测试

#### 6.3.1 使用 curl

```bash
# 健康检查
curl http://localhost:8080/api/health

# 用户注册
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123!",
    "invite_code": "AHAVAULT2026"
  }'

# 用户登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123!"
  }'

# 获取文件列表（需要 Token）
curl http://localhost:8080/api/files \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

#### 6.3.2 Postman 集合

导入以下 JSON 到 Postman:
```json
{
  "info": {
    "name": "AhaVault API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Auth",
      "item": [
        {
          "name": "Register",
          "request": {
            "method": "POST",
            "url": "{{base_url}}/auth/register",
            "body": {
              "mode": "raw",
              "raw": "{\n  \"email\": \"test@example.com\",\n  \"password\": \"Test123!\"\n}"
            }
          }
        }
      ]
    }
  ],
  "variable": [
    {
      "key": "base_url",
      "value": "http://localhost:8080/api"
    }
  ]
}
```

---

## 7. Claude Code 会话管理

### 7.1 会话恢复

**继续最近的会话**（推荐）:
```bash
# 继续当前目录下最近的会话
claude code --continue

# 或简写
claude code -c
```

**通过交互式选择器恢复**:
```bash
# 打开会话选择器
claude code --resume

# 或简写
claude code -r

# 通过搜索关键词筛选会话
claude code --resume "AhaVault"
```

**恢复特定会话 ID**:
```bash
# 如果知道会话 ID
claude code --resume <session-id>
```

### 7.2 更新 Claude Code 后恢复会话

```bash
# 1. 退出当前会话
exit

# 2. 更新 Claude Code
npm update -g @anthropic-ai/claude-code
# 或
brew upgrade claude-code

# 3. 继续最近的会话（最简单）
claude code --continue

# 或通过交互式选择器恢复
claude code --resume
```

### 7.3 会话数据存储

**存储位置**:
```bash
# macOS / Linux
~/.config/claude-code/sessions/

# Windows
%APPDATA%/claude-code/sessions/
```

**备份会话**:
```bash
# 创建备份
tar -czf claude-sessions-backup-$(date +%Y%m%d).tar.gz \
    ~/.config/claude-code/sessions/

# 恢复备份
tar -xzf claude-sessions-backup-20260204.tar.gz -C ~/
```

### 7.4 会话最佳实践

1. **为重要会话命名**:
   ```bash
   claude code --title "AhaVault 开发 - Phase 2"
   ```

2. **定期检查会话状态**:
   ```bash
   /tasks         # 查看任务列表
   /help          # 查看可用命令
   ```

3. **长期项目使用同一会话**:
   - 避免频繁创建新会话
   - 利用上下文累积提高效率

4. **会话清理**:
   ```bash
   # 删除旧会话（释放空间）
   claude code --clean-sessions --older-than 30d
   ```

---

## 8. 常见问题

### 7.1 Docker 相关

**Q: Docker 容器无法启动**

```bash
# 查看容器日志
docker-compose logs postgres

# 常见原因：端口占用
lsof -i :5432
# 解决：停止占用端口的进程或修改 docker-compose.yml 端口映射
```

**Q: 数据库连接失败**

```bash
# 检查容器是否运行
docker-compose ps

# 检查网络连接
docker exec ahavault-postgres pg_isready -U ahavault

# 重启容器
docker-compose restart postgres
```

### 7.2 前端相关

**Q: npm install 失败**

```bash
# 清理缓存
npm cache clean --force
rm -rf node_modules package-lock.json

# 重新安装
npm install

# 或使用 pnpm
pnpm install
```

**Q: 端口 5173 被占用**

```bash
# 查找占用进程
lsof -i :5173

# 杀死进程
kill -9 <PID>

# 或使用其他端口
npm run dev -- --port 3000
```

### 7.3 后端相关

**Q: go mod download 失败**

```bash
# 使用代理（中国大陆）
export GOPROXY=https://goproxy.cn,direct

# 重新下载
go mod download
```

**Q: 环境变量未加载**

```bash
# 方式 1: 使用 source
source .env && go run cmd/server/main.go

# 方式 2: 使用 godotenv
go get github.com/joho/godotenv

# 在 main.go 中加载
import "github.com/joho/godotenv"

func main() {
    godotenv.Load()
    // ...
}
```

**Q: 编译错误 "undefined: xxx"**

```bash
# 更新依赖
go mod tidy

# 重新下载
go mod download

# 验证
go mod verify
```

### 7.4 数据库相关

**Q: PostgreSQL 连接被拒绝**

```bash
# 检查容器状态
docker-compose ps

# 检查网络
docker inspect ahavault-postgres | grep IPAddress

# 使用正确的连接字符串
postgres://ahavault:password@localhost:5432/ahavault?sslmode=disable
```

**Q: 表不存在**

```sql
-- 检查当前数据库
SELECT current_database();

-- 检查所有表
\dt

-- 执行迁移
\i /path/to/001_init.sql
```

### 8.1 Claude Code 相关

**Q: 更新 Claude Code 后会话丢失？**

会话不会丢失！使用以下命令恢复：
```bash
claude code --continue        # 继续最近的会话
claude code --resume          # 打开交互式选择器
claude code --resume <id>     # 恢复特定会话
```

**Q: 如何查看可用的命令？**

```bash
# 查看所有命令和选项
claude code --help

# 在会话中查看斜杠命令
/help
```

**Q: 会话数据占用空间太大？**

清理旧会话：
```bash
# 删除 30 天前的会话
claude code --clean-sessions --older-than 30d

# 手动删除
rm -rf ~/.config/claude-code/sessions/<session-id>
```

### 8.2 性能问题

**Q: 前端加载缓慢**

```bash
# 检查网络请求
# 在 Chrome DevTools Network 面板查看瀑布图

# 启用 Vite 预构建缓存
rm -rf node_modules/.vite

# 优化依赖
npm run build -- --profile
```

**Q: 后端响应慢**

```go
// 添加性能日志
import "time"

func (h *Handler) Upload(c *gin.Context) {
    start := time.Now()
    defer func() {
        log.Infof("Upload took %v", time.Since(start))
    }()

    // 业务逻辑...
}
```

---

## 📚 参考资料

- [Vite 官方文档](https://vitejs.dev/)
- [React 官方文档](https://react.dev/)
- [Go 官方文档](https://go.dev/doc/)
- [Gin 框架文档](https://gin-gonic.com/docs/)
- [PostgreSQL 文档](https://www.postgresql.org/docs/)
- [Docker Compose 文档](https://docs.docker.com/compose/)

---

**文档维护**: 遇到新问题请及时补充到"常见问题"章节。

**最后审核**: 2026-02-04
