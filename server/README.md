# AhaVault - 后端项目

## 技术栈
- Go 1.21+
- Gin / Echo (Web 框架)
- PostgreSQL (主数据库)
- Redis (缓存 + 限流)
- Tus Protocol (断点续传)

## 目录结构

```
server/
├── cmd/
│   └── server/           # 主程序入口
│       └── main.go
├── internal/
│   ├── api/              # HTTP 路由与控制器
│   │   ├── routes.go
│   │   ├── handlers/
│   │   │   ├── file.go
│   │   │   ├── share.go
│   │   │   ├── user.go
│   │   │   └── admin.go
│   ├── models/           # 数据模型
│   │   ├── user.go
│   │   ├── file.go
│   │   ├── share.go
│   │   └── blob.go
│   ├── services/         # 业务逻辑层
│   │   ├── file_service.go
│   │   ├── share_service.go
│   │   └── user_service.go
│   ├── storage/          # 存储引擎
│   │   ├── interface.go
│   │   ├── local.go      # 本地存储实现
│   │   └── s3.go         # S3 存储实现
│   ├── crypto/           # 加密模块
│   │   ├── envelope.go   # 信封加密
│   │   └── hash.go       # 哈希计算
│   ├── middleware/       # 中间件
│   │   ├── auth.go
│   │   ├── ratelimit.go
│   │   └── captcha.go
│   ├── config/           # 配置管理
│   │   └── config.go
│   └── tasks/            # 后台任务
│       ├── gc.go         # 垃圾回收
│       └── lifecycle.go  # 生命周期检查
├── pkg/                  # 可复用的公共包
├── migrations/           # 数据库迁移文件
│   └── 001_init.sql
└── go.mod
```

## 核心模块说明

### 1. 信封加密 (crypto/envelope.go)
```go
// KEK 加密 DEK
func EncryptDEK(dek []byte, kek []byte) ([]byte, error)

// KEK 解密 DEK
func DecryptDEK(encryptedDEK []byte, kek []byte) ([]byte, error)

// DEK 加密文件流
func EncryptStream(reader io.Reader, dek []byte) (io.Reader, error)
```

### 2. 存储引擎 (storage/interface.go)
```go
type StorageEngine interface {
    Put(hash string, reader io.Reader) error
    Get(hash string) (io.ReadCloser, error)
    Delete(hash string) error
    Exists(hash string) (bool, error)
}
```

### 3. 引用计数管理 (services/file_service.go)
- 所有涉及 `ref_count` 操作必须在数据库事务中完成
- 禁止应用层计算，避免并发问题

### 4. 后台任务 (tasks/)
- 碎片清理: 每小时
- 元数据清理: 每日凌晨
- 孤儿文件清理: 每日凌晨
- 生命周期检查: 每分钟

## 开发指南

### 安装依赖
```bash
go mod download
```

### 运行开发服务器
```bash
go run cmd/server/main.go
```

### 数据库迁移
```bash
# 使用 golang-migrate
migrate -path migrations -database "postgres://localhost:5432/ahavault" up
```

### 构建生产版本
```bash
CGO_ENABLED=0 GOOS=linux go build -o ahavault cmd/server/main.go
```

## 环境变量

```bash
# 核心配置
APP_MASTER_KEY=xxx          # KEK 密钥 (32字节)
APP_INVITE_CODE=xxx         # 全局邀请码

# 数据库
POSTGRES_DSN=postgres://user:pass@localhost:5432/ahavault
REDIS_URL=redis://localhost:6379/0

# 存储
STORAGE_TYPE=local|s3
STORAGE_PATH=/data/storage  # local 模式
S3_ENDPOINT=xxx
S3_BUCKET=xxx
S3_ACCESS_KEY=xxx
S3_SECRET_KEY=xxx

# 安全
TURNSTILE_SECRET_KEY=xxx
MAX_FILE_SIZE=2147483648    # 2GB
```

## 安全规范

1. **密钥管理**: KEK 必须通过环境变量注入，严禁硬编码
2. **事务管理**: 引用计数操作必须使用事务
3. **输入校验**: 所有用户输入必须校验 (文件类型、大小、格式)
4. **错误处理**: 敏感错误不得暴露给客户端
5. **日志安全**: 禁止记录密钥、密码等敏感信息

## API 接口规范

### RESTful 风格
```
POST   /api/auth/login
POST   /api/auth/register
GET    /api/files
POST   /api/files/upload
DELETE /api/files/:id
POST   /api/shares
GET    /api/shares/:code
```

### Tus 端点
```
POST   /api/tus/upload  # 创建上传
PATCH  /api/tus/upload/:id  # 分片上传
HEAD   /api/tus/upload/:id  # 获取上传进度
```
