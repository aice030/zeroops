package repository

import (
	"context"
	"database/sql"
	"fmt"
	"mocks3/shared/models"

	_ "github.com/lib/pq"
)

// PostgreSQLRepository PostgreSQL仓库
type PostgreSQLRepository struct {
	db *sql.DB
}

// NewPostgreSQLRepository 创建PostgreSQL仓库
func NewPostgreSQLRepository(dsn string) (*PostgreSQLRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	repo := &PostgreSQLRepository{db: db}
	if err := repo.initTables(); err != nil {
		return nil, fmt.Errorf("failed to init tables: %w", err)
	}

	return repo, nil
}

// initTables 初始化数据库表
func (r *PostgreSQLRepository) initTables() error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS metadata (
		bucket VARCHAR(255) NOT NULL,
		key VARCHAR(1024) NOT NULL,
		size BIGINT NOT NULL CHECK (size >= 0),
		content_type VARCHAR(255) NOT NULL,
		md5_hash CHAR(32) NOT NULL,
		status VARCHAR(20) NOT NULL DEFAULT 'active',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		
		PRIMARY KEY (bucket, key)
	);

	CREATE INDEX IF NOT EXISTS idx_metadata_bucket ON metadata(bucket);
	CREATE INDEX IF NOT EXISTS idx_metadata_status ON metadata(status);
	`
	_, err := r.db.Exec(createTableSQL)
	return err
}

// Create 创建元数据记录
func (r *PostgreSQLRepository) Create(ctx context.Context, metadata *models.Metadata) error {
	query := `
		INSERT INTO metadata (bucket, key, size, content_type, md5_hash, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (bucket, key) 
		DO UPDATE SET 
			size = EXCLUDED.size,
			content_type = EXCLUDED.content_type,
			md5_hash = EXCLUDED.md5_hash,
			status = EXCLUDED.status,
			created_at = EXCLUDED.created_at
	`

	_, err := r.db.ExecContext(ctx, query,
		metadata.Bucket, metadata.Key, metadata.Size,
		metadata.ContentType, metadata.MD5Hash, metadata.Status,
		metadata.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create metadata: %w", err)
	}
	return nil
}

// GetByKey 根据键获取元数据
func (r *PostgreSQLRepository) GetByKey(ctx context.Context, bucket, key string) (*models.Metadata, error) {
	query := `
		SELECT bucket, key, size, content_type, md5_hash, status, created_at
		FROM metadata 
		WHERE bucket = $1 AND key = $2 AND status = 'active'
	`

	var metadata models.Metadata
	err := r.db.QueryRowContext(ctx, query, bucket, key).Scan(
		&metadata.Bucket, &metadata.Key, &metadata.Size,
		&metadata.ContentType, &metadata.MD5Hash, &metadata.Status,
		&metadata.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("metadata not found: %s/%s", bucket, key)
		}
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	return &metadata, nil
}

// Update 更新元数据
func (r *PostgreSQLRepository) Update(ctx context.Context, metadata *models.Metadata) error {
	query := `
		UPDATE metadata 
		SET size = $3, content_type = $4, md5_hash = $5, status = $6
		WHERE bucket = $1 AND key = $2
	`

	result, err := r.db.ExecContext(ctx, query,
		metadata.Bucket, metadata.Key, metadata.Size,
		metadata.ContentType, metadata.MD5Hash, metadata.Status)

	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("metadata not found for update: %s/%s", metadata.Bucket, metadata.Key)
	}

	return nil
}

// Delete 删除元数据（软删除）
func (r *PostgreSQLRepository) Delete(ctx context.Context, bucket, key string) error {
	query := `UPDATE metadata SET status = 'deleted' WHERE bucket = $1 AND key = $2`

	result, err := r.db.ExecContext(ctx, query, bucket, key)
	if err != nil {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("metadata not found for delete: %s/%s", bucket, key)
	}

	return nil
}

// List 列出元数据
func (r *PostgreSQLRepository) List(ctx context.Context, bucket, prefix string, limit, offset int) ([]*models.Metadata, error) {
	var query string
	var args []any

	if prefix != "" {
		query = `
			SELECT bucket, key, size, content_type, md5_hash, status, created_at
			FROM metadata 
			WHERE bucket = $1 AND key LIKE $2 AND status = 'active'
			ORDER BY key
			LIMIT $3 OFFSET $4
		`
		args = []any{bucket, prefix + "%", limit, offset}
	} else {
		query = `
			SELECT bucket, key, size, content_type, md5_hash, status, created_at
			FROM metadata 
			WHERE bucket = $1 AND status = 'active'
			ORDER BY key
			LIMIT $2 OFFSET $3
		`
		args = []any{bucket, limit, offset}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list metadata: %w", err)
	}
	defer rows.Close()

	var metadataList []*models.Metadata
	for rows.Next() {
		var metadata models.Metadata
		err := rows.Scan(
			&metadata.Bucket, &metadata.Key, &metadata.Size,
			&metadata.ContentType, &metadata.MD5Hash, &metadata.Status,
			&metadata.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metadata row: %w", err)
		}
		metadataList = append(metadataList, &metadata)
	}

	return metadataList, nil
}

// Search 搜索元数据（简单的LIKE查询）
func (r *PostgreSQLRepository) Search(ctx context.Context, query string, limit int) ([]*models.Metadata, error) {
	searchSQL := `
		SELECT bucket, key, size, content_type, md5_hash, status, created_at
		FROM metadata 
		WHERE key LIKE $1 AND status = 'active'
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, searchSQL, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search metadata: %w", err)
	}
	defer rows.Close()

	var metadataList []*models.Metadata
	for rows.Next() {
		var metadata models.Metadata
		err := rows.Scan(
			&metadata.Bucket, &metadata.Key, &metadata.Size,
			&metadata.ContentType, &metadata.MD5Hash, &metadata.Status,
			&metadata.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}
		metadataList = append(metadataList, &metadata)
	}

	return metadataList, nil
}

// GetStats 获取统计信息
func (r *PostgreSQLRepository) GetStats(ctx context.Context) (*models.Stats, error) {
	query := `
		SELECT 
			COUNT(*) as total_objects,
			COALESCE(SUM(size), 0) as total_size,
			COALESCE(MAX(created_at), NOW()) as last_updated
		FROM metadata 
		WHERE status = 'active'
	`

	var stats models.Stats
	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalObjects, &stats.TotalSize, &stats.LastUpdated)

	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return &stats, nil
}

// Close 关闭数据库连接
func (r *PostgreSQLRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
