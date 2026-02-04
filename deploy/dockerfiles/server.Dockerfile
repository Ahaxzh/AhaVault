# ==============================================================================
# AhaVault 后端 Docker 构建脚本 (Server Dockerfile)
# ==============================================================================
# 采用多阶段构建优化镜像体积
# Multi-stage build for optimized image size

# --- 阶段 1: 编译阶段 (Build Stage) ---
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git

# 缓存依赖层
COPY server/go.mod server/go.sum ./
RUN go mod download

# 复制源码
COPY server/ .

# 交叉编译
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags="-w -s" -o ahavault ./cmd/server/main.go


# --- 阶段 2: 运行时阶段 (Final Stage) ---
FROM alpine:latest

# 安装运行时依赖（时区、证书）
RUN apk add --no-cache tzdata ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 复制编译产物
COPY --from=builder /app/ahavault .

# 创建数据目录
RUN mkdir -p /data/storage /data/temp

# 暴露端口
EXPOSE 8080

# 启动服务
CMD ["./ahavault"]
