package xiangxinai

import "fmt"

// XiangxinAIError Xiangxin AI Guardrails base error class
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

// NewXiangxinAIError Create new XiangxinAI error
func NewXiangxinAIError(message string, cause error) *XiangxinAIError {
	return &XiangxinAIError{
		Message: message,
		Cause:   cause,
	}
}

// AuthenticationError Authentication error
type AuthenticationError struct {
	*XiangxinAIError
}

// NewAuthenticationError Create authentication error
func NewAuthenticationError(message string) *AuthenticationError {
	return &AuthenticationError{
		XiangxinAIError: &XiangxinAIError{Message: message},
	}
}

// RateLimitError Rate limit error
type RateLimitError struct {
	*XiangxinAIError
}

// NewRateLimitError Create rate limit error
func NewRateLimitError(message string) *RateLimitError {
	return &RateLimitError{
		XiangxinAIError: &XiangxinAIError{Message: message},
	}
}

// ValidationError Input validation error
type ValidationError struct {
	*XiangxinAIError
}

// NewValidationError Create validation error
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		XiangxinAIError: &XiangxinAIError{Message: message},
	}
}

// NetworkError Network error
type NetworkError struct {
	*XiangxinAIError
}

// NewNetworkError Create network error
func NewNetworkError(message string, cause error) *NetworkError {
	return &NetworkError{
		XiangxinAIError: &XiangxinAIError{Message: message, Cause: cause},
	}
}

// ServerError Server error
type ServerError struct {
	*XiangxinAIError
}

// NewServerError Create server error
func NewServerError(message string) *ServerError {
	return &ServerError{
		XiangxinAIError: &XiangxinAIError{Message: message},
	}
}