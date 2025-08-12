package database

import (
	"context"
	"database/sql"
	"shared/config"
	"time"
)

// Database 数据库接口
type Database interface {
	// Connect 连接数据库
	Connect(ctx context.Context) error

	// Close 关闭数据库连接
	Close() error

	// Ping 测试数据库连接
	Ping(ctx context.Context) error
}

// SQLDatabase SQL数据库接口（PostgreSQL）
type SQLDatabase interface {
	Database

	// Query 执行查询
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)

	// QueryRow 执行查询单行
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row

	// Exec 执行SQL语句
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)

	// BeginTx 开始事务
	BeginTx(ctx context.Context) (*sql.Tx, error)
}

// RedisDatabase Redis数据库接口
type RedisDatabase interface {
	Database

	// Get 获取值
	Get(ctx context.Context, key string) (string, error)

	// Set 设置值
	Set(ctx context.Context, key string, value any, expiration time.Duration) error

	// Delete 删除键
	Delete(ctx context.Context, keys ...string) error

	// HGet 获取哈希字段值
	HGet(ctx context.Context, key, field string) (string, error)

	// HSet 设置哈希字段值
	HSet(ctx context.Context, key string, values ...any) error

	// HGetAll 获取哈希所有字段
	HGetAll(ctx context.Context, key string) (map[string]string, error)
}

// NewSQL 创建SQL数据库连接
func NewSQL(config config.DatabaseConfig) (SQLDatabase, error) {
	// TODO: 实现PostgreSQL连接
	return nil, nil
}

// NewRedis 创建Redis数据库连接
func NewRedis(config config.DatabaseConfig) (RedisDatabase, error) {
	// TODO: 实现Redis连接
	return nil, nil
}
