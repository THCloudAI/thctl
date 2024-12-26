package lotus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	ma "github.com/multiformats/go-multiaddr"
)

// Config represents the configuration for a Lotus client
type Config struct {
	APIURL     string
	AuthToken  string
	Timeout    time.Duration
	RetryCount int
}

// DefaultConfig returns the default configuration for a Lotus client
func DefaultConfig() *Config {
	return &Config{
		Timeout:    30 * time.Second,
		RetryCount: 3,
	}
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
	Error   *rpcError       `json:"error"`
	ID      int             `json:"id"`
}

// rpcError represents a JSON-RPC error
type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Client represents a Lotus API client
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient creates a new Lotus client with the given configuration
func NewClient(cfg *Config) *Client {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	apiInfo := os.Getenv("FULLNODE_API_INFO")
	if apiInfo != "" {
		parts := strings.Split(apiInfo, ":")
		if len(parts) == 2 {
			// Parse multiaddr from FULLNODE_API_INFO
			maddr, err := ma.NewMultiaddr(parts[1])
			if err == nil {
				// Extract host and port
				if host, err := maddr.ValueForProtocol(ma.P_IP4); err == nil {
					if port, err := maddr.ValueForProtocol(ma.P_TCP); err == nil {
						cfg.APIURL = fmt.Sprintf("http://%s:%s/rpc/v0", host, port)
						cfg.AuthToken = parts[0]
					}
				}
			}
		}
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	if cfg.RetryCount == 0 {
		cfg.RetryCount = 3
	}

	// Set default API URL if not set
	if cfg.APIURL == "" {
		cfg.APIURL = "http://127.0.0.1:1234/rpc/v0"
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
	// Create request
	req := rpcRequest{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.APIURL, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	if c.config.AuthToken != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.AuthToken))
	}

	// Make request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Parse response
	var rpcResp rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// Check for error
	if rpcResp.Error != nil {
		return fmt.Errorf("RPC error: %s (code: %d)", rpcResp.Error.Message, rpcResp.Error.Code)
	}

	// Unmarshal result
	if err := json.Unmarshal(rpcResp.Result, result); err != nil {
		return fmt.Errorf("failed to unmarshal result: %v", err)
	}

	return nil
}

// ListSectors lists all sectors for a miner
func (c *Client) ListSectors(ctx context.Context, minerID string) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	params := []interface{}{minerID, []interface{}{}, true} // Empty array for tipset key and show committed sectors only (true)
	
	if err := c.callRPC(ctx, "Filecoin.StateMinerSectors", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetSectorStatus gets the status of a sector
func (c *Client) GetSectorStatus(ctx context.Context, minerID string, sectorNum uint64) (map[string]interface{}, error) {
	var result map[string]interface{}
	params := []interface{}{minerID, sectorNum}
	
	if err := c.callRPC(ctx, "Filecoin.StateSectorGetInfo", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetSectorInfo gets detailed information about a sector
func (c *Client) GetSectorInfo(ctx context.Context, minerID string, sectorNum uint64) (map[string]interface{}, error) {
	var result map[string]interface{}
	params := []interface{}{minerID, sectorNum}
	
	if err := c.callRPC(ctx, "Filecoin.StateSectorGetInfo", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetSectorPenalty gets the penalty for a sector
func (c *Client) GetSectorPenalty(ctx context.Context, minerID string, sectorNum uint64) (*SectorPenaltyResponse, error) {
	var result SectorPenaltyResponse
	params := []interface{}{minerID, sectorNum}
	
	if err := c.callRPC(ctx, "Filecoin.StateMinerInitialPledgeCollateral", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetVestedFunds gets the vested funds for a miner
func (c *Client) GetVestedFunds(ctx context.Context, minerID string) (*VestedFundsResponse, error) {
	var result VestedFundsResponse
	params := []interface{}{minerID}
	
	if err := c.callRPC(ctx, "Filecoin.StateMinerAvailableBalance", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SectorPenaltyResponse represents the response for the GetSectorPenalty method
type SectorPenaltyResponse struct {
	InitialPledge string `json:"InitialPledge"`
}

// VestedFundsResponse represents the response for the GetVestedFunds method
type VestedFundsResponse struct {
	AvailableBalance string `json:"AvailableBalance"`
}
