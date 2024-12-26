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
	APIURL     string        `mapstructure:"api_url"`
	AuthToken  string        `mapstructure:"token"`
	Timeout    time.Duration `mapstructure:"timeout"`
	RetryCount int           `mapstructure:"retry_count"`
}

// DefaultConfig returns the default configuration for a Lotus client
func DefaultConfig() *Config {
	return &Config{
		APIURL:  "http://127.0.0.1:1234/rpc/v0",
		Timeout: 30 * time.Second,
	}
}

// Client represents a Lotus API client
type Client struct {
	cfg   Config
	client *http.Client
}

// New creates a new Lotus client
func New(cfg Config) *Client {
	if cfg.APIURL == "" {
		cfg.APIURL = "http://127.0.0.1:1234/rpc/v0"
	}
	if cfg.AuthToken == "" {
		if token := os.Getenv("LOTUS_API_TOKEN"); token != "" {
			cfg.AuthToken = token
		}
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &Client{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// NewFromEnv creates a new Lotus client from environment variables
func NewFromEnv() *Client {
	config := DefaultConfig()

	if url := os.Getenv("LOTUS_API_URL"); url != "" {
		config.APIURL = url
	}
	if token := os.Getenv("LOTUS_API_TOKEN"); token != "" {
		config.AuthToken = token
	}

	return New(*config)
}

// callRPC makes a JSON-RPC call to the Lotus API
func (c *Client) callRPC(ctx context.Context, method string, params interface{}, result interface{}) error {
	// Convert multiaddr URL to HTTP URL if necessary
	apiURL := c.cfg.APIURL
	if strings.HasPrefix(apiURL, "/ip4/") || strings.HasPrefix(apiURL, "/ip6/") {
		maddr, err := ma.NewMultiaddr(apiURL)
		if err != nil {
			return fmt.Errorf("failed to parse multiaddr: %v", err)
		}

		// Extract IP and port from multiaddr
		ip, err := maddr.ValueForProtocol(ma.P_IP4)
		if err != nil {
			ip, err = maddr.ValueForProtocol(ma.P_IP6)
			if err != nil {
				return fmt.Errorf("failed to get IP from multiaddr: %v", err)
			}
		}
		port, err := maddr.ValueForProtocol(ma.P_TCP)
		if err != nil {
			return fmt.Errorf("failed to get port from multiaddr: %v", err)
		}

		// Construct HTTP URL
		apiURL = fmt.Sprintf("http://%s:%s/rpc/v0", ip, port)
	}

	// Prepare request body
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":     1,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(reqBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.AuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.cfg.AuthToken))
	}

	// Make request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Parse response
	var rpcResp struct {
		JSONRPC string          `json:"jsonrpc"`
		Result  json.RawMessage `json:"result"`
		Error   *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
		ID int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// Check for RPC error
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
	params := []interface{}{minerID, nil, false} // Use null for tipset key and false for show uncommitted
	
	if err := c.callRPC(ctx, "Filecoin.StateMinerSectors", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetSectorInfo gets information about a specific sector
func (c *Client) GetSectorInfo(ctx context.Context, minerID string, sectorNumber int64) (map[string]interface{}, error) {
	var result map[string]interface{}
	params := []interface{}{minerID, sectorNumber}
	
	if err := c.callRPC(ctx, "Filecoin.StateSectorGetInfo", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetSectorPenalty gets penalty information for a specific sector
func (c *Client) GetSectorPenalty(ctx context.Context, minerID string, sectorNumber int64) (map[string]interface{}, error) {
	var result map[string]interface{}
	params := []interface{}{minerID, sectorNumber}
	
	if err := c.callRPC(ctx, "Filecoin.StateSectorPenalty", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetSectorVested gets vesting information for a specific sector
func (c *Client) GetSectorVested(ctx context.Context, minerID string, sectorNumber int64) (map[string]interface{}, error) {
	var result map[string]interface{}
	params := []interface{}{minerID, sectorNumber}
	
	if err := c.callRPC(ctx, "Filecoin.StateSectorVested", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetMinerInfo gets basic information about a miner
func (c *Client) GetMinerInfo(ctx context.Context, minerID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	params := []interface{}{minerID, nil} // Use null for tipset key
	
	if err := c.callRPC(ctx, "Filecoin.StateMinerInfo", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetMinerPower gets power information about a miner
func (c *Client) GetMinerPower(ctx context.Context, minerID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	params := []interface{}{minerID, nil} // Use null for tipset key
	
	if err := c.callRPC(ctx, "Filecoin.StateMinerPower", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}
