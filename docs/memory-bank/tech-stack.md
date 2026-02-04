# 🔧 AhaVault 技术栈

**最后更新**: 2026-02-04
**维护者**: 开发团队

---

## 📊 技术选型总览

| 层级 | 技术 | 版本 | 选型理由 |
|------|------|------|---------|
| **后端框架** | Go + Gin | Go 1.21+ | 高性能、并发友好、类型安全 |
| **数据库** | PostgreSQL | 16 | ACID 保证、JSON 支持、成熟稳定 |
| **缓存** | Redis | 7 | 高性能、支持限流、会话管理 |
| **前端框架** | React | 18+ | 生态成熟、组件化、TypeScript 支持 |
| **构建工具** | Vite | 5+ | 快速热更新、现代化工具链 |
| **样式方案** | TailwindCSS | 3+ | 实用优先、快速开发、可定制 |
| **类型安全** | TypeScript | 5+ | 静态类型检查、IDE 支持好 |
| **文件上传** | Tus Protocol | 1.0 | 断点续传标准、可靠性高 |
| **容器化** | Docker | 20.10+ | 环境一致性、易于部署 |
| **CI/CD** | GitHub Actions | - | 免费、与仓库集成紧密 |

---

## 🔐 后端技术栈

### 核心框架与库

```go
require (
    github.com/gin-gonic/gin v1.9+          // HTTP 框架
    gorm.io/gorm v1.25+                     // ORM 框架
    gorm.io/driver/postgres v1.5+           // PostgreSQL 驱动
    github.com/redis/go-redis/v9 v9.0+      // Redis 客户端
    github.com/golang-jwt/jwt/v5 v5.0+      // JWT 认证
    golang.org/x/crypto v0.17+              // 密码加密 (bcrypt)
)
```

### 为什么选择 Go + Gin？

**优势**:
- ✅ **高性能**: 原生支持并发，处理文件上传下载效率高
- ✅ **类型安全**: 编译期错误检查，减少运行时 Bug
- ✅ **部署简单**: 单二进制文件，无运行时依赖
- ✅ **生态成熟**: GORM、Redis、Tus 都有成熟的 Go 实现
- ✅ **学习曲线**: 语法简洁，易于维护

**决策记录**: [decisions/0001-choose-go-gin.md](../decisions/0001-choose-go-gin.md)

---

### 为什么选择 PostgreSQL？

**优势**:
- ✅ **ACID 事务**: 引用计数更新必须在事务中保证一致性
- ✅ **JSON 支持**: `audit_logs.details` 使用 JSONB 存储
- ✅ **成熟稳定**: 生产环境验证充分
- ✅ **开源免费**: 无许可证成本

**替代方案**:
- ❌ MySQL: JSON 支持较弱，事务隔离级别不如 PostgreSQL
- ❌ SQLite: 不适合并发写入，无法用于生产环境
- ❌ MongoDB: 无 ACID 事务，引用计数一致性难以保证

**决策记录**: [decisions/0002-choose-postgresql.md](../decisions/0002-choose-postgresql.md)

---

### 为什么选择 Redis？

**优势**:
- ✅ **高性能**: 内存存储，支持百万级 QPS
- ✅ **原子操作**: `INCR` 命令用于实现限流计数器
- ✅ **过期机制**: 自动清理过期的限流记录和会话
- ✅ **数据结构丰富**: String、Hash、Set 满足多种场景

**用途**:
1. **限流**: IP 限流、API 限流（基于 Redis INCR + EXPIRE）
2. **会话缓存**: 用户登录状态（可选，目前使用 JWT）
3. **分享信息缓存**: 减少数据库查询

---

## 🎨 前端技术栈

### 核心框架与库

```json
{
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-router-dom": "^6.20.0",
    "axios": "^1.6.0",
    "zustand": "^4.4.0"
  },
  "devDependencies": {
    "vite": "^5.0.0",
    "typescript": "^5.3.0",
    "tailwindcss": "^3.4.0",
    "vitest": "^1.0.0",
    "@playwright/test": "^1.40.0"
  }
}
```

### 为什么选择 React + Vite？

**React 优势**:
- ✅ 生态成熟，组件库丰富
- ✅ TypeScript 支持优秀
- ✅ Hooks 简化状态管理
- ✅ 团队熟悉度高

**Vite 优势**:
- ✅ 开发服务器启动速度极快（< 1 秒）
- ✅ 热更新响应迅速
- ✅ 原生 ESM 支持
- ✅ 构建速度快（基于 Rollup）

**替代方案**:
- ❌ Next.js: 过于重量级，不需要 SSR
- ❌ Vue: 团队不熟悉
- ❌ Svelte: 生态不够成熟

---

### 为什么选择 TailwindCSS？

**优势**:
- ✅ 实用优先，开发速度快
- ✅ 无需切换文件，CSS 与 HTML 在一起
- ✅ Tree-shaking，生产体积小
- ✅ 响应式设计友好
- ✅ 深色模式支持内置

**示例**:
```tsx
<button className="px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg transition">
  上传文件
</button>
```

---

## 🔒 加密与安全

### 加密算法

| 用途 | 算法 | 密钥长度 | 理由 |
|------|------|---------|------|
| 文件加密 | AES-256-GCM | 256-bit | AEAD 加密，带认证 |
| 密钥加密 | AES-256-GCM | 256-bit | 加密 DEK |
| 哈希计算 | SHA-256 | - | 内容去重、完整性校验 |
| 密码加密 | bcrypt | cost=10 | 抗暴力破解 |
| JWT 签名 | HS256 | 256-bit | 对称签名，性能好 |

**决策记录**: [decisions/0003-envelope-encryption.md](../decisions/0003-envelope-encryption.md)

---

## 📦 文件上传协议

### 为什么选择 Tus Protocol？

**Tus 优势**:
- ✅ **断点续传**: 网络中断后可恢复上传
- ✅ **分片上传**: 大文件分片传输，减少内存占用
- ✅ **标准协议**: 开源标准，客户端库丰富
- ✅ **进度跟踪**: 实时获取上传进度

**Go 实现**: `github.com/tus/tusd`
**前端库**: `tus-js-client`

**替代方案**:
- ❌ 普通 Multipart Upload: 不支持断点续传
- ❌ 自定义协议: 维护成本高

---

## 🐳 部署与运维

### Docker 镜像选择

| 服务 | 基础镜像 | 理由 |
|------|---------|------|
| Go Server | `alpine:latest` | 最小体积（5MB），安全 |
| Web (Nginx) | `nginx:stable-alpine` | 稳定版，体积小 |
| PostgreSQL | `postgres:16-alpine` | 官方镜像，Alpine 变体 |
| Redis | `redis:7-alpine` | 官方镜像，最新稳定版 |

### 为什么选择 Alpine Linux？

- ✅ 体积极小（5MB vs Ubuntu 的 70MB）
- ✅ 安全性高（攻击面小）
- ✅ 包管理器 apk 高效
- ✅ Docker 官方推荐

---

## 🧪 测试工具

### 后端测试

| 工具 | 用途 | 理由 |
|------|------|------|
| `go test` | 单元测试 | Go 官方测试框架 |
| `testify/assert` | 断言库 | 语法简洁，易读 |
| `testify/mock` | Mock 工具 | 隔离依赖 |

### 前端测试

| 工具 | 用途 | 理由 |
|------|------|------|
| Vitest | 单元测试 | 与 Vite 集成，速度快 |
| @testing-library/react | 组件测试 | React 官方推荐 |
| Playwright | E2E 测试 | 跨浏览器，调试友好 |

---

## 📚 相关文档

- **架构设计**: [architecture/](../architecture/)
- **开发指南**: [guides/development.md](../guides/development.md)
- **部署指南**: [guides/deployment.md](../guides/deployment.md)
- **API 文档**: [api/](../api/)

---

**维护者**: 开发团队
**最后审核**: 2026-02-04
