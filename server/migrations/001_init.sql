-- AhaVault Database Schema Initialization
-- Version: 1.0.0
-- Created: 2026-02-04
-- Description: 初始化数据库表结构，包括用户、文件、分享等核心表

-- ==========================================
-- 1. 启用 UUID 扩展
-- ==========================================
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ==========================================
-- 2. 用户表 (users)
-- ==========================================
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user' NOT NULL,
    status VARCHAR(50) DEFAULT 'active' NOT NULL,

    -- 存储配额 (bytes)
    storage_quota BIGINT DEFAULT 10737418240 NOT NULL,  -- 默认 10GB
    storage_used BIGINT DEFAULT 0 NOT NULL,

    -- 时间戳
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW() NOT NULL,
    last_login_at TIMESTAMP,

    -- 约束
    CONSTRAINT chk_role CHECK (role IN ('user', 'admin')),
    CONSTRAINT chk_status CHECK (status IN ('active', 'disabled')),
    CONSTRAINT chk_storage_used CHECK (storage_used >= 0),
    CONSTRAINT chk_storage_quota CHECK (storage_quota > 0)
);

-- 用户表索引
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_created_at ON users(created_at);

-- 用户表注释
COMMENT ON TABLE users IS '用户账户表';
COMMENT ON COLUMN users.role IS '用户角色：user=普通用户, admin=管理员';
COMMENT ON COLUMN users.status IS '账户状态：active=正常, disabled=已禁用';
COMMENT ON COLUMN users.storage_quota IS '存储配额（字节），默认 10GB';
COMMENT ON COLUMN users.storage_used IS '已使用存储（字节）';

-- ==========================================
-- 3. 物理文件表 (file_blobs)
-- ==========================================
CREATE TABLE IF NOT EXISTS file_blobs (
    hash VARCHAR(64) PRIMARY KEY,  -- SHA-256 哈希值（64 字符 HEX）
    store_path VARCHAR(255) NOT NULL,  -- 物理存储路径
    encrypted_dek TEXT NOT NULL,  -- 加密后的 DEK (Base64)
    size BIGINT NOT NULL,  -- 文件大小（字节）
    mime_type VARCHAR(128),  -- MIME 类型

    -- 引用计数（核心字段）
    ref_count INT DEFAULT 1 NOT NULL,

    -- 管理字段
    is_banned BOOLEAN DEFAULT FALSE NOT NULL,  -- 是否被管理员禁止
    ban_reason TEXT,  -- 禁止原因

    -- 时间戳
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW() NOT NULL,

    -- 约束
    CONSTRAINT chk_ref_count CHECK (ref_count >= 0),
    CONSTRAINT chk_size CHECK (size > 0)
);

-- 物理文件表索引
CREATE INDEX idx_file_blobs_ref_count ON file_blobs(ref_count);
CREATE INDEX idx_file_blobs_is_banned ON file_blobs(is_banned);
CREATE INDEX idx_file_blobs_created_at ON file_blobs(created_at);
CREATE INDEX idx_file_blobs_size ON file_blobs(size);

-- 物理文件表注释
COMMENT ON TABLE file_blobs IS '物理文件存储表（CAS 内容寻址存储）';
COMMENT ON COLUMN file_blobs.hash IS 'SHA-256 文件哈希值（唯一标识）';
COMMENT ON COLUMN file_blobs.encrypted_dek IS '使用 KEK 加密后的 DEK（信封加密）';
COMMENT ON COLUMN file_blobs.ref_count IS '引用计数，当归零时由 GC 清理';
COMMENT ON COLUMN file_blobs.is_banned IS '管理员禁止标记，禁止后无法创建新分享';

-- ==========================================
-- 4. 逻辑文件表 (files_metadata)
-- ==========================================
CREATE TABLE IF NOT EXISTS files_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_blob_hash VARCHAR(64) NOT NULL REFERENCES file_blobs(hash) ON DELETE RESTRICT,

    -- 用户自定义字段
    filename VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL,  -- 冗余存储，方便查询

    -- 生命周期管理
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    expires_at TIMESTAMP,  -- 文件过期时间（可选）
    deleted_at TIMESTAMP,  -- 软删除时间

    -- 约束
    CONSTRAINT chk_filename_length CHECK (char_length(filename) > 0),
    CONSTRAINT chk_size_positive CHECK (size > 0)
);

-- 逻辑文件表索引
CREATE INDEX idx_files_metadata_user_id ON files_metadata(user_id);
CREATE INDEX idx_files_metadata_blob_hash ON files_metadata(file_blob_hash);
CREATE INDEX idx_files_metadata_deleted_at ON files_metadata(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_files_metadata_expires_at ON files_metadata(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_files_metadata_created_at ON files_metadata(created_at);
CREATE INDEX idx_files_metadata_user_active ON files_metadata(user_id, deleted_at) WHERE deleted_at IS NULL;

-- 逻辑文件表注释
COMMENT ON TABLE files_metadata IS '用户文件元数据表（逻辑文件）';
COMMENT ON COLUMN files_metadata.file_blob_hash IS '关联的物理文件哈希';
COMMENT ON COLUMN files_metadata.filename IS '用户自定义文件名';
COMMENT ON COLUMN files_metadata.deleted_at IS '软删除时间，7 天后由 GC 物理删除';
COMMENT ON COLUMN files_metadata.expires_at IS '文件过期时间（可选），过期后自动删除';

-- ==========================================
-- 5. 分享会话表 (share_sessions)
-- ==========================================
CREATE TABLE IF NOT EXISTS share_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pickup_code VARCHAR(8) UNIQUE NOT NULL,  -- 取件码（8 位）
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- 访问控制
    password_hash VARCHAR(255),  -- 访问密码（可选）
    max_downloads INT DEFAULT 0 NOT NULL,  -- 最大下载次数，0=不限
    current_downloads INT DEFAULT 0 NOT NULL,  -- 当前下载次数

    -- 生命周期
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    expires_at TIMESTAMP NOT NULL,  -- 过期时间（必填）
    stopped_at TIMESTAMP,  -- 手动停止时间

    -- 约束
    CONSTRAINT chk_pickup_code_format CHECK (pickup_code ~ '^[2-9A-Z]{8}$'),
    CONSTRAINT chk_max_downloads CHECK (max_downloads >= 0),
    CONSTRAINT chk_current_downloads CHECK (current_downloads >= 0),
    CONSTRAINT chk_expires_at_future CHECK (expires_at > created_at)
);

-- 分享会话表索引
CREATE INDEX idx_share_sessions_pickup_code ON share_sessions(pickup_code);
CREATE INDEX idx_share_sessions_creator_id ON share_sessions(creator_id);
CREATE INDEX idx_share_sessions_expires_at ON share_sessions(expires_at);
CREATE INDEX idx_share_sessions_created_at ON share_sessions(created_at);
CREATE INDEX idx_share_sessions_active ON share_sessions(expires_at, stopped_at) WHERE stopped_at IS NULL;

-- 分享会话表注释
COMMENT ON TABLE share_sessions IS '分享会话表（取件码系统）';
COMMENT ON COLUMN share_sessions.pickup_code IS '8 位取件码，字符集 [2-9A-Z]，排除易混淆字符';
COMMENT ON COLUMN share_sessions.max_downloads IS '最大下载次数，0 表示不限制';
COMMENT ON COLUMN share_sessions.stopped_at IS '手动停止分享的时间（Kill Link）';

-- ==========================================
-- 6. 分享文件关联表 (share_files)
-- ==========================================
CREATE TABLE IF NOT EXISTS share_files (
    share_id UUID NOT NULL REFERENCES share_sessions(id) ON DELETE CASCADE,
    file_id UUID NOT NULL REFERENCES files_metadata(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,

    -- 联合主键
    PRIMARY KEY (share_id, file_id)
);

-- 分享文件关联表索引
CREATE INDEX idx_share_files_share_id ON share_files(share_id);
CREATE INDEX idx_share_files_file_id ON share_files(file_id);

-- 分享文件关联表注释
COMMENT ON TABLE share_files IS '分享会话与文件的多对多关联表';

-- ==========================================
-- 7. 上传会话表 (upload_sessions) - 用于 Tus 协议
-- ==========================================
CREATE TABLE IF NOT EXISTS upload_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- 上传状态
    status VARCHAR(50) DEFAULT 'uploading' NOT NULL,  -- uploading/completed/failed
    upload_offset BIGINT DEFAULT 0 NOT NULL,  -- 已上传字节数
    upload_length BIGINT NOT NULL,  -- 文件总大小

    -- 文件信息
    filename VARCHAR(255) NOT NULL,
    mime_type VARCHAR(128),
    hash VARCHAR(64),  -- 前端计算的哈希（上传完成后验证）

    -- 临时存储路径
    temp_path VARCHAR(255),

    -- 时间戳
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW() NOT NULL,
    completed_at TIMESTAMP,

    -- 约束
    CONSTRAINT chk_upload_status CHECK (status IN ('uploading', 'completed', 'failed')),
    CONSTRAINT chk_upload_offset CHECK (upload_offset >= 0),
    CONSTRAINT chk_upload_length CHECK (upload_length > 0),
    CONSTRAINT chk_offset_le_length CHECK (upload_offset <= upload_length)
);

-- 上传会话表索引
CREATE INDEX idx_upload_sessions_user_id ON upload_sessions(user_id);
CREATE INDEX idx_upload_sessions_status ON upload_sessions(status);
CREATE INDEX idx_upload_sessions_updated_at ON upload_sessions(updated_at);

-- 上传会话表注释
COMMENT ON TABLE upload_sessions IS 'Tus 协议上传会话表（支持断点续传）';
COMMENT ON COLUMN upload_sessions.upload_offset IS '已上传的字节数（断点续传的关键）';
COMMENT ON COLUMN upload_sessions.hash IS '前端计算的 SHA-256 哈希，上传完成后验证';

-- ==========================================
-- 8. 系统配置表 (system_settings)
-- ==========================================
CREATE TABLE IF NOT EXISTS system_settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT NOW() NOT NULL
);

-- 插入默认配置
INSERT INTO system_settings (key, value, description) VALUES
    ('registration_enabled', 'true', '是否开启用户注册'),
    ('invite_code_required', 'true', '注册是否需要邀请码'),
    ('max_file_size', '2147483648', '单文件大小限制（2GB）'),
    ('storage_type', 'local', '存储引擎类型：local/s3'),
    ('default_user_quota', '10737418240', '新用户默认配额（10GB）'),
    ('share_code_length', '8', '取件码长度'),
    ('gc_retention_days', '7', '软删除保留天数')
ON CONFLICT (key) DO NOTHING;

-- 系统配置表注释
COMMENT ON TABLE system_settings IS '系统配置表（Key-Value 存储）';

-- ==========================================
-- 9. 审计日志表 (audit_logs) - 可选
-- ==========================================
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,  -- 操作类型
    resource_type VARCHAR(50),  -- 资源类型：file/share/user
    resource_id VARCHAR(100),  -- 资源 ID
    ip_address INET,  -- 客户端 IP
    user_agent TEXT,  -- User-Agent
    details JSONB,  -- 详细信息（JSON 格式）
    created_at TIMESTAMP DEFAULT NOW() NOT NULL
);

-- 审计日志表索引
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);

-- 审计日志表注释
COMMENT ON TABLE audit_logs IS '操作审计日志表';
COMMENT ON COLUMN audit_logs.action IS '操作类型：login/upload/download/share/delete 等';

-- ==========================================
-- 10. 创建更新时间自动更新触发器
-- ==========================================

-- 创建通用的更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 为需要的表添加触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_file_blobs_updated_at BEFORE UPDATE ON file_blobs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_upload_sessions_updated_at BEFORE UPDATE ON upload_sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_system_settings_updated_at BEFORE UPDATE ON system_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ==========================================
-- 11. 创建视图 - 用户存储统计
-- ==========================================
CREATE OR REPLACE VIEW user_storage_stats AS
SELECT
    u.id AS user_id,
    u.email,
    u.storage_quota,
    u.storage_used,
    COUNT(fm.id) AS file_count,
    COALESCE(SUM(fm.size), 0) AS actual_storage_used,
    u.storage_quota - u.storage_used AS storage_available
FROM users u
LEFT JOIN files_metadata fm ON fm.user_id = u.id AND fm.deleted_at IS NULL
GROUP BY u.id, u.email, u.storage_quota, u.storage_used;

COMMENT ON VIEW user_storage_stats IS '用户存储统计视图';

-- ==========================================
-- 12. 创建视图 - 活跃分享统计
-- ==========================================
CREATE OR REPLACE VIEW active_shares AS
SELECT
    ss.id,
    ss.pickup_code,
    ss.creator_id,
    u.email AS creator_email,
    ss.created_at,
    ss.expires_at,
    ss.max_downloads,
    ss.current_downloads,
    COUNT(sf.file_id) AS file_count,
    CASE
        WHEN ss.stopped_at IS NOT NULL THEN 'stopped'
        WHEN ss.expires_at < NOW() THEN 'expired'
        WHEN ss.max_downloads > 0 AND ss.current_downloads >= ss.max_downloads THEN 'exhausted'
        ELSE 'active'
    END AS status
FROM share_sessions ss
JOIN users u ON u.id = ss.creator_id
LEFT JOIN share_files sf ON sf.share_id = ss.id
GROUP BY ss.id, ss.pickup_code, ss.creator_id, u.email, ss.created_at, ss.expires_at, ss.max_downloads, ss.current_downloads, ss.stopped_at;

COMMENT ON VIEW active_shares IS '活跃分享统计视图';

-- ==========================================
-- 完成初始化
-- ==========================================

-- 显示所有表
SELECT table_name, table_type
FROM information_schema.tables
WHERE table_schema = 'public'
ORDER BY table_name;

-- 显示统计信息
SELECT
    'Tables Created' AS metric,
    COUNT(*) AS count
FROM information_schema.tables
WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
UNION ALL
SELECT
    'Views Created' AS metric,
    COUNT(*) AS count
FROM information_schema.views
WHERE table_schema = 'public'
UNION ALL
SELECT
    'Indexes Created' AS metric,
    COUNT(*) AS count
FROM pg_indexes
WHERE schemaname = 'public';
