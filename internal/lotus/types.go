// Copyright (c) 2024 THCloud.AI
// Author: OC
// Last Updated: 2024-12-26
// Description: Lotus API types.

package lotus

import (
	"encoding/json"
	"time"
)

// Config represents Lotus client configuration
type Config struct {
	APIURL     string        `mapstructure:"api_url"`
	AuthToken  string        `mapstructure:"token"`
	Timeout    time.Duration `mapstructure:"timeout"`
	RetryCount int           `mapstructure:"retry_count"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		APIURL:     "http://127.0.0.1:1234/rpc/v0",
		Timeout:    30 * time.Second,
		RetryCount: 3,
	}
}

// SectorPenaltyResponse represents the response for the GetSectorPenalty method
type SectorPenaltyResponse struct {
	Penalty string `json:"penalty"`
}

// VestedFundsResponse represents the response for the GetVestedFunds method
type VestedFundsResponse struct {
	Vested string `json:"vested"`
}

// SectorInfo represents the information for a sector
type SectorInfo struct {
	SectorNumber uint64 `json:"sector_number"`
	SealProof    string `json:"seal_proof"`
	Activation   uint64 `json:"activation"`
	Expiration   uint64 `json:"expiration"`
}

// SectorStatus represents the status of a sector
type SectorStatus struct {
	Status string `json:"status"`
}

// rpcRequest represents a JSON-RPC request
type rpcRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

// rpcResponse represents a JSON-RPC response
type rpcResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error,omitempty"`
	ID      int             `json:"id"`
}

// rpcError represents a JSON-RPC error
type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
