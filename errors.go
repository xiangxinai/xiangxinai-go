package xiangxinai

import "fmt"

// XiangxinAIError 象信AI安全护栏基础错误类
type XiangxinAIError struct {
	Message string
	Cause   error
}

func (e *XiangxinAIError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *XiangxinAIError) Unwrap() error {
	return e.Cause
}

// NewXiangxinAIError 创建新的XiangxinAI错误
func NewXiangxinAIError(message string, cause error) *XiangxinAIError {
	return &XiangxinAIError{
		Message: message,
		Cause:   cause,
	}
}

// AuthenticationError 认证错误
type AuthenticationError struct {
	*XiangxinAIError
}

// NewAuthenticationError 创建认证错误
func NewAuthenticationError(message string) *AuthenticationError {
	return &AuthenticationError{
		XiangxinAIError: &XiangxinAIError{Message: message},
	}
}

// RateLimitError 速率限制错误
type RateLimitError struct {
	*XiangxinAIError
}

// NewRateLimitError 创建速率限制错误
func NewRateLimitError(message string) *RateLimitError {
	return &RateLimitError{
		XiangxinAIError: &XiangxinAIError{Message: message},
	}
}

// ValidationError 输入验证错误
type ValidationError struct {
	*XiangxinAIError
}

// NewValidationError 创建验证错误
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		XiangxinAIError: &XiangxinAIError{Message: message},
	}
}

// NetworkError 网络错误
type NetworkError struct {
	*XiangxinAIError
}

// NewNetworkError 创建网络错误
func NewNetworkError(message string, cause error) *NetworkError {
	return &NetworkError{
		XiangxinAIError: &XiangxinAIError{Message: message, Cause: cause},
	}
}

// ServerError 服务器错误
type ServerError struct {
	*XiangxinAIError
}

// NewServerError 创建服务器错误
func NewServerError(message string) *ServerError {
	return &ServerError{
		XiangxinAIError: &XiangxinAIError{Message: message},
	}
}