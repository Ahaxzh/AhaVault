# ==============================================================================
# AhaVault 前端 Docker 构建脚本 (Web Dockerfile)
# ==============================================================================
# 采用多阶段构建：Node.js 构建 + Nginx 服务

# --- 阶段 1: 编译阶段 (Build Stage) ---
# 使用 --platform=$BUILDPLATFORM 确保 Node.js 在宿主架构原生运行
FROM --platform=$BUILDPLATFORM node:20-alpine AS build-stage

WORKDIR /app

# 定义构建参数（API 地址）
ARG VITE_API_URL=/api
ENV VITE_API_URL=$VITE_API_URL

# 1. 缓存依赖安装
COPY web/package*.json ./
RUN npm install

# 2. 复制源码并构建
COPY web/ .
RUN npm run build


# --- 阶段 2: 生产阶段 (Production Stage) ---
# 使用目标架构的 Nginx
FROM nginx:stable-alpine AS production-stage

# 复制构建产物
COPY --from=build-stage /app/dist /usr/share/nginx/html

# 复制 Nginx 配置
COPY deploy/dockerfiles/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
