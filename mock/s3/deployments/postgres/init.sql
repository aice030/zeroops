-- Mock S3 Metadata Database Initialization Script

-- 创建元数据表 (匹配 models.Metadata)
CREATE TABLE IF NOT EXISTS metadata (
    bucket VARCHAR(255) NOT NULL,
    key VARCHAR(1024) NOT NULL,
    size BIGINT NOT NULL CHECK (size >= 0),
    content_type VARCHAR(255) NOT NULL,
    md5_hash CHAR(32) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'deleted', 'corrupted')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    PRIMARY KEY (bucket, key)
);

-- 创建索引以提升查询性能
CREATE INDEX IF NOT EXISTS idx_metadata_bucket ON metadata(bucket);
CREATE INDEX IF NOT EXISTS idx_metadata_status ON metadata(status);
CREATE INDEX IF NOT EXISTS idx_metadata_created_at ON metadata(created_at);

-- 权限设置
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO admin;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO admin;

COMMIT;