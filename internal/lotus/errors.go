package lotus

import (
    "fmt"
)

// ErrorCode represents specific error types in Lotus operations
type ErrorCode int

const (
    // ErrUnknown represents an unknown error
    ErrUnknown ErrorCode = iota
    // ErrConnection represents connection related errors
    ErrConnection
    // ErrAuthentication represents authentication related errors
    ErrAuthentication
    // ErrInvalidParams represents invalid parameter errors
    ErrInvalidParams
    // ErrNotFound represents resource not found errors
    ErrNotFound
    // ErrRPCTimeout represents RPC timeout errors
    ErrRPCTimeout
)

// LotusError represents a structured error from Lotus operations
type LotusError struct {
    Code    ErrorCode
    Message string
    Cause   error
}

func (e *LotusError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Cause)
    }
    return e.Message
}

// NewLotusError creates a new LotusError
func NewLotusError(code ErrorCode, message string, cause error) *LotusError {
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
