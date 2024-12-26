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

// GetMinerAvailableBalance gets the available balance of a miner
func (c *Client) GetMinerAvailableBalance(ctx context.Context, minerID string) (string, error) {
	var result string
	params := []interface{}{minerID, nil}
	
	if err := c.callRPC(ctx, "Filecoin.StateMinerAvailableBalance", params, &result); err != nil {
		return "", err
	}

	return result, nil
}

// GetMinerFaults gets the faulty sectors of a miner
func (c *Client) GetMinerFaults(ctx context.Context, minerID string) ([]uint64, error) {
	var result []uint64
	params := []interface{}{minerID, nil}
	
	if err := c.callRPC(ctx, "Filecoin.StateMinerFaults", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetMinerDeadlines gets the deadlines information of a miner
func (c *Client) GetMinerDeadlines(ctx context.Context, minerID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	params := []interface{}{minerID, nil}
	
	if err := c.callRPC(ctx, "Filecoin.StateMinerDeadlines", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetMinerPartitions gets the partitions information of a miner
func (c *Client) GetMinerPartitions(ctx context.Context, minerID string, dlIdx uint64) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	params := []interface{}{minerID, dlIdx, nil}
	
	if err := c.callRPC(ctx, "Filecoin.StateMinerPartitions", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetMinerProvingDeadline gets the current proving deadline of a miner
func (c *Client) GetMinerProvingDeadline(ctx context.Context, minerID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	params := []interface{}{minerID, nil}
	
	if err := c.callRPC(ctx, "Filecoin.StateMinerProvingDeadline", params, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetMinerPreCommitDeposit gets the pre-commit deposit required for a sector
func (c *Client) GetMinerPreCommitDeposit(ctx context.Context, minerID string, sectorNumber uint64) (string, error) {
	var result string
	params := []interface{}{minerID, sectorNumber, nil}
	
	if err := c.callRPC(ctx, "Filecoin.StateMinerPreCommitDeposit", params, &result); err != nil {
		return "", err
	}

	return result, nil
}

// GetMinerInitialPledgeCollateral gets the initial pledge collateral required for a sector
func (c *Client) GetMinerInitialPledgeCollateral(ctx context.Context, minerID string, sectorNumber uint64) (string, error) {
	var result string
	params := []interface{}{minerID, sectorNumber, nil}
	
	if err := c.callRPC(ctx, "Filecoin.StateMinerInitialPledgeCollateral", params, &result); err != nil {
		return "", err
	}

	return result, nil
}

// GetMinerSectorAllocated checks if a sector number is allocated
func (c *Client) GetMinerSectorAllocated(ctx context.Context, minerID string, sectorNumber uint64) (bool, error) {
	var result bool
	params := []interface{}{minerID, sectorNumber, nil}
	
	if err := c.callRPC(ctx, "Filecoin.StateMinerSectorAllocated", params, &result); err != nil {
		return false, err
	}

	return result, nil
}

// GetMinerWorkerAddress gets the worker address of a miner
func (c *Client) GetMinerWorkerAddress(ctx context.Context, minerID string) (string, error) {
	minerInfo, err := c.GetMinerInfo(ctx, minerID)
	if err != nil {
		return "", err
	}

	worker, ok := minerInfo["Worker"].(string)
	if !ok {
		return "", fmt.Errorf("worker address not found in miner info")
	}

	return worker, nil
}

// GetMinerOwnerAddress gets the owner address of a miner
func (c *Client) GetMinerOwnerAddress(ctx context.Context, minerID string) (string, error) {
	minerInfo, err := c.GetMinerInfo(ctx, minerID)
	if err != nil {
		return "", err
	}

	owner, ok := minerInfo["Owner"].(string)
	if !ok {
		return "", fmt.Errorf("owner address not found in miner info")
	}

	return owner, nil
}

// GetMinerBeneficiaryAddress gets the beneficiary address of a miner
func (c *Client) GetMinerBeneficiaryAddress(ctx context.Context, minerID string) (string, error) {
	minerInfo, err := c.GetMinerInfo(ctx, minerID)
	if err != nil {
		return "", err
	}

	beneficiary, ok := minerInfo["Beneficiary"].(string)
	if !ok {
		return "", fmt.Errorf("beneficiary address not found in miner info")
	}

	return beneficiary, nil
}

// MinerInfo represents comprehensive information about a miner
type MinerInfo struct {
	BasicInfo        map[string]interface{} `json:"basic_info"`
	Power           map[string]interface{} `json:"power"`
	AvailableBalance string                 `json:"available_balance"`
	Faults          []uint64               `json:"faults"`
	Deadlines       map[string]interface{} `json:"deadlines"`
	ProvingDeadline map[string]interface{} `json:"proving_deadline"`
	WorkerAddress   string                 `json:"worker_address"`
	OwnerAddress    string                 `json:"owner_address"`
	Beneficiary     string                 `json:"beneficiary"`
}

// GetComprehensiveMinerInfo gets all available information about a miner
func (c *Client) GetComprehensiveMinerInfo(ctx context.Context, minerID string) (*MinerInfo, error) {
	info := &MinerInfo{}
	var err error

	// Get basic miner info
	info.BasicInfo, err = c.GetMinerInfo(ctx, minerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic info: %v", err)
	}

	// Get power info
	info.Power, err = c.GetMinerPower(ctx, minerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get power info: %v", err)
	}

	// Get available balance
	info.AvailableBalance, err = c.GetMinerAvailableBalance(ctx, minerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get available balance: %v", err)
	}

	// Get faults
	info.Faults, err = c.GetMinerFaults(ctx, minerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get faults: %v", err)
	}

	// Get deadlines
	info.Deadlines, err = c.GetMinerDeadlines(ctx, minerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deadlines: %v", err)
	}

	// Get proving deadline
	info.ProvingDeadline, err = c.GetMinerProvingDeadline(ctx, minerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proving deadline: %v", err)
	}

	// Get addresses
	info.WorkerAddress, err = c.GetMinerWorkerAddress(ctx, minerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get worker address: %v", err)
	}

	info.OwnerAddress, err = c.GetMinerOwnerAddress(ctx, minerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get owner address: %v", err)
	}

	info.Beneficiary, err = c.GetMinerBeneficiaryAddress(ctx, minerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get beneficiary address: %v", err)
	}

	return info, nil
}
