# AhaVault - 安全文件分享系统

<div align="center">

**私有化 · 轻量级 · 极致安全**

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18+-61DAFB?logo=react)](https://react.dev/)
[![PRD Version](https://img.shields.io/badge/PRD-v1.2-green)](docs/PRD.md)

</div>

## 📖 项目简介

AhaVault 是一个基于"文件柜"与"取件码"概念的文件分享系统，专注于文件在不同设备、不同人员之间的高效、隐私交换。

### ✨ 核心特性

- 🔒 **安全优先**: 信封加密架构，全链路 HTTPS 传输
- 🎯 **极简设计**: 无广告、无冗余功能，即用即走
- ⚡ **高效传输**: 内容去重秒传、断点续传、Web Worker 计算
- 🛡️ **可控管理**: 管理员致盲管理，完善日志监控

## 🏗️ 技术架构

### 前端
- **框架**: React 18 + TypeScript + Vite
- **UI**: TailwindCSS (深色/浅色主题)
- **上传**: Tus-JS-Client (断点续传)
- **多线程**: Web Worker (SHA-256 哈希计算)

### 后端
- **语言**: Go 1.21+
- **框架**: Gin / Echo
- **数据库**: PostgreSQL + Redis
- **存储**: Local Filesystem / S3

### 基础设施
- **容器化**: Docker + Docker Compose
- **反向代理**: Nginx

## 🚀 快速开始

### 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- (可选) Node.js 18+ / Go 1.21+

### 一键部署

```bash
# 1. 克隆项目
git clone https://github.com/yourusername/AhaVault.git
cd AhaVault

# 2. 配置环境变量
cp .env.example .env
# 编辑 .env 文件，至少修改以下项:
# - APP_MASTER_KEY (使用 openssl rand -hex 32 生成)
# - POSTGRES_PASSWORD
# - REDIS_PASSWORD

# 3. 启动服务
docker-compose up -d

# 4. 访问应用
# - 前端: http://localhost
# - 后端 API: http://localhost:8080
```

### 本地开发

<details>
<summary>前端开发</summary>

```bash
cd web
npm install
npm run dev
# 访问 http://localhost:5173
```
</details>

<details>
<summary>后端开发</summary>

```bash
cd server
go mod download
go run cmd/server/main.go
# API 运行在 http://localhost:8080
```
</details>

## 📁 项目结构

```
AhaVault/
├── web/              # 前端项目
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── services/
│   │   └── workers/
│   └── package.json
├── server/           # 后端项目
│   ├── cmd/
│   ├── internal/
│   │   ├── api/
│   │   ├── models/
│   │   ├── services/
│   │   ├── storage/
│   │   └── crypto/
│   └── go.mod
├── docs/             # 文档
│   ├── PRD.md        # 产品需求文档
│   └── API.md        # API 文档
├── docker/           # Docker 配置
├── nginx/            # Nginx 配置
└── docker-compose.yml
```

## 🔐 核心设计

### 信封加密 (Envelope Encryption)

```
用户文件 → DEK 加密 → 存储密文
         ↓
       KEK 加密 → 加密后的 DEK → 存储到数据库
```

- **KEK (Master Key)**: 全局密钥，环境变量注入
- **DEK (Data Key)**: 每个文件独立密钥，随机生成
- **优势**: 支持密钥轮换，数据库泄露不影响文件安全

### 内容寻址存储 (CAS)

- 基于 SHA-256 哈希去重
- 引用计数管理 (数据库事务保证一致性)
- 目录结构: `/storage/{aa}/{bb}/{sha256_hash}`

### 取件码系统

- **格式**: 8 位字符 (数字 + 大写字母)
- **安全**: IP 限流 + Captcha 验证
- **功能**: 访问密码、次数限制、时效控制

## 📚 文档

- [完整 PRD](docs/PRD.md) - 软件需求规格说明书
- [PRD 分析](docs/PRD_Analysis.md) - 技术分析与改进建议
- [Claude 上下文](Claude.md) - AI 助手开发文档
- [前端 README](web/README.md) - 前端开发指南
- [后端 README](server/README.md) - 后端开发指南

## 🛠️ 开发规范

### Git 提交规范

```
feat: 新功能
fix: 修复 Bug
docs: 文档更新
refactor: 重构
test: 测试相关
chore: 构建/工具链
```

### 代码风格

- **前端**: ESLint + Prettier
- **后端**: gofmt + golangci-lint

## 🔒 安全声明

本系统采用端到端加密设计，服务端不对文件内容进行扫描。用户下载文件后请自行进行安全检查。

## 📝 License

[MIT License](LICENSE)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request!

## 📮 联系方式

- Issue: [GitHub Issues](https://github.com/yourusername/AhaVault/issues)
- Email: your-email@example.com

---

<div align="center">
Made with ❤️ by AhaVault Team
</div>
