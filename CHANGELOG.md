# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added - 后端核心功能完成 (2026-02-04)

#### 🎯 核心里程碑
**后端所有核心功能已完成并编译成功！**
- ✅ 可执行文件: `bin/ahavault` (36MB)
- ✅ 编译状态: 成功通过
- ✅ 测试覆盖率: Config 80.6%, Crypto 75.0%, Storage 65.6%
- ✅ 总代码量: 26 个源文件，3000+ 行代码

---

#### 数据库层
- **数据库迁移脚本** `migrations/001_init.sql`
  - 9 张核心表（users, file_blobs, files_metadata, share_sessions, share_files, upload_sessions, system_settings, audit_logs）
  - 2 个实用视图（user_storage_stats, active_shares）
  - 完整的约束、索引、触发器
  - 默认系统配置种子数据

- **PostgreSQL 连接池管理** `internal/database/postgres.go`
  - GORM ORM 集成
  - 连接池配置（最大连接数、空闲连接数、生命周期）
  - 健康检查与 Ping 测试

- **Redis 客户端封装** `internal/database/redis.go`
  - Redis v8 客户端
  - 连接池管理
  - 常用操作封装（Set/Get/SetNX/Delete 等）

---

#### 核心服务层

##### **用户服务** `internal/services/user_service.go`
- ✅ 用户注册
  - 邮箱格式验证
  - 密码强度检查（至少 8 位，包含字母和数字）
  - bcrypt 密码哈希（默认 Cost）
  - 默认存储配额分配（10GB）

- ✅ 用户登录
  - 邮箱 + 密码认证
  - bcrypt 密码验证
  - JWT Token 生成（24 小时有效期）
  - 账户状态检查

- ✅ JWT 认证
  - Token 生成（HS256 签名）
  - Token 验证与解析
  - Claims 包含：user_id, email, is_admin, exp

##### **文件服务** `internal/services/file_service.go`
- ✅ 秒传检测
  - 基于 SHA-256 哈希检测
  - 文件禁止状态检查
  - 返回已存在文件的 Blob 信息

- ✅ 文件上传
  - 存储空间检查
  - SHA-256 哈希计算（流式）
  - 二次秒传检测
  - DEK 生成与文件加密（AES-256-GCM）
  - 物理文件存储（CAS）
  - 数据库事务管理（file_blobs + files_metadata + 存储使用量更新）
  - 引用计数初始化

- ✅ 文件下载
  - 文件权限验证
  - 过期状态检查
  - DEK 解密
  - 流式解密与传输

- ✅ 文件删除（软删除）
  - 软删除标记（deleted_at）
  - 引用计数递减（原子操作）
  - 存储使用量更新
  - 事务保证一致性

- ✅ 文件列表
  - 分页查询
  - 仅返回未删除文件
  - 按创建时间倒序

##### **分享服务** `internal/services/share_service.go`
- ✅ 创建分享
  - 文件所有权验证
  - 生成唯一取件码（8 位）
  - 密码哈希（可选）
  - 过期时间计算
  - 下载次数限制设置
  - 事务创建会话和文件关联

- ✅ 通过取件码访问
  - 取件码格式验证
  - 分享会话查询
  - 访问权限检查（过期、次数限制、禁用状态）
  - 密码验证（bcrypt）
  - 返回文件列表

- ✅ 转存到文件柜
  - 分享验证
  - 逐文件执行秒传逻辑
  - 引用计数管理
  - 下载计数增加

- ✅ 停止分享
  - 所有者验证
  - 软删除分享会话

- ✅ 我的分享列表
  - 分页查询
  - 按创建时间倒序

##### **取件码生成器** `internal/services/pickup_code.go`
- ✅ 随机生成
  - 8 位字符码
  - 字符集: `23456789ABCDEFGHJKLMNPQRSTUVWXYZ` (32 个字符)
  - 排除易混淆字符（0/O/1/I）
  - 加密安全随机数（crypto/rand）

- ✅ 唯一性保证
  - 数据库碰撞检测
  - 最多重试 10 次
  - 组合空间: 32^8 ≈ 1.1 万亿

- ✅ 格式验证
  - 长度检查
  - 字符集合法性验证

---

#### 加密模块

##### **信封加密** `internal/crypto/envelope.go`
- ✅ DEK 管理
  - 生成 32 字节随机 DEK
  - KEK 加密 DEK（AES-256-GCM）
  - DEK Base64 编码存储
  - KEK 解密 DEK

- ✅ 文件加密
  - AES-256-GCM 认证加密
  - 内存加密（适合小文件）
  - 流式加密（适合大文件）
  - 随机 Nonce 生成

- ✅ 文件解密
  - GCM 认证解密
  - 内存解密
  - 流式解密
  - 密钥安全清零

##### **哈希计算** `internal/crypto/hash.go`
- ✅ SHA-256 哈希
  - 内存哈希计算
  - 流式哈希（支持大文件）
  - 哈希验证
  - 哈希一致性测试

---

#### 存储引擎

##### **本地存储** `internal/storage/local.go`
- ✅ 内容寻址存储（CAS）
  - 基于 SHA-256 哈希的全局去重
  - 两级哈希分片（aa/bb/hash）
  - 65,536 个子目录

- ✅ 文件操作
  - 原子文件写入（temp + rename）
  - 流式读取
  - 删除操作
  - 存在性检查
  - 文件信息统计

- ✅ 安全性
  - 哈希格式验证
  - 目录自动创建
  - 防止重复写入

---

#### 配置管理

##### **统一配置** `internal/config/config.go`
- ✅ 7 个配置模块
  1. **App** - 应用基础配置（环境、调试、日志级别）
  2. **Database** - PostgreSQL 配置（连接池、超时）
  3. **Redis** - Redis 配置（连接池、超时）
  4. **Storage** - 存储引擎配置（Local/S3）
  5. **Crypto** - 加密配置（Master Key, JWT Secret）
  6. **Server** - HTTP 服务器配置（端口、超时）
  7. **Business** - 业务规则配置（配额、过期时间）

- ✅ 环境变量加载
  - 支持 .env 文件
  - 环境变量优先级
  - 默认值回退

- ✅ 配置验证
  - Master Key 长度验证（必须 32 字节）
  - 数据库密码必需
  - 存储类型验证

- ✅ 辅助方法
  - GetDSN() - 数据库连接字符串
  - GetRedisAddr() - Redis 地址
  - GetServerAddr() - 服务器地址

---

#### 中间件

##### **认证中间件** `internal/middleware/auth.go`
- ✅ JWT Token 验证
  - Authorization Header 解析
  - Bearer Token 提取
  - Token 签名验证
  - Claims 解析
  - 用户 ID 上下文传递

- ✅ 管理员认证
  - 普通认证 + 管理员权限检查
  - 用户角色验证
  - 403 Forbidden 响应

##### **CORS 中间件** `internal/middleware/cors.go`
- ✅ 跨域资源共享
  - Access-Control-Allow-Origin
  - Access-Control-Allow-Credentials
  - Access-Control-Allow-Headers
  - Access-Control-Allow-Methods
  - OPTIONS 预检请求处理

##### **错误处理** `internal/middleware/error.go`
- ✅ 全局错误捕获
  - Panic 恢复
  - 错误日志记录
  - 统一错误响应格式
  - 500 Internal Server Error

---

#### API 层

##### **认证接口** `internal/api/handlers/auth.go`
- ✅ POST `/api/auth/register` - 用户注册
  - 请求验证（邮箱、密码）
  - 调用 UserService.Register
  - 返回 Token + 用户信息

- ✅ POST `/api/auth/login` - 用户登录
  - 邮箱 + 密码验证
  - 调用 UserService.Login
  - 返回 Token + 用户信息

- ✅ POST `/api/auth/logout` - 用户登出
  - 客户端删除 Token

- ✅ GET `/api/user/me` - 获取当前用户
  - JWT 认证保护
  - 返回用户信息（隐藏密码）

##### **文件接口** `internal/api/handlers/file.go`
- ✅ GET `/api/files` - 文件列表
  - 分页参数（page, page_size）
  - 返回文件列表 + 总数

- ✅ POST `/api/files/check` - 秒传检测
  - 接收 SHA-256 哈希
  - 返回是否存在 + Blob 信息

- ✅ POST `/api/files` - 创建文件元数据（秒传）
  - 接收 hash, filename, size
  - 调用 FileService.CreateFileMetadata

- ✅ POST `/api/files/upload` - 上传新文件
  - Multipart 表单上传
  - 流式处理
  - 返回文件元数据

- ✅ GET `/api/files/:id/download` - 下载文件
  - 文件 ID 验证
  - 权限检查
  - 流式传输
  - Content-Disposition 头设置

- ✅ DELETE `/api/files/:id` - 删除文件
  - 文件 ID 验证
  - 软删除操作

##### **分享接口** `internal/api/handlers/share.go`
- ✅ GET `/api/shares` - 我的分享列表
  - 分页查询
  - 返回分享列表 + 总数

- ✅ POST `/api/shares` - 创建分享
  - 文件 ID 列表
  - 过期时间、下载次数、密码（可选）
  - 返回取件码

- ✅ POST `/api/public/shares/:code` - 通过取件码访问
  - 无需认证
  - 取件码验证
  - 密码验证（可选）
  - 返回分享信息 + 文件列表

- ✅ POST `/api/shares/:code/save` - 转存到文件柜
  - 需要认证
  - 选择文件 ID 列表
  - 秒传逻辑复用
  - 返回已转存文件 ID

- ✅ DELETE `/api/shares/:id` - 停止分享
  - 所有者验证
  - 软删除分享

##### **路由配置** `internal/api/routes.go`
- ✅ 全局中间件
  - CORS
  - ErrorHandler

- ✅ 公开路由
  - 认证接口（注册/登录）
  - 取件接口

- ✅ 认证路由
  - 文件管理
  - 分享管理
  - 用户信息

- ✅ 健康检查
  - GET `/health` - 返回 {"status": "ok"}

---

#### 主程序

##### **应用入口** `cmd/server/main.go`
- ✅ 配置加载
  - 环境变量加载
  - 配置验证

- ✅ 依赖初始化
  - PostgreSQL 连接池
  - Redis 客户端
  - 存储引擎（Local）

- ✅ 服务实例创建
  - UserService
  - FileService
  - ShareService

- ✅ 路由设置
  - Gin 路由器初始化
  - 中间件应用
  - 路由注册

- ✅ 服务器启动
  - HTTP 服务监听
  - 优雅日志输出

---

#### 数据模型

##### **User** `internal/models/user.go`
- 字段: ID, Email, Password, Role, Status, StorageQuota, StorageUsed
- 方法: IsAdmin(), IsActive(), HasStorageSpace(), AvailableStorage(), UpdateLastLogin()

##### **FileBlob** `internal/models/file_blob.go`
- 字段: Hash(主键), StorePath, EncryptedDEK, Size, MimeType, RefCount, IsBanned
- 方法: IncrementRefCount(), DecrementRefCount(), IsOrphan(), CanShare(), Ban(), Unban(), FormatSize()

##### **FileMetadata** `internal/models/file_metadata.go`
- 字段: ID, UserID, FileBlobHash, Filename, Size, ExpiresAt, DeletedAt
- 方法: IsExpired(), SoftDelete(), Restore()

##### **ShareSession** `internal/models/share_session.go`
- 字段: ID, PickupCode, CreatorID, PasswordHash, MaxDownloads, CurrentDownloads, ExpiresAt, StoppedAt
- 方法: HasPassword(), CanAccess(), IncrementDownloadCount(), Stop()

##### **ShareFile** `internal/models/share_file.go`
- 字段: ID, ShareID, FileID

##### **UploadSession** `internal/models/upload_session.go`
- 字段: ID, UserID, FileHash, UploadID, BytesUploaded, TotalBytes, Status

##### **SystemSetting** `internal/models/system_setting.go`
- 字段: Key, Value, Description
- 方法: GetValue(), SetValue(), GetBool(), GetInt(), GetInt64()

##### **AuditLog** `internal/models/audit_log.go`
- 字段: ID, UserID, Action, ResourceType, ResourceID, IPAddress, UserAgent, Details
- 方法: CreateLog()

---

### Fixed - 编译错误修复

#### Models 包错误
- ✅ 修复 file_blob.go 缺少导入（fmt, gorm.io/gorm）
- ✅ 修复 system_setting.go 缺少导入（fmt）
- ✅ 修复 audit_log.go 的 JSON 序列化方式
- ✅ 删除重复的 utils.go 文件（formatBytes 函数重复）

#### User 模型不一致
- ✅ 统一登录方式：使用 Email 替代 Username
- ✅ 修正字段访问：Password 替代 PasswordHash
- ✅ 修正方法调用：IsAdmin() 和 IsActive() 是方法而非字段

#### 配置文件错误
- ✅ 添加 JWTSecret 字段到 CryptoConfig
- ✅ 修正 loadCryptoConfig 函数加载 JWT Secret

#### 主程序错误
- ✅ 修正配置字段名：cfg.App.Env 替代 cfg.App.Environment
- ✅ 修正数据库初始化：database.InitPostgreSQL() 替代 database.Connect()
- ✅ 修正 Redis 初始化：database.InitRedis() 替代 database.ConnectRedis()
- ✅ 修正服务实例化：使用 database.DB 全局变量

#### 中间件错误
- ✅ 修正 AdminAuth 中的 IsAdmin 方法调用

#### API Handlers 错误
- ✅ 移除未使用的 models 导入
- ✅ 统一请求字段：Email 替代 Username

---

### Changed - 架构调整

#### 用户认证方式
- 从 Username + Password 改为 Email + Password
- JWT Claims 包含 email 而非 username
- 更符合现代 Web 应用习惯

#### User 模型设计
- Role 和 Status 改为字符串字段 + 方法访问
- IsAdmin() 和 IsActive() 提供业务逻辑封装
- 更灵活的角色扩展性

#### 数据库初始化方式
- 使用全局 database.DB 变量
- 简化服务实例化逻辑
- 统一数据库访问入口

---

### Dependencies - 依赖管理

#### 新增依赖
- ✅ `github.com/gin-gonic/gin` v1.11.0 - HTTP Web 框架
- ✅ `github.com/golang-jwt/jwt/v5` v5.3.1 - JWT 认证
- ✅ `github.com/go-redis/redis/v8` - Redis 客户端
- ✅ `gorm.io/gorm` v1.25.12 - ORM 框架
- ✅ `gorm.io/driver/postgres` - PostgreSQL 驱动
- ✅ `golang.org/x/crypto` - 加密库（bcrypt, AES）
- ✅ `github.com/google/uuid` - UUID 生成

#### Go 版本升级
- 从 Go 1.21 升级到 Go 1.23

---

### Tests - 测试覆盖

#### 已完成测试
- ✅ Config 模块: 80.6% 覆盖率
  - 环境变量加载测试
  - 配置验证测试
  - DSN 生成测试

- ✅ Crypto 模块: 75.0% 覆盖率
  - DEK 生成与加密测试
  - 文件加密/解密测试
  - SHA-256 哈希测试
  - 流式加密测试

- ✅ Storage 模块: 65.6% 覆盖率
  - 本地存储引擎测试
  - 文件 Put/Get/Delete 测试
  - 哈希验证测试
  - 大文件测试

- ✅ Services 模块: 88.2% 覆盖率（取件码生成器）
  - 取件码生成测试
  - 唯一性验证测试
  - 格式验证测试

#### 待补充测试
- ⚠️ UserService 集成测试
- ⚠️ FileService 集成测试
- ⚠️ ShareService 集成测试
- ⚠️ API Handlers E2E 测试
- ⚠️ Middleware 单元测试

---

### Documentation - 文档更新

#### 新增文档
- ✅ **server/README.md** - 后端完整开发文档（300+ 行）
  - 项目概述与技术栈
  - 完整的目录结构说明
  - 核心模块详细解析
  - 快速开始指南
  - API 接口总览
  - 配置说明（环境变量清单）
  - 测试指南
  - 故障排查
  - 性能优化建议
  - 安全注意事项
  - 开发者指南

#### 更新文档
- ✅ **CHANGELOG.md** - 本次更新（新增 200+ 行变更记录）
- ✅ **Claude.md** - 保持最新的协作规范

#### 待更新文档
- ⚠️ docs/api/API.md - 需要同步最新接口定义
- ⚠️ README.md - 需要更新项目进度

---

## [0.1.0] - 2026-02-04

### Added - 项目初始化
- 初始化项目结构 (前端 web/ 和后端 server/)
- 完整的 Docker Compose 编排配置
- 环境变量模板 (.env.example)
- 项目文档框架 (docs/)
- 统一的人机协作规范体系 (Claude.md)

### Backend - Core Infrastructure (阶段 2)
- **数据库迁移**: server/migrations/001_init.sql
- **配置管理**: server/internal/config/
- **数据库连接层**: server/internal/database/
- **数据模型**: server/internal/models/

### Documentation
- docs/PRD.md - 产品需求文档 v1.2
- docs/PRD_Analysis.md - PRD 技术分析报告
- docs/api/API.md - RESTful API 接口文档
- docs/architecture/encryption.md - 信封加密架构设计
- docs/architecture/storage.md - CAS 存储引擎架构设计
- docs/guides/development.md - 本地开发环境指南

---

**注意**: 本项目遵循 [Conventional Commits](https://www.conventionalcommits.org/) 提交规范。

**版本说明**:
- `[Unreleased]` - 开发中的功能，已完成后端核心模块
- `[0.1.0]` - 首个里程碑版本
