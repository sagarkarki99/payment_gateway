package models

// ServiceError represents a service-level error
type ServiceError struct {
	Code    ErrorCode
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

// ErrorCode represents different types of errors
type ErrorCode int

const (
	ErrorCodeUnknown ErrorCode = iota
	ErrorCodeValidation
	ErrorCodeNotFound
	ErrorCodeInsufficientFunds
	ErrorCodeGatewayError
	ErrorCodeUnauthorized
)

// NewServiceError creates a new ServiceError
func NewServiceError(code ErrorCode, message string) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
	}
}

// Error response mapping to HTTP status codes
var errorToStatusCode = map[ErrorCode]int{
	ErrorCodeUnknown:           500,
	ErrorCodeValidation:        400,
	ErrorCodeNotFound:          404,
	ErrorCodeInsufficientFunds: 422,
	ErrorCodeGatewayError:      502,
	ErrorCodeUnauthorized:      401,
}

// GetStatusCode returns the appropriate HTTP status code for an error
func GetStatusCode(err error) int {
	if serviceErr, ok := err.(*ServiceError); ok {
		if code, exists := errorToStatusCode[serviceErr.Code]; exists {
			return code
		}
	}
	return 500 // Default to internal server error
}
