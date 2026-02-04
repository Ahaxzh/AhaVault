# AhaVault - RESTful API 接口文档

**版本**: v1.0
**最后更新**: 2026-02-04
**负责人**: Claude AI
**关联模块**: server/internal/api

---

## 📋 目录

- [1. 概述](#1-概述)
- [2. 认证接口](#2-认证接口)
- [3. 文件管理接口](#3-文件管理接口)
- [4. 分享管理接口](#4-分享管理接口)
- [5. 管理员接口](#5-管理员接口)
- [6. 错误码说明](#6-错误码说明)

---

## 1. 概述

### 1.1 基础信息

- **协议**: HTTPS (强制)
- **Base URL**: `https://your-domain.com/api`
- **Content-Type**: `application/json`
- **认证方式**: JWT Bearer Token

### 1.2 通用请求头

```http
Content-Type: application/json
Authorization: Bearer <JWT_TOKEN>  # 需要认证的接口
X-Request-ID: <UUID>               # 可选，用于追踪请求
```

### 1.3 通用响应格式

**成功响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    // 具体业务数据
  },
  "timestamp": 1704412800
}
```

**错误响应**:
```json
{
  "code": 4001,
  "message": "Invalid pickup code",
  "data": null,
  "timestamp": 1704412800
}
```

### 1.4 分页参数

```json
{
  "page": 1,      // 页码，从 1 开始
  "page_size": 20 // 每页数量，默认 20，最大 100
}
```

**分页响应**:
```json
{
  "items": [...],
  "total": 100,
  "page": 1,
  "page_size": 20
}
```

---

## 2. 认证接口

### 2.1 用户注册

**端点**: `POST /auth/register`

**权限**: 公开（需邀请码或开启注册）

**请求体**:
```json
{
  "email": "user@example.com",
  "password": "StrongPassword123!",
  "invite_code": "AHAVAULT2026"  // 可选，取决于系统配置
}
```

**响应**:
```json
{
  "code": 0,
  "message": "Registration successful",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400
  }
}
```

**错误码**:
- `4001`: 邮箱已注册
- `4002`: 邀请码无效或未提供
- `4003`: 注册功能已关闭

---

### 2.2 用户登录

**端点**: `POST /auth/login`

**权限**: 公开

**请求体**:
```json
{
  "email": "user@example.com",
  "password": "StrongPassword123!",
  "captcha_token": "turnstile_token_here"  // 可选，触发限流时需要
}
```

**响应**:
```json
{
  "code": 0,
  "message": "Login successful",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "role": "user",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400
  }
}
```

**错误码**:
- `4010`: 邮箱或密码错误
- `4011`: 账户已被禁用
- `4012`: 需要人机验证（返回 captcha_required: true）
- `4013`: 验证码校验失败

---

### 2.3 刷新 Token

**端点**: `POST /auth/refresh`

**权限**: 需要认证

**请求头**:
```http
Authorization: Bearer <OLD_TOKEN>
```

**响应**:
```json
{
  "code": 0,
  "message": "Token refreshed",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400
  }
}
```

---

### 2.4 退出登录

**端点**: `POST /auth/logout`

**权限**: 需要认证

**响应**:
```json
{
  "code": 0,
  "message": "Logout successful"
}
```

---

## 3. 文件管理接口

### 3.1 获取文件列表

**端点**: `GET /files`

**权限**: 需要认证

**查询参数**:
```
?page=1&page_size=20&search=filename&type=image&sort=created_at&order=desc
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页数量，默认 20 |
| search | string | 否 | 文件名搜索（模糊匹配） |
| type | string | 否 | 文件类型筛选：image/video/document/archive |
| sort | string | 否 | 排序字段：created_at/size/filename |
| order | string | 否 | 排序方向：asc/desc，默认 desc |

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "filename": "vacation_photo.jpg",
        "size": 2048576,
        "mime_type": "image/jpeg",
        "hash": "aabbccddeeff11223344556677889900...",
        "created_at": "2026-02-04T10:30:00Z",
        "expires_at": "2026-03-04T10:30:00Z",
        "is_shared": true,
        "share_count": 2
      }
    ],
    "total": 42,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 3.2 秒传检测

**端点**: `POST /files/check`

**权限**: 需要认证

**请求体**:
```json
{
  "hash": "aabbccddeeff1122334455667788990011223344556677889900aabbccddeeff",
  "size": 2048576
}
```

**响应**:

**情况 1: 文件已存在（秒传成功）**
```json
{
  "code": 0,
  "message": "File exists, instant upload available",
  "data": {
    "exists": true,
    "file_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**情况 2: 文件不存在（需要上传）**
```json
{
  "code": 0,
  "message": "File does not exist, please upload",
  "data": {
    "exists": false,
    "upload_url": "/api/tus/upload"  // Tus 协议上传端点
  }
}
```

---

### 3.3 完成秒传（创建文件元数据）

**端点**: `POST /files`

**权限**: 需要认证

**说明**: 当秒传检测返回 `exists: true` 时，调用此接口创建用户的文件元数据记录（不上传物理文件）。

**请求体**:
```json
{
  "hash": "aabbccddeeff1122334455667788990011223344556677889900aabbccddeeff",
  "filename": "my_document.pdf",
  "size": 2048576,
  "mime_type": "application/pdf"
}
```

**响应**:
```json
{
  "code": 0,
  "message": "File created successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "filename": "my_document.pdf",
    "size": 2048576,
    "created_at": "2026-02-04T10:30:00Z"
  }
}
```

---

### 3.4 Tus 协议上传（分片上传）

**基础端点**: `/api/tus`

**权限**: 需要认证

**说明**: 使用 [Tus Protocol](https://tus.io/) 实现断点续传，支持大文件上传。

#### 3.4.1 创建上传

**端点**: `POST /tus/upload`

**请求头**:
```http
Upload-Length: 10485760          # 文件总大小（字节）
Upload-Metadata: filename bXlfZmlsZS5wZGY=,filetype YXBwbGljYXRpb24vcGRm  # Base64 编码
Tus-Resumable: 1.0.0
```

**响应**:
```http
HTTP/1.1 201 Created
Location: /api/tus/upload/550e8400-e29b-41d4-a716-446655440000
Tus-Resumable: 1.0.0
```

#### 3.4.2 分片上传

**端点**: `PATCH /tus/upload/:upload_id`

**请求头**:
```http
Content-Type: application/offset+octet-stream
Content-Length: 1048576          # 本次上传分片大小
Upload-Offset: 0                 # 上传偏移量
Tus-Resumable: 1.0.0
```

**请求体**: 二进制文件数据

**响应**:
```http
HTTP/1.1 204 No Content
Upload-Offset: 1048576           # 已上传的总字节数
Tus-Resumable: 1.0.0
```

#### 3.4.3 查询上传进度

**端点**: `HEAD /tus/upload/:upload_id`

**响应**:
```http
HTTP/1.1 200 OK
Upload-Offset: 5242880           # 已上传的字节数
Upload-Length: 10485760          # 文件总大小
Tus-Resumable: 1.0.0
```

---

### 3.5 文件重命名

**端点**: `PATCH /files/:file_id`

**权限**: 需要认证（仅文件所有者）

**请求体**:
```json
{
  "filename": "new_filename.pdf"
}
```

**响应**:
```json
{
  "code": 0,
  "message": "File renamed successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "filename": "new_filename.pdf"
  }
}
```

---

### 3.6 删除文件（逻辑删除）

**端点**: `DELETE /files/:file_id`

**权限**: 需要认证（仅文件所有者）

**响应**:
```json
{
  "code": 0,
  "message": "File deleted successfully, will be permanently removed in 7 days"
}
```

**说明**: 文件进入回收倒计时（7 天），后台 GC 任务会在 7 天后物理删除。

---

### 3.7 批量删除文件

**端点**: `POST /files/batch-delete`

**权限**: 需要认证

**请求体**:
```json
{
  "file_ids": [
    "550e8400-e29b-41d4-a716-446655440000",
    "660e8400-e29b-41d4-a716-446655440001"
  ]
}
```

**响应**:
```json
{
  "code": 0,
  "message": "2 files deleted successfully",
  "data": {
    "deleted_count": 2
  }
}
```

---

### 3.8 下载文件

**端点**: `GET /files/:file_id/download`

**权限**: 需要认证（仅文件所有者）

**响应**: 文件流

**响应头**:
```http
Content-Type: application/octet-stream
Content-Disposition: attachment; filename="my_document.pdf"
Content-Length: 2048576
```

---

## 4. 分享管理接口

### 4.1 创建分享

**端点**: `POST /shares`

**权限**: 需要认证

**请求体**:
```json
{
  "file_ids": [
    "550e8400-e29b-41d4-a716-446655440000",
    "660e8400-e29b-41d4-a716-446655440001"
  ],
  "expires_in": 86400,           // 有效期（秒），1小时=3600, 24小时=86400, 7天=604800
  "max_downloads": 5,            // 最大下载次数，0=不限
  "password": "optional123"      // 访问密码（可选）
}
```

**响应**:
```json
{
  "code": 0,
  "message": "Share created successfully",
  "data": {
    "share_id": "770e8400-e29b-41d4-a716-446655440002",
    "pickup_code": "A2B3C4D5",
    "share_url": "https://your-domain.com/?code=A2B3C4D5",
    "expires_at": "2026-02-05T10:30:00Z",
    "max_downloads": 5,
    "has_password": true
  }
}
```

**错误码**:
- `4030`: 文件不存在或无权限
- `4031`: 超过最大分享数量限制
- `4032`: 有效期超过文件过期时间

---

### 4.2 获取我的分享列表

**端点**: `GET /shares`

**权限**: 需要认证

**查询参数**:
```
?page=1&page_size=20&status=active
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页数量，默认 20 |
| status | string | 否 | 状态筛选：active/expired/stopped |

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "share_id": "770e8400-e29b-41d4-a716-446655440002",
        "pickup_code": "A2B3C4D5",
        "file_count": 2,
        "created_at": "2026-02-04T10:30:00Z",
        "expires_at": "2026-02-05T10:30:00Z",
        "max_downloads": 5,
        "current_downloads": 2,
        "status": "active",  // active/expired/stopped
        "has_password": true
      }
    ],
    "total": 10,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 4.3 停止分享（Kill Link）

**端点**: `DELETE /shares/:share_id`

**权限**: 需要认证（仅分享创建者）

**响应**:
```json
{
  "code": 0,
  "message": "Share stopped successfully"
}
```

---

### 4.4 取件 - 获取分享信息（公开端点）

**端点**: `GET /pickup/:pickup_code`

**权限**: 公开

**查询参数**:
```
?password=optional123&captcha_token=turnstile_token
```

**响应**:

**情况 1: 需要密码**
```json
{
  "code": 4040,
  "message": "Password required",
  "data": {
    "requires_password": true
  }
}
```

**情况 2: 需要验证码**
```json
{
  "code": 4041,
  "message": "Captcha required",
  "data": {
    "requires_captcha": true,
    "reason": "Too many failed attempts"
  }
}
```

**情况 3: 成功获取文件列表**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "share_id": "770e8400-e29b-41d4-a716-446655440002",
    "files": [
      {
        "file_id": "550e8400-e29b-41d4-a716-446655440000",
        "filename": "vacation_photo.jpg",
        "size": 2048576,
        "mime_type": "image/jpeg"
      }
    ],
    "expires_at": "2026-02-05T10:30:00Z",
    "remaining_downloads": 3
  }
}
```

**错误码**:
- `4042`: 取件码不存在或已失效
- `4043`: 密码错误
- `4044`: 下载次数已用尽
- `4045`: 分享已过期

---

### 4.5 取件 - 下载文件

**端点**: `GET /pickup/:pickup_code/files/:file_id/download`

**权限**: 公开（需通过取件码验证）

**查询参数**:
```
?password=optional123  # 如果分享设置了密码
```

**响应**: 文件流

**响应头**:
```http
Content-Type: application/octet-stream
Content-Disposition: attachment; filename="vacation_photo.jpg"
Content-Length: 2048576
```

**说明**:
- 每次下载会增加 `current_downloads` 计数
- 当 `current_downloads >= max_downloads` 时，分享自动失效

---

### 4.6 取件 - 打包下载（ZIP）

**端点**: `GET /pickup/:pickup_code/download`

**权限**: 公开（需通过取件码验证）

**查询参数**:
```
?password=optional123  # 如果分享设置了密码
```

**响应**: ZIP 文件流

**响应头**:
```http
Content-Type: application/zip
Content-Disposition: attachment; filename="share_A2B3C4D5.zip"
Transfer-Encoding: chunked  # 流式打包
```

---

### 4.7 转存到我的文件柜

**端点**: `POST /pickup/:pickup_code/save`

**权限**: 需要认证

**请求体**:
```json
{
  "file_ids": [
    "550e8400-e29b-41d4-a716-446655440000"
  ],
  "password": "optional123"  // 可选，如果分享设置了密码
}
```

**响应**:
```json
{
  "code": 0,
  "message": "Files saved to your cabinet successfully",
  "data": {
    "saved_count": 1,
    "file_ids": [
      "880e8400-e29b-41d4-a716-446655440003"  // 新创建的文件 ID
    ]
  }
}
```

**说明**:
- 后端执行"逻辑复制"（事务内增加引用计数，创建新的 files_metadata 记录）
- 不会重复上传物理文件

---

## 5. 管理员接口

### 5.1 获取系统仪表盘

**端点**: `GET /admin/dashboard`

**权限**: 需要认证（仅管理员）

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "system": {
      "cpu_usage": 23.5,          // CPU 使用率 (%)
      "memory_usage": 1536.0,     // 内存使用 (MB)
      "memory_total": 4096.0,     // 内存总量 (MB)
      "disk_usage": 102400.0,     // 磁盘使用 (MB)
      "disk_total": 512000.0,     // 磁盘总量 (MB)
      "network_in": 1024.5,       // 入网流量 (KB/s)
      "network_out": 512.3        // 出网流量 (KB/s)
    },
    "storage": {
      "total_size": 104857600,    // 总存储大小 (Bytes)
      "used_size": 52428800,      // 已用大小 (Bytes)
      "file_count": 1234,         // 物理文件数量
      "ref_count_total": 5678     // 总引用次数
    },
    "business": {
      "total_users": 256,
      "active_users_today": 42,
      "total_files": 5678,
      "total_shares": 1234,
      "uploads_today": 89,
      "downloads_today": 156
    }
  }
}
```

---

### 5.2 获取全局文件列表

**端点**: `GET /admin/files`

**权限**: 需要认证（仅管理员）

**查询参数**:
```
?page=1&page_size=20&search=filename&user_id=uuid
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "user_id": "660e8400-e29b-41d4-a716-446655440001",
        "user_email": "user@example.com",
        "filename": "vacation_photo.jpg",
        "size": 2048576,
        "hash": "aabbccdd...",
        "is_banned": false,
        "ref_count": 3,
        "created_at": "2026-02-04T10:30:00Z"
      }
    ],
    "total": 5678,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 5.3 禁止文件分享

**端点**: `POST /admin/files/:hash/ban`

**权限**: 需要认证（仅管理员）

**请求体**:
```json
{
  "reason": "Violation of terms of service"
}
```

**响应**:
```json
{
  "code": 0,
  "message": "File banned successfully, all shares invalidated"
}
```

**说明**:
- 标记 `file_blobs.is_banned = true`
- 所有基于此 hash 的分享链接失效
- 禁止创建新的分享

---

### 5.4 物理删除文件

**端点**: `DELETE /admin/files/:hash`

**权限**: 需要认证（仅管理员）

**响应**:
```json
{
  "code": 0,
  "message": "File deleted permanently, all references removed"
}
```

**说明**:
- 从物理存储删除文件
- 级联删除所有用户的 `files_metadata` 记录
- 不可恢复，慎用！

---

### 5.5 系统设置 - 获取配置

**端点**: `GET /admin/settings`

**权限**: 需要认证（仅管理员）

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "registration_enabled": true,      // 是否开启注册
    "invite_code_required": true,      // 是否需要邀请码
    "max_file_size": 2147483648,       // 单文件大小限制 (2GB)
    "allowed_mime_types": [            // 允许的文件类型
      "image/*",
      "application/pdf",
      "application/zip"
    ],
    "storage_type": "local",           // 存储类型: local/s3
    "storage_quota_per_user": 10737418240  // 用户配额 (10GB)
  }
}
```

---

### 5.6 系统设置 - 更新配置

**端点**: `PATCH /admin/settings`

**权限**: 需要认证（仅管理员）

**请求体**:
```json
{
  "registration_enabled": false,
  "invite_code_required": true
}
```

**响应**:
```json
{
  "code": 0,
  "message": "Settings updated successfully"
}
```

---

### 5.7 用户管理 - 获取用户列表

**端点**: `GET /admin/users`

**权限**: 需要认证（仅管理员）

**查询参数**:
```
?page=1&page_size=20&search=email&status=active
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "user_id": "550e8400-e29b-41d4-a716-446655440000",
        "email": "user@example.com",
        "role": "user",
        "status": "active",     // active/disabled
        "file_count": 42,
        "total_size": 104857600,
        "created_at": "2026-01-01T00:00:00Z",
        "last_login_at": "2026-02-04T09:00:00Z"
      }
    ],
    "total": 256,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 5.8 用户管理 - 禁用/启用用户

**端点**: `PATCH /admin/users/:user_id`

**权限**: 需要认证（仅管理员）

**请求体**:
```json
{
  "status": "disabled",  // active/disabled
  "reason": "Abuse detected"
}
```

**响应**:
```json
{
  "code": 0,
  "message": "User status updated successfully"
}
```

---

## 6. 错误码说明

### 6.1 通用错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1000 | 未知错误 |
| 1001 | 请求参数错误 |
| 1002 | 请求体解析失败 |
| 1003 | 数据库错误 |
| 1004 | 服务器内部错误 |

### 6.2 认证错误码 (4000-4099)

| 错误码 | 说明 |
|--------|------|
| 4001 | 邮箱已注册 |
| 4002 | 邀请码无效 |
| 4003 | 注册功能已关闭 |
| 4010 | 邮箱或密码错误 |
| 4011 | 账户已被禁用 |
| 4012 | 需要人机验证 |
| 4013 | 验证码校验失败 |
| 4020 | Token 无效或已过期 |
| 4021 | Token 缺失 |
| 4022 | 权限不足 |

### 6.3 文件错误码 (4100-4199)

| 错误码 | 说明 |
|--------|------|
| 4100 | 文件不存在 |
| 4101 | 文件已被删除 |
| 4102 | 文件超过大小限制 |
| 4103 | 文件类型不允许 |
| 4104 | 用户存储配额已满 |
| 4105 | 文件哈希校验失败 |
| 4106 | 文件已被管理员禁止 |

### 6.4 分享错误码 (4200-4299)

| 错误码 | 说明 |
|--------|------|
| 4030 | 文件不存在或无权限 |
| 4031 | 超过最大分享数量限制 |
| 4032 | 有效期超过文件过期时间 |
| 4040 | 需要密码 |
| 4041 | 需要验证码 |
| 4042 | 取件码不存在或已失效 |
| 4043 | 密码错误 |
| 4044 | 下载次数已用尽 |
| 4045 | 分享已过期 |

### 6.5 管理员错误码 (4300-4399)

| 错误码 | 说明 |
|--------|------|
| 4300 | 非管理员用户 |
| 4301 | 操作被拒绝 |

---

## 📌 附录

### A. 文件类型映射

```json
{
  "image": ["image/jpeg", "image/png", "image/gif", "image/webp"],
  "video": ["video/mp4", "video/mpeg", "video/quicktime"],
  "document": ["application/pdf", "application/msword", "text/plain"],
  "archive": ["application/zip", "application/x-rar-compressed"]
}
```

### B. Magic Bytes 校验规则

后端会在接收首个分片时校验文件头部（Magic Bytes），防止伪装文件类型：

| 文件类型 | Magic Bytes (HEX) |
|----------|-------------------|
| JPEG | `FF D8 FF` |
| PNG | `89 50 4E 47` |
| PDF | `25 50 44 46` |
| ZIP | `50 4B 03 04` |

### C. 限流策略

| 操作 | 限制 | 窗口期 |
|------|------|--------|
| 登录失败 | 5 次 | 1 分钟 |
| 取件码错误 | 5 次 | 1 分钟 |
| 文件上传 | 100 个 | 1 小时 |
| 创建分享 | 50 个 | 1 小时 |

---

**文档维护**: 本文档应与代码实现保持同步，任何 API 变更必须同步更新此文档。

**最后审核**: 2026-02-04
