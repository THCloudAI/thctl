package lotus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/THCloudAI/thctl/internal/config"
	"github.com/multiformats/go-multiaddr"
)

// Config represents the configuration for a Lotus client
type Config struct {
	APIURL     string        `mapstructure:"api_url"`
	AuthToken  string        `mapstructure:"token"`
	Timeout    time.Duration `mapstructure:"timeout"`
	RetryCount int          `mapstructure:"retry_count"`
}

var defaultTimeout = 30 * time.Second

// New creates a new Lotus client
func New(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}
	if cfg.RetryCount == 0 {
		cfg.RetryCount = 3
	}

	// Convert multiaddr to HTTP URL if needed
	apiURL := cfg.APIURL
	if strings.HasPrefix(apiURL, "/ip4/") || strings.HasPrefix(apiURL, "/ip6/") {
		maddr, err := multiaddr.NewMultiaddr(apiURL)
		if err != nil {
			apiURL = fmt.Sprintf("http://%s", strings.TrimPrefix(apiURL, "/ip4/"))
		} else {
			// Extract host and port from multiaddr
			host, err := maddr.ValueForProtocol(multiaddr.P_IP4)
			if err != nil {
				host, _ = maddr.ValueForProtocol(multiaddr.P_IP6)
			}
			port, _ := maddr.ValueForProtocol(multiaddr.P_TCP)
			apiURL = fmt.Sprintf("http://%s:%s/rpc/v0", host, port)
		}
	}

	httpClient := &http.Client{
		Timeout: cfg.Timeout,
	}

	return &Client{
		apiURL:     apiURL,
		token:      cfg.AuthToken,
		httpClient: httpClient,
	}
}

// NewFromEnv creates a new Lotus client from environment variables
func NewFromEnv() (*Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	if cfg.Lotus.APIURL == "" {
		return nil, fmt.Errorf("LOTUS_API_URL environment variable is not set")
	}

	return New(Config{
		APIURL:    cfg.Lotus.APIURL,
		AuthToken: cfg.Lotus.AuthToken,
		Timeout:   cfg.Lotus.Timeout,
	}), nil
}

// Client represents a Lotus API client
type Client struct {
	apiURL     string
	token      string
	httpClient *http.Client
}

// callRPCWithRetry makes a JSON-RPC call to the Lotus API with retry
func (c *Client) callRPCWithRetry(ctx context.Context, method string, params interface{}, result interface{}) error {
	if c.apiURL == "" {
		return fmt.Errorf("LOTUS_API_URL is not set")
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      1,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var rpcResponse struct {
		Error  *struct{ Message string } `json:"error,omitempty"`
		Result json.RawMessage         `json:"result,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rpcResponse); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if rpcResponse.Error != nil {
		return fmt.Errorf("RPC error: %s", rpcResponse.Error.Message)
	}

	if err := json.Unmarshal(rpcResponse.Result, result); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return nil
}

// GetComprehensiveMinerInfo retrieves comprehensive information about a miner
func (c *Client) GetComprehensiveMinerInfo(ctx context.Context, minerID string) (*MinerInfo, error) {
	info := &MinerInfo{
		ID:                 minerID,
		Address:           minerID,
		Actor:             "storageminer",
		OwnedMiners:       make([]string, 0),
		WorkerMiners:      make([]string, 0),
		BenefitedMiners:   make([]string, 0),
		CreateTimestamp:   time.Now().Unix(),
		LastSeenTimestamp: time.Now().Unix(),
	}

	// Initialize nested structs
	info.Miner.Owner = struct {
		Address string `json:"address"`
		Balance string `json:"balance"`
		ID      string `json:"id,omitempty"`
	}{}
	info.Miner.Worker = struct {
		Address string `json:"address"`
		Balance string `json:"balance"`
		ID      string `json:"id,omitempty"`
	}{}
	info.Miner.Beneficiary = struct {
		Address string `json:"address"`
		Balance string `json:"balance"`
		ID      string `json:"id,omitempty"`
	}{}
	info.Miner.ControlAddresses = make([]ControlAddress, 0)
	info.Miner.MultiAddresses = make([]string, 0)
	info.Miner.Sectors = struct {
		Live       uint64 `json:"live"`
		Active     uint64 `json:"active"`
		Faulty     uint64 `json:"faulty"`
		Recovering uint64 `json:"recovering"`
	}{}

	// Initialize default values for strings
	info.Miner.TotalRewards = "0"
	info.Miner.PreCommitDeposits = "0"
	info.Miner.VestingFunds = "0"
	info.Miner.InitialPledgeRequirement = "0"
	info.Miner.AvailableBalance = "0"
	info.Miner.SectorPledgeBalance = "0"
	info.Miner.PledgeBalance = "0"
	info.Balance = "0"

	// Prepare batch RPC calls
	var (
		minerInfo  interface{}
		minerPower interface{}
		state      interface{}
		faults     interface{}
		recoveries interface{}
		active     interface{}
	)

	// First batch: Get basic miner info and state
	requests := []map[string]interface{}{
		{
			"jsonrpc": "2.0",
			"method":  "Filecoin.StateMinerInfo",
			"params":  []interface{}{minerID, nil},
			"id":      1,
		},
		{
			"jsonrpc": "2.0",
			"method":  "Filecoin.StateMinerPower",
			"params":  []interface{}{minerID, nil},
			"id":      2,
		},
		{
			"jsonrpc": "2.0",
			"method":  "Filecoin.StateReadState",
			"params":  []interface{}{minerID, nil},
			"id":      3,
		},
		{
			"jsonrpc": "2.0",
			"method":  "Filecoin.StateLookupRobustAddress",
			"params":  []interface{}{minerID, nil},
			"id":      4,
		},
		{
			"jsonrpc": "2.0",
			"method":  "Filecoin.StateMinerFaults",
			"params":  []interface{}{minerID, nil},
			"id":      5,
		},
		{
			"jsonrpc": "2.0",
			"method":  "Filecoin.StateMinerRecoveries",
			"params":  []interface{}{minerID, nil},
			"id":      6,
		},
		{
			"jsonrpc": "2.0",
			"method":  "Filecoin.StateMinerActiveSectors",
			"params":  []interface{}{minerID, nil},
			"id":      7,
		},
	}

	// Execute first batch request
	responses, err := c.BatchCallWithRetry(ctx, requests)
	if err != nil {
		return nil, fmt.Errorf("failed to execute first batch request: %w", err)
	}

	// Check for individual call errors and process responses
	for _, resp := range responses {
		if resp["error"] != nil {
			fmt.Printf("Error in RPC call: %v\n", resp["error"])
			continue
		}

		result, ok := resp["result"]
		if !ok {
			continue
		}

		id, ok := resp["id"].(float64)
		if !ok {
			continue
		}

		switch id {
		case 1:
			minerInfo = result
		case 2:
			minerPower = result
		case 3:
			state = result
		case 4:
			// Process robust address
			if result != nil {
				robustBytes, _ := json.MarshalIndent(result, "", "  ")
				fmt.Printf("Debug - StateLookupRobustAddress response:\n%s\n", string(robustBytes))
				if robustAddr, ok := result.(string); ok {
					info.Robust = robustAddr
				}
			} else {
				fmt.Printf("Debug - StateLookupRobustAddress returned nil\n")
			}
		case 5:
			faults = result
		case 6:
			recoveries = result
		case 7:
			active = result
		}
	}

	// Process results
	if err := c.processBasicInfo(ctx, info, minerInfo); err != nil {
		return nil, fmt.Errorf("failed to process basic info: %v", err)
	}

	c.processPowerInfo(info, minerPower)
	c.processStateInfo(info, state)
	c.processFaults(info, faults)
	c.processRecoveries(info, recoveries)
	c.processActiveSectors(info, active)

	return info, nil
}

// Process basic miner info
func (c *Client) processBasicInfo(ctx context.Context, info *MinerInfo, minerInfo interface{}) error {
	if minerInfo == nil {
		return fmt.Errorf("miner info is nil")
	}

	basicInfo, ok := minerInfo.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid miner info format")
	}

	// Extract owner address
	if owner, ok := basicInfo["Owner"].(string); ok {
		info.Miner.Owner.Address = owner
	}

	// Extract worker address
	if worker, ok := basicInfo["Worker"].(string); ok {
		info.Miner.Worker.Address = worker
	}

	// Extract beneficiary address
	if beneficiary, ok := basicInfo["Beneficiary"].(string); ok {
		info.Miner.Beneficiary.Address = beneficiary
	}

	// Extract control addresses
	if control, ok := basicInfo["ControlAddresses"].([]interface{}); ok {
		for _, addr := range control {
			if strAddr, ok := addr.(string); ok {
				info.Miner.ControlAddresses = append(info.Miner.ControlAddresses, ControlAddress{Address: strAddr})
			}
		}
	}

	// Extract peer ID
	if peerID, ok := basicInfo["PeerId"].(string); ok {
		info.Miner.PeerID = peerID
	}

	// Extract multiaddresses
	if multiaddrs, ok := basicInfo["Multiaddrs"].([]interface{}); ok {
		for _, addr := range multiaddrs {
			if bytes, ok := addr.([]byte); ok {
				maddr, err := multiaddr.NewMultiaddrBytes(bytes)
				if err == nil {
					info.Miner.MultiAddresses = append(info.Miner.MultiAddresses, maddr.String())
				}
			}
		}
	}

	// Extract sector size
	if sectorSize, ok := basicInfo["SectorSize"].(float64); ok {
		info.Miner.SectorSize = uint64(sectorSize)
	}

	return nil
}

// Process power info
func (c *Client) processPowerInfo(info *MinerInfo, powerInfo interface{}) {
	if powerInfo == nil {
		return
	}

	power, ok := powerInfo.(map[string]interface{})
	if !ok {
		return
	}

	if minerPower, ok := power["MinerPower"].(map[string]interface{}); ok {
		if raw, ok := minerPower["RawBytePower"].(string); ok {
			info.Miner.RawBytePower = raw
		}
		if quality, ok := minerPower["QualityAdjPower"].(string); ok {
			info.Miner.QualityAdjPower = quality
		}
	}

	if totalPower, ok := power["TotalPower"].(map[string]interface{}); ok {
		if raw, ok := totalPower["RawBytePower"].(string); ok {
			info.Miner.NetworkRawBytePower = raw
		}
		if quality, ok := totalPower["QualityAdjPower"].(string); ok {
			info.Miner.NetworkQualityAdjPower = quality
		}
	}
}

// Process chain head info
func (c *Client) processChainHead(info *MinerInfo, headInfo interface{}) {
	if headInfo == nil {
		return
	}

	head, ok := headInfo.(map[string]interface{})
	if !ok {
		return
	}

	if height, ok := head["Height"].(float64); ok {
		info.LastSeenHeight = uint64(height)
		info.CreateHeight = uint64(height) // 临时设置，实际应该从其他API获取
	}

	if blocks, ok := head["Blocks"].([]interface{}); ok && len(blocks) > 0 {
		if block, ok := blocks[0].(map[string]interface{}); ok {
			if timestamp, ok := block["Timestamp"].(float64); ok {
				info.LastSeenTimestamp = int64(timestamp)
				info.CreateTimestamp = int64(timestamp) // 临时设置，实际应该从其他API获取
			}
		}
	}
}

// Process actor info
func (c *Client) processActorInfo(info *MinerInfo, actorInfo interface{}) {
	if actorInfo == nil {
		return
	}

	actor, ok := actorInfo.(map[string]interface{})
	if !ok {
		return
	}

	if balance, ok := actor["Balance"].(string); ok {
		info.Balance = balance
	}
}

// Process state info from StateReadState
func (c *Client) processStateInfo(info *MinerInfo, stateInfo interface{}) {
	if stateInfo == nil {
		return
	}

	state, ok := stateInfo.(map[string]interface{})
	if !ok {
		return
	}

	// Process balance
	if balance, ok := state["Balance"].(string); ok {
		info.Balance = balance
	}

	if stateObj, ok := state["State"].(map[string]interface{}); ok {
		// Process balances
		if lockedFunds, ok := stateObj["LockedFunds"].(string); ok {
			info.Miner.AvailableBalance = lockedFunds
		}
		if pledge, ok := stateObj["InitialPledge"].(string); ok {
			info.Miner.InitialPledgeRequirement = pledge
			info.Miner.SectorPledgeBalance = pledge
			info.Miner.PledgeBalance = pledge
		}
		if deposits, ok := stateObj["PreCommitDeposits"].(string); ok {
			info.Miner.PreCommitDeposits = deposits
		}
		if vesting, ok := stateObj["VestingFunds"].(map[string]interface{}); ok {
			if vestingFunds, ok := vesting["/"].(string); ok {
				info.Miner.VestingFunds = vestingFunds
			}
		}
	}
}

// Process faults from StateMinerFaults
func (c *Client) processFaults(info *MinerInfo, faults interface{}) {
	if faults == nil {
		return
	}

	if faultArray, ok := faults.([]interface{}); ok {
		// Count non-zero values as faults
		faultCount := 0
		for _, fault := range faultArray {
			if val, ok := fault.(float64); ok && val != 0 {
				faultCount++
			}
		}
		info.Miner.Sectors.Faulty = uint64(faultCount)
	}
}

// Process recoveries from StateMinerRecoveries
func (c *Client) processRecoveries(info *MinerInfo, recoveries interface{}) {
	if recoveries == nil {
		return
	}

	if recoveryArray, ok := recoveries.([]interface{}); ok {
		// Count non-zero values as recoveries
		recoveryCount := 0
		for _, recovery := range recoveryArray {
			if val, ok := recovery.(float64); ok && val != 0 {
				recoveryCount++
			}
		}
		info.Miner.Sectors.Recovering = uint64(recoveryCount)
	}
}

// Process active sectors from StateMinerActiveSectors
func (c *Client) processActiveSectors(info *MinerInfo, active interface{}) {
	if active == nil {
		return
	}

	if sectors, ok := active.([]interface{}); ok {
		activeSectors := uint64(len(sectors))
		info.Miner.Sectors.Active = activeSectors
		info.Miner.Sectors.Live = activeSectors
	}
}

// BatchCall executes multiple RPC calls in a single request
func (c *Client) BatchCall(ctx context.Context, requests []map[string]interface{}) ([]map[string]interface{}, error) {
	if len(requests) == 0 {
		return nil, fmt.Errorf("no requests in batch")
	}

	// Marshal requests
	data, err := json.Marshal(requests)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal requests: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode response
	var responses []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responses); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return responses, nil
}

// BatchCallWithRetry executes batch RPC calls with retry mechanism
func (c *Client) BatchCallWithRetry(ctx context.Context, requests []map[string]interface{}) ([]map[string]interface{}, error) {
	var lastErr error
	for i := 0; i < 3; i++ {
		responses, err := c.BatchCall(ctx, requests)
		if err == nil {
			return responses, nil
		}
		if !isRetryableError(err) {
			return nil, err
		}
		lastErr = err
		time.Sleep(time.Second * time.Duration(i+1))
	}
	return nil, fmt.Errorf("failed after 3 retries: %v", lastErr)
}

// isRetryableError determines if an error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Network errors, timeouts, and 5xx status codes are retryable
	if netErr, ok := err.(net.Error); ok {
		return netErr.Temporary() || netErr.Timeout()
	}

	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "i/o timeout") ||
		strings.Contains(errStr, "unexpected status code: 5")
}

// GetSectorInfo retrieves information about a specific sector
func (c *Client) GetSectorInfo(ctx context.Context, minerID string, sectorNumber uint64) (*SectorInfo, error) {
	var result SectorInfo
	err := c.callRPCWithRetry(ctx, "Filecoin.StateSectorGetInfo", []interface{}{minerID, sectorNumber, nil}, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get sector info: %w", err)
	}
	return &result, nil
}

// ListSectors retrieves a list of sectors for a miner
func (c *Client) ListSectors(ctx context.Context, minerID string) ([]uint64, error) {
	var result []uint64
	err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerSectors", []interface{}{minerID, nil, nil}, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to list sectors: %w", err)
	}
	return result, nil
}

// GetSectorPenalty retrieves penalty information for a sector
func (c *Client) GetSectorPenalty(ctx context.Context, minerID string, sectorNumber uint64) (*SectorPenalty, error) {
	var result SectorPenalty
	err := c.callRPCWithRetry(ctx, "Filecoin.StateSectorPenalty", []interface{}{minerID, sectorNumber, nil}, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get sector penalty: %w", err)
	}
	return &result, nil
}

// GetSectorVested retrieves vesting information for a sector
func (c *Client) GetSectorVested(ctx context.Context, minerID string, sectorNumber uint64) (*SectorVested, error) {
	var result SectorVested
	err := c.callRPCWithRetry(ctx, "Filecoin.StateSectorVested", []interface{}{minerID, sectorNumber, nil}, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get sector vested: %w", err)
	}
	return &result, nil
}
