package lotus

import "fmt"

// Error codes for Lotus API errors
const (
	ErrAuthentication = iota + 1
	ErrNotFound
	ErrConnection
	ErrUnknown
	ErrInvalidParams
	ErrMethodNotFound
	ErrInvalidRequest
)

// LotusError represents a Lotus API error
type LotusError struct {
	Code    int
	Message string
	Cause   error
}

// Error implements the error interface for LotusError
func (e *LotusError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// NewLotusError creates a new LotusError
func NewLotusError(code int, message string, cause error) *LotusError {
	return &LotusError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// IsNotFound checks if the error is a NotFound error
func IsNotFound(err error) bool {
	if lotusErr, ok := err.(*LotusError); ok {
		return lotusErr.Code == ErrNotFound
	}
	return false
}

// IsConnectionError checks if the error is a connection error
func IsConnectionError(err error) bool {
	if lotusErr, ok := err.(*LotusError); ok {
		return lotusErr.Code == ErrConnection
	}
	return false
}

// IsAuthError checks if the error is an authentication error
func IsAuthError(err error) bool {
	if lotusErr, ok := err.(*LotusError); ok {
		return lotusErr.Code == ErrAuthentication
	}
	return false
}

// IsMethodNotFound checks if the error is a method not found error
func IsMethodNotFound(err error) bool {
	if lotusErr, ok := err.(*LotusError); ok {
		return lotusErr.Code == ErrMethodNotFound
	}
	return false
}

// IsInvalidRequest checks if the error is an invalid request error
func IsInvalidRequest(err error) bool {
	if lotusErr, ok := err.(*LotusError); ok {
		return lotusErr.Code == ErrInvalidRequest
	}
	return false
}
