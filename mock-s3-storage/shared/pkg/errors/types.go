package errors

import (
	"fmt"
	"net/http"
	"time"
)

// ErrorType 错误类型枚举
type ErrorType string

const (
	// 客户端错误类型
	ErrorTypeBadRequest          ErrorType = "BAD_REQUEST"
	ErrorTypeUnauthorized        ErrorType = "UNAUTHORIZED"
	ErrorTypeForbidden           ErrorType = "FORBIDDEN"
	ErrorTypeNotFound            ErrorType = "NOT_FOUND"
	ErrorTypeConflict            ErrorType = "CONFLICT"
	ErrorTypeValidation          ErrorType = "VALIDATION"
	ErrorTypeRateLimit           ErrorType = "RATE_LIMIT"
	
	// 服务端错误类型
	ErrorTypeInternal            ErrorType = "INTERNAL_ERROR"
	ErrorTypeServiceUnavailable  ErrorType = "SERVICE_UNAVAILABLE"
	ErrorTypeTimeout             ErrorType = "TIMEOUT"
	ErrorTypeDatabase            ErrorType = "DATABASE_ERROR"
	ErrorTypeStorage             ErrorType = "STORAGE_ERROR"
	ErrorTypeNetwork             ErrorType = "NETWORK_ERROR"
	ErrorTypeThirdParty          ErrorType = "THIRD_PARTY_ERROR"
	
	// 业务逻辑错误
	ErrorTypeBucketExists        ErrorType = "BUCKET_ALREADY_EXISTS"
	ErrorTypeBucketNotFound      ErrorType = "BUCKET_NOT_FOUND"
	ErrorTypeObjectNotFound      ErrorType = "OBJECT_NOT_FOUND"
	ErrorTypeInsufficientStorage ErrorType = "INSUFFICIENT_STORAGE"
	ErrorTypeQuotaExceeded       ErrorType = "QUOTA_EXCEEDED"
)

// ErrorSeverity 错误严重程度
type ErrorSeverity string

const (
	SeverityInfo     ErrorSeverity = "INFO"
	SeverityWarning  ErrorSeverity = "WARNING"
	SeverityError    ErrorSeverity = "ERROR"
	SeverityCritical ErrorSeverity = "CRITICAL"
)

// AppError 应用程序统一错误结构
type AppError struct {
	Type       ErrorType                `json:"type"`
	Code       string                   `json:"code"`
	Message    string                   `json:"message"`
	Details    string                   `json:"details,omitempty"`
	Severity   ErrorSeverity            `json:"severity"`
	Timestamp  time.Time                `json:"timestamp"`
	RequestID  string                   `json:"request_id,omitempty"`
	Service    string                   `json:"service,omitempty"`
	Operation  string                   `json:"operation,omitempty"`
	Metadata   map[string]any           `json:"metadata,omitempty"`
	Cause      error                    `json:"-"` // 原始错误，不序列化
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 支持Go 1.13+的错误链
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Is 支持errors.Is()判断
func (e *AppError) Is(target error) bool {
	if appErr, ok := target.(*AppError); ok {
		return e.Type == appErr.Type && e.Code == appErr.Code
	}
	return false
}

// HTTPStatusCode 返回对应的HTTP状态码
func (e *AppError) HTTPStatusCode() int {
	switch e.Type {
	case ErrorTypeBadRequest, ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeUnauthorized:
		return http.StatusUnauthorized
	case ErrorTypeForbidden:
		return http.StatusForbidden
	case ErrorTypeNotFound, ErrorTypeBucketNotFound, ErrorTypeObjectNotFound:
		return http.StatusNotFound
	case ErrorTypeConflict, ErrorTypeBucketExists:
		return http.StatusConflict
	case ErrorTypeRateLimit:
		return http.StatusTooManyRequests
	case ErrorTypeServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrorTypeTimeout:
		return http.StatusRequestTimeout
	case ErrorTypeInsufficientStorage:
		return http.StatusInsufficientStorage
	case ErrorTypeQuotaExceeded:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// ErrorBuilder 错误构建器，支持链式调用
type ErrorBuilder struct {
	err *AppError
}

// NewError 创建新的错误构建器
func NewError(errorType ErrorType, code, message string) *ErrorBuilder {
	return &ErrorBuilder{
		err: &AppError{
			Type:      errorType,
			Code:      code,
			Message:   message,
			Severity:  SeverityError, // 默认错误级别
			Timestamp: time.Now(),
			Metadata:  make(map[string]any),
		},
	}
}

// WithDetails 添加详细信息
func (b *ErrorBuilder) WithDetails(details string) *ErrorBuilder {
	b.err.Details = details
	return b
}

// WithSeverity 设置严重程度
func (b *ErrorBuilder) WithSeverity(severity ErrorSeverity) *ErrorBuilder {
	b.err.Severity = severity
	return b
}

// WithRequestID 设置请求ID
func (b *ErrorBuilder) WithRequestID(requestID string) *ErrorBuilder {
	b.err.RequestID = requestID
	return b
}

// WithService 设置服务名称
func (b *ErrorBuilder) WithService(service string) *ErrorBuilder {
	b.err.Service = service
	return b
}

// WithOperation 设置操作名称
func (b *ErrorBuilder) WithOperation(operation string) *ErrorBuilder {
	b.err.Operation = operation
	return b
}

// WithMetadata 添加元数据
func (b *ErrorBuilder) WithMetadata(key string, value any) *ErrorBuilder {
	b.err.Metadata[key] = value
	return b
}

// WithCause 设置原始错误
func (b *ErrorBuilder) WithCause(cause error) *ErrorBuilder {
	b.err.Cause = cause
	return b
}

// Build 构建最终的错误对象
func (b *ErrorBuilder) Build() *AppError {
	return b.err
}

// 预定义的常见错误构建函数
func BadRequest(code, message string) *ErrorBuilder {
	return NewError(ErrorTypeBadRequest, code, message).WithSeverity(SeverityWarning)
}

func NotFound(code, message string) *ErrorBuilder {
	return NewError(ErrorTypeNotFound, code, message).WithSeverity(SeverityInfo)
}

func Internal(code, message string) *ErrorBuilder {
	return NewError(ErrorTypeInternal, code, message).WithSeverity(SeverityCritical)
}

func Validation(code, message string) *ErrorBuilder {
	return NewError(ErrorTypeValidation, code, message).WithSeverity(SeverityWarning)
}

func Unauthorized(code, message string) *ErrorBuilder {
	return NewError(ErrorTypeUnauthorized, code, message).WithSeverity(SeverityWarning)
}

func ServiceUnavailable(code, message string) *ErrorBuilder {
	return NewError(ErrorTypeServiceUnavailable, code, message).WithSeverity(SeverityError)
}

func Storage(code, message string) *ErrorBuilder {
	return NewError(ErrorTypeStorage, code, message).WithSeverity(SeverityError)
}

func Database(code, message string) *ErrorBuilder {
	return NewError(ErrorTypeDatabase, code, message).WithSeverity(SeverityError)
}