// Copyright (c) 2024 THCloud.AI
// Author: OC
// Last Updated: 2024-12-26
// Description: Lotus API client implementation.

package lotus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client represents a Lotus API client
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient creates a new Lotus API client
func NewClient(cfg *Config) *Client {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// callRPC makes a JSON-RPC call to the Lotus API
func (c *Client) callRPC(ctx context.Context, method string, params []interface{}, result interface{}) error {
	// Prepare request
	reqBody := rpcRequest{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.config.APIURL, bytes.NewReader(reqBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.config.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.AuthToken)
	}

	// Implement retry logic
	var lastErr error
	for i := 0; i <= c.config.RetryCount; i++ {
		// Make request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to make request: %v", err)
			if i < c.config.RetryCount {
				time.Sleep(time.Duration(i+1) * time.Second)
				continue
			}
			return lastErr
		}
		defer resp.Body.Close()

		// Parse response
		var rpcResp rpcResponse
		if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
			lastErr = fmt.Errorf("failed to decode response: %v", err)
			if i < c.config.RetryCount {
				time.Sleep(time.Duration(i+1) * time.Second)
				continue
			}
			return lastErr
		}

		// Check for RPC error
		if rpcResp.Error != nil {
			lastErr = fmt.Errorf("RPC error: %s (code: %d)", rpcResp.Error.Message, rpcResp.Error.Code)
			if i < c.config.RetryCount {
				time.Sleep(time.Duration(i+1) * time.Second)
				continue
			}
			return lastErr
		}

		// Unmarshal result
		if err := json.Unmarshal(rpcResp.Result, result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %v", err)
		}

		return nil
	}

	return lastErr
}

// GetSectorPenalty calculates the penalty for a specific sector
func (c *Client) GetSectorPenalty(ctx context.Context, minerID string, sectorNum uint64) (string, error) {
	var result map[string]interface{}
	params := []interface{}{minerID, sectorNum}
	
	if err := c.callRPC(ctx, "Filecoin.SectorPenalty", params, &result); err != nil {
		return "", err
	}

	return result["penalty"].(string), nil
}

// GetVestedFunds gets the total vested funds for a miner
func (c *Client) GetVestedFunds(ctx context.Context, minerID string) (string, error) {
	var result map[string]interface{}
	params := []interface{}{minerID}
	
	if err := c.callRPC(ctx, "Filecoin.GetVestedFunds", params, &result); err != nil {
		return "", err
	}

	return result["vested"].(string), nil
}

// GetSectorInfo gets detailed information about a sector
func (c *Client) GetSectorInfo(ctx context.Context, minerID string, sectorNum uint64) (map[string]interface{}, error) {
	var result map[string]interface{}
	params := []interface{}{minerID, sectorNum}
	
	if err := c.callRPC(ctx, "Filecoin.StateSectorInfo", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetSectorStatus gets the current status of a sector
func (c *Client) GetSectorStatus(ctx context.Context, minerID string, sectorNum uint64) (map[string]interface{}, error) {
	var result map[string]interface{}
	params := []interface{}{minerID, sectorNum}
	
	if err := c.callRPC(ctx, "Filecoin.SectorStatus", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// ListSectors lists all sectors for a miner
func (c *Client) ListSectors(ctx context.Context, minerID string) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	params := []interface{}{minerID}
	
	if err := c.callRPC(ctx, "Filecoin.StateMinerSectors", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}
