package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
)

// CalculateMD5 计算数据的MD5哈希
func CalculateMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// CalculateSHA1 计算数据的SHA1哈希
func CalculateSHA1(data []byte) string {
	hash := sha1.Sum(data)
	return hex.EncodeToString(hash[:])
}

// CalculateSHA256 计算数据的SHA256哈希
func CalculateSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// VerifyMD5 验证数据的MD5哈希
func VerifyMD5(data []byte, expectedHash string) bool {
	actualHash := CalculateMD5(data)
	return actualHash == expectedHash
}

// VerifySHA256 验证数据的SHA256哈希
func VerifySHA256(data []byte, expectedHash string) bool {
	actualHash := CalculateSHA256(data)
	return actualHash == expectedHash
}