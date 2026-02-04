# AhaVault 部署指南

**版本**: v1.0
**最后更新**: 2026-02-04
**适用环境**: 生产环境

---

## 📋 目录

- [快速开始](#快速开始)
- [部署架构](#部署架构)
- [环境要求](#环境要求)
- [部署步骤](#部署步骤)
- [配置说明](#配置说明)
- [维护管理](#维护管理)
- [故障排查](#故障排查)

---

## 快速开始

### 最小化部署（3 步）

```bash
# 1. 克隆项目
git clone https://github.com/Ahaxzh/AhaVault.git
cd AhaVault/deploy

# 2. 配置环境变量
cp .env.example .env
vim .env  # 修改必需的配置项

# 3. 启动服务
docker-compose up -d
```

访问：http://your-server-ip

---

## 部署架构

### 容器架构图

```
┌─────────────────────────────────────────────────────┐
│                   Internet                           │
└───────────────────┬─────────────────────────────────┘
                    │ Port 80/443
         ┌──────────▼──────────┐
         │   Web (Nginx)       │
         │  - 静态文件托管      │
         │  - API 反向代理     │
         └────────┬────────────┘
                  │ /api/*
       ┌──────────▼──────────┐
       │   Server (Go)       │
       │  - 业务逻辑         │
       │  - API 服务         │
       └──┬───────────────┬──┘
          │               │
    ┌─────▼─────┐   ┌────▼─────┐
    │ PostgreSQL│   │  Redis   │
    │  - 数据库  │   │  - 缓存   │
    └───────────┘   └──────────┘
```

### 数据持久化

- `ahavault_postgres_data` - PostgreSQL 数据
- `ahavault_redis_data` - Redis 数据
- `ahavault_storage_data` - 用户文件存储（加密）
- `ahavault_temp_data` - 临时文件

---

## 环境要求

### 硬件要求

**最小配置**:
- CPU: 2 核心
- 内存: 4GB RAM
- 磁盘: 20GB SSD

**推荐配置**:
- CPU: 4 核心
- 内存: 8GB RAM
- 磁盘: 100GB SSD

### 软件要求

- **Docker**: 20.10+
- **Docker Compose**: 2.0+
- **操作系统**:
  - Ubuntu 20.04+
  - Debian 11+
  - CentOS 8+
  - macOS 12+

---

## 部署步骤

### 1. 准备服务器

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装 Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# 安装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 验证安装
docker --version
docker-compose --version
```

### 2. 克隆项目

```bash
git clone https://github.com/Ahaxzh/AhaVault.git
cd AhaVault/deploy
```

### 3. 配置环境变量

```bash
# 复制模板
cp .env.example .env

# 生成密钥
echo "APP_MASTER_KEY=$(openssl rand -hex 32)" >> .env
echo "JWT_SECRET=$(openssl rand -base64 64)" >> .env

# 编辑配置
vim .env
```

**必需修改的配置**:
```env
# 关键安全配置
APP_MASTER_KEY=your-64-char-hex-key-here
JWT_SECRET=your-jwt-secret-here

# 数据库密码
POSTGRES_PASSWORD=your-strong-password

# Redis 密码
REDIS_PASSWORD=your-redis-password

# 前端 API 地址（如果使用域名）
VITE_API_URL=https://yourdomain.com/api
```

### 4. 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看启动状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 5. 验证部署

```bash
# 检查服务健康状态
docker-compose ps

# 预期输出：所有服务状态为 Up (healthy)
# NAME                 IMAGE                              STATUS
# ahavault_postgres    postgres:16-alpine                 Up (healthy)
# ahavault_redis       redis:7-alpine                     Up (healthy)
# ahavault_server      ghcr.io/.../ahavault-server:latest Up
# ahavault_web         ghcr.io/.../ahavault-web:latest    Up

# 测试 API
curl http://localhost/health
# 返回: {"status":"ok"}
```

---

## 配置说明

### 环境变量详解

详细配置说明请参考 `.env.example` 文件中的注释。

### 关键配置

#### 1. **APP_MASTER_KEY** (必需)
- 用途：加密所有用户文件的 DEK
- 格式：64 字符 HEX（32 字节）
- 生成：`openssl rand -hex 32`
- ⚠️ **重要**：丢失此密钥将无法解密任何文件！

#### 2. **JWT_SECRET** (必需)
- 用途：签署用户认证令牌
- 生成：`openssl rand -base64 64`
- ⚠️ **重要**：修改此密钥会导致所有用户登出

#### 3. **POSTGRES_PASSWORD** (必需)
- 用途：数据库认证密码
- 生成：`openssl rand -base64 32`

#### 4. **REDIS_PASSWORD** (必需)
- 用途：Redis 认证密码
- 生成：`openssl rand -base64 32`

---

## 维护管理

### 日志管理

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f server
docker-compose logs -f web

# 清理日志（Docker 自动轮转）
docker system prune -a --volumes
```

### 数据备份

```bash
# 备份 PostgreSQL
docker exec ahavault_postgres pg_dump -U ahavault ahavault > backup_$(date +%Y%m%d).sql

# 备份存储文件
docker run --rm -v ahavault_storage_data:/data -v $(pwd):/backup alpine tar czf /backup/storage_$(date +%Y%m%d).tar.gz -C /data .

# 备份 APP_MASTER_KEY
echo "⚠️ 请将 APP_MASTER_KEY 保存到安全位置！"
```

### 数据恢复

```bash
# 恢复 PostgreSQL
docker exec -i ahavault_postgres psql -U ahavault -d ahavault < backup_20260204.sql

# 恢复存储文件
docker run --rm -v ahavault_storage_data:/data -v $(pwd):/backup alpine tar xzf /backup/storage_20260204.tar.gz -C /data
```

### 更新服务

```bash
# 拉取最新镜像
docker-compose pull

# 重启服务
docker-compose up -d

# 清理旧镜像
docker image prune -a
```

### 扩容指南

#### 增加用户存储配额
```bash
# 进入数据库
docker exec -it ahavault_postgres psql -U ahavault -d ahavault

# 增加特定用户配额（单位：字节）
UPDATE users SET storage_quota = 107374182400 WHERE email = 'user@example.com';

# 查看用户配额使用情况
SELECT email, storage_used, storage_quota,
       ROUND((storage_used::float / storage_quota * 100)::numeric, 2) as usage_percent
FROM users;
```

#### 增加文件大小限制
编辑 `.env`:
```env
MAX_FILE_SIZE=4294967296  # 4GB
```

重启服务：
```bash
docker-compose up -d
```

---

## 故障排查

### 常见问题

#### Q1: 服务无法启动

```bash
# 查看详细错误
docker-compose logs

# 检查端口占用
sudo netstat -tulpn | grep :80

# 检查磁盘空间
df -h
```

#### Q2: 数据库连接失败

```bash
# 检查数据库健康状态
docker-compose ps

# 重启数据库
docker-compose restart postgres

# 查看数据库日志
docker-compose logs postgres
```

#### Q3: 文件上传失败

**可能原因**:
1. 文件超过大小限制
2. 存储空间不足
3. 权限问题

**解决方案**:
```bash
# 检查存储卷空间
docker system df -v

# 检查存储目录权限
docker exec -it ahavault_server ls -la /data/storage

# 增加文件大小限制（修改 .env）
MAX_FILE_SIZE=4294967296
```

#### Q4: 无法登录

```bash
# 检查 JWT_SECRET 是否正确
docker exec ahavault_server env | grep JWT_SECRET

# 重置用户密码（进入数据库）
docker exec -it ahavault_postgres psql -U ahavault -d ahavault
UPDATE users SET password = '$2a$10$...' WHERE email = 'user@example.com';
```

---

## 安全加固

### 1. 使用 HTTPS

推荐使用 Let's Encrypt + Nginx：

```bash
# 安装 Certbot
sudo apt install certbot python3-certbot-nginx

# 获取证书
sudo certbot --nginx -d yourdomain.com

# 自动续期
sudo certbot renew --dry-run
```

### 2. 防火墙配置

```bash
# 仅开放必要端口
sudo ufw allow 22/tcp   # SSH
sudo ufw allow 80/tcp   # HTTP
sudo ufw allow 443/tcp  # HTTPS
sudo ufw enable
```

### 3. 定期更新

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 更新 Docker 镜像
docker-compose pull
docker-compose up -d
```

---

## 性能优化

### 1. 数据库优化

编辑 `.env`:
```env
# 根据服务器配置调整
DB_MAX_OPEN_CONNS=100
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=1h
```

### 2. Redis 优化

```bash
# 增加 Redis 内存限制
docker-compose down
# 编辑 docker-compose.yml，修改 redis command:
# command: ["redis-server", "--requirepass", "${REDIS_PASSWORD}", "--maxmemory", "256mb"]
docker-compose up -d
```

### 3. Nginx 缓存

参考 `deploy/dockerfiles/nginx.conf` 添加缓存配置。

---

## 监控建议

### Prometheus + Grafana

可选：集成 Prometheus 监控

```bash
# 添加 Prometheus exporter
# 在 docker-compose.yml 中添加 prometheus 和 grafana 服务
```

### 日志聚合

可选：使用 ELK Stack 或 Loki

---

## 相关文档

- **本地开发指南**: `docs/guides/development.md`
- **API 文档**: `docs/api/API.md`
- **后端开发文档**: `server/README.md`
- **架构设计**: `docs/architecture/`

---

**维护者**: 开发团队
**最后更新**: 2026-02-04
