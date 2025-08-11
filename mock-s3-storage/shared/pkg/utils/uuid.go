package utils

import (
	"strings"

	"github.com/google/uuid"
)

// GenerateUUID 生成新的UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateUUIDWithoutHyphens 生成不带连字符的UUID
func GenerateUUIDWithoutHyphens() string {
	id := uuid.New()
	uuidStr := id.String()
	// 使用strings.Builder提升性能
	var builder strings.Builder
	builder.Grow(32) // 预分配容量
	for _, r := range uuidStr {
		if r != '-' {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

// IsValidUUID 验证UUID格式是否有效
func IsValidUUID(uuidStr string) bool {
	_, err := uuid.Parse(uuidStr)
	return err == nil
}

// ParseUUID 解析UUID字符串
func ParseUUID(uuidStr string) (uuid.UUID, error) {
	return uuid.Parse(uuidStr)
}
