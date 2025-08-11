package utils

import (
	"net"
	"regexp"
	"strings"
	"unicode"
)

// 预编译正则表达式提升性能
var (
	bucketNameRegex = regexp.MustCompile(`^[a-z0-9.\-]+$`)
	emailRegex      = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	urlRegex        = regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(:\d+)?(/.*)?$`)
)

// IsValidBucketName 验证S3存储桶名称是否有效
func IsValidBucketName(bucketName string) bool {
	if len(bucketName) < 3 || len(bucketName) > 63 {
		return false
	}

	// 不能以点或连字符开头或结尾
	if strings.HasPrefix(bucketName, ".") || strings.HasPrefix(bucketName, "-") ||
		strings.HasSuffix(bucketName, ".") || strings.HasSuffix(bucketName, "-") {
		return false
	}

	// 只能包含小写字母、数字、点和连字符
	if !bucketNameRegex.MatchString(bucketName) {
		return false
	}

	// 不能包含连续的点
	if strings.Contains(bucketName, "..") {
		return false
	}

	return true
}

// IsValidObjectKey 验证S3对象键名是否有效
func IsValidObjectKey(objectKey string) bool {
	if len(objectKey) == 0 || len(objectKey) > 1024 {
		return false
	}

	// 不能以斜杠开头
	if strings.HasPrefix(objectKey, "/") {
		return false
	}

	return true
}

// IsValidEmail 验证邮箱格式
func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// IsValidPassword 验证密码强度
func IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
		// 提前退出优化：所有条件都满足时直接返回
		if hasUpper && hasLower && hasDigit && hasSpecial {
			return true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

// IsValidPort 验证端口号是否有效
func IsValidPort(port int) bool {
	return port > 0 && port <= 65535
}

// IsValidIPv4 验证IPv4地址格式
func IsValidIPv4(ip string) bool {
	// 使用标准库net.ParseIP来验证IPv4地址，性能更高且更准确
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	// 检查是否为IPv4地址（不包含IPv6）
	return parsedIP.To4() != nil
}

// IsValidURL 验证URL格式
func IsValidURL(url string) bool {
	return urlRegex.MatchString(url)
}

// SanitizeString 清理字符串，移除危险字符
func SanitizeString(input string) string {
	// 移除控制字符
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, input)

	// 去除首尾空白
	return strings.TrimSpace(cleaned)
}