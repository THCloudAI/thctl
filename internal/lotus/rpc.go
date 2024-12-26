package lotus

import (
    "encoding/json"
    "fmt"
    "net/http"
)

// RPCRequest represents a JSON-RPC request
type RPCRequest struct {
    Jsonrpc string      `json:"jsonrpc"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params"`
    ID      int64      `json:"id"`
}

// RPCResponse represents a JSON-RPC response
type RPCResponse struct {
    Jsonrpc string          `json:"jsonrpc"`
    Result  json.RawMessage `json:"result,omitempty"`
    Error   *RPCError       `json:"error,omitempty"`
    ID      int64          `json:"id"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

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
        return NewLotusError(ErrNotFound, err.Message, nil)
    default:
        return NewLotusError(ErrUnknown, err.Message, nil)
    }
}
