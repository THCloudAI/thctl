package lotus

import (
	"fmt"
	"net/http"
)

// handleHTTPError converts HTTP errors to LotusError
func handleHTTPError(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return NewLotusError(ErrAuthentication, "unauthorized", nil)
	case http.StatusNotFound:
		return NewLotusError(ErrNotFound, "not found", nil)
	case http.StatusInternalServerError:
		return NewLotusError(ErrConnection, "internal server error", nil)
	default:
		return NewLotusError(ErrUnknown, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}
}

// handleRPCError converts RPC errors to LotusError
func handleRPCError(err *RPCError) error {
	switch err.Code {
	case -32000: // Generic server error
		return NewLotusError(ErrUnknown, err.Message, nil)
	case -32001: // Invalid params
		return NewLotusError(ErrInvalidParams, err.Message, nil)
	case -32002: // Method not found
		return NewLotusError(ErrMethodNotFound, err.Message, nil)
	case -32003: // Invalid request
		return NewLotusError(ErrInvalidRequest, err.Message, nil)
	default:
		return NewLotusError(ErrUnknown, err.Message, nil)
	}
}
