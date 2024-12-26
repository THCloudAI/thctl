package lotus

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "math/big"
    "net/http"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "sync"
    "time"

    "github.com/multiformats/go-multiaddr"
)

// Config represents the configuration for a Lotus client
type Config struct {
    APIURL     string        `mapstructure:"api_url"`
    AuthToken  string        `mapstructure:"token"`
    Timeout    time.Duration `mapstructure:"timeout"`
    RetryCount int           `mapstructure:"retry_count"`
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

    httpClient := &http.Client{
        Timeout: cfg.Timeout,
    }

    return &Client{
        apiURL:     cfg.APIURL,
        token:      cfg.AuthToken,
        httpClient: httpClient,
    }
}

// NewFromEnv creates a new Lotus client from environment variables
func NewFromEnv() (*Client, error) {
    // Try to read from .thctl.env file first
    home, err := os.UserHomeDir()
    if err != nil {
        return nil, fmt.Errorf("failed to get home directory: %v", err)
    }

    envFile := filepath.Join(home, ".thctl.env")
    if _, err := os.Stat(envFile); err == nil {
        data, err := os.ReadFile(envFile)
        if err != nil {
            return nil, fmt.Errorf("failed to read .thctl.env: %v", err)
        }

        // Parse environment variables format
        var apiURL, authToken string
        lines := strings.Split(string(data), "\n")
        for _, line := range lines {
            line = strings.TrimSpace(line)
            if line == "" || strings.HasPrefix(line, "#") {
                continue
            }

            parts := strings.SplitN(line, "=", 2)
            if len(parts) != 2 {
                continue
            }

            key := strings.TrimSpace(parts[0])
            value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)

            switch key {
            case "LOTUS_API_URL":
                apiURL = value
            case "LOTUS_API_TOKEN":
                authToken = value
            }
        }

        if apiURL != "" && authToken != "" {
            // Validate multiaddr format
            if !strings.HasPrefix(apiURL, "/ip4/") && !strings.HasPrefix(apiURL, "/ip6/") {
                return nil, fmt.Errorf("LOTUS_API_URL must be in multiaddr format (e.g., /ip4/127.0.0.1/tcp/1234)")
            }

            cfg := Config{
                APIURL:    apiURL,
                AuthToken: authToken,
                Timeout:   30 * time.Second,
            }

            return New(cfg), nil
        }
    }

    return nil, fmt.Errorf("LOTUS_API_URL and LOTUS_API_TOKEN must be set in .thctl.env")
}

// Client represents a Lotus API client
type Client struct {
    apiURL     string
    token      string
    httpClient *http.Client
}

// callRPCWithRetry makes a JSON-RPC call to the Lotus API with retry
func (c *Client) callRPCWithRetry(ctx context.Context, method string, params []interface{}, result interface{}) error {
    if c.apiURL == "" {
        return fmt.Errorf("LOTUS_API_URL is not set")
    }

    // Parse multiaddr
    maddr, err := multiaddr.NewMultiaddr(c.apiURL)
    if err != nil {
        return fmt.Errorf("invalid multiaddr: %v", err)
    }

    // Extract host and port
    var host, port string
    multiaddr.ForEach(maddr, func(comp multiaddr.Component) bool {
        switch comp.Protocol().Code {
        case multiaddr.P_IP4, multiaddr.P_IP6:
            host = comp.Value()
        case multiaddr.P_TCP:
            port = comp.Value()
        }
        return true
    })

    if host == "" || port == "" {
        return fmt.Errorf("invalid multiaddr: missing host or port")
    }

    // Construct HTTP URL
    httpURL := fmt.Sprintf("http://%s:%s/rpc/v0", host, port)

    requestBody, err := json.Marshal(map[string]interface{}{
        "jsonrpc": "2.0",
        "method":  method,
        "params":  params,
        "id":     1,
    })
    if err != nil {
        return fmt.Errorf("failed to marshal request: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, "POST", httpURL, bytes.NewBuffer(requestBody))
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

// ListSectors lists all sectors for a miner
func (c *Client) ListSectors(ctx context.Context, minerID string) ([]map[string]interface{}, error) {
    var result []map[string]interface{}
    params := []interface{}{minerID, nil, false} // Use null for tipset key and false for show uncommitted
    
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerSectors", params, &result); err != nil {
        return nil, err
    }

    return result, nil
}

// GetSectorInfo gets information about a specific sector
func (c *Client) GetSectorInfo(ctx context.Context, minerID string, sectorNumber int64) (map[string]interface{}, error) {
    var result map[string]interface{}
    params := []interface{}{minerID, sectorNumber}
    
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateSectorGetInfo", params, &result); err != nil {
        return nil, err
    }

    return result, nil
}

// GetSectorPenalty gets penalty information for a specific sector
func (c *Client) GetSectorPenalty(ctx context.Context, minerID string, sectorNumber int64) (map[string]interface{}, error) {
    var result map[string]interface{}
    params := []interface{}{minerID, sectorNumber}
    
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateSectorPenalty", params, &result); err != nil {
        return nil, err
    }

    return result, nil
}

// GetSectorVested gets vesting information for a specific sector
func (c *Client) GetSectorVested(ctx context.Context, minerID string, sectorNumber int64) (map[string]interface{}, error) {
    var result map[string]interface{}
    params := []interface{}{minerID, sectorNumber}
    
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateSectorVested", params, &result); err != nil {
        return nil, err
    }

    return result, nil
}

// GetMinerInfo gets basic information about a miner
func (c *Client) GetMinerInfo(ctx context.Context, minerID string) (map[string]interface{}, error) {
    if minerID == "" {
        return nil, errors.New("miner ID cannot be empty")
    }

    var result map[string]interface{}
    params := []interface{}{minerID, nil}
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerInfo", params, &result); err != nil {
        return nil, fmt.Errorf("failed to get miner info: %w", err)
    }

    // Get miner sectors info
    var sectorsInfo map[string]interface{}
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerSectors", params, &sectorsInfo); err == nil {
        result["SectorsInfo"] = sectorsInfo
    }

    // Get miner proving deadline info
    var deadlineInfo map[string]interface{}
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerProvingDeadline", params, &deadlineInfo); err == nil {
        result["DeadlineInfo"] = deadlineInfo
    }

    // Get miner available balance
    var availableBalance string
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerAvailableBalance", params, &availableBalance); err == nil {
        result["AvailableBalance"] = availableBalance
    }

    // Get miner vesting funds
    var vestingFunds map[string]interface{}
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerVestingFunds", params, &vestingFunds); err == nil {
        result["VestingFunds"] = vestingFunds
    }

    // Get miner faults
    var faults []uint64
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerFaults", params, &faults); err == nil {
        result["Faults"] = faults
    }

    // Get miner recoveries
    var recoveries []uint64
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerRecoveries", params, &recoveries); err == nil {
        result["Recoveries"] = recoveries
    }

    return result, nil
}

// GetMinerPower gets power information about a miner
func (c *Client) GetMinerPower(ctx context.Context, minerID string) (map[string]interface{}, error) {
    var result map[string]interface{}
    params := []interface{}{minerID, nil} // Use null for tipset key
    
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerPower", params, &result); err != nil {
        return nil, err
    }

    return result, nil
}

// GetMinerAvailableBalance gets the available balance of a miner
func (c *Client) GetMinerAvailableBalance(ctx context.Context, minerID string) (string, error) {
    var result string
    params := []interface{}{minerID, nil}
    
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerAvailableBalance", params, &result); err != nil {
        return "", err
    }

    return result, nil
}

// GetMinerFaults gets the faulty sectors of a miner
func (c *Client) GetMinerFaults(ctx context.Context, minerID string) ([]uint64, error) {
    var result []uint64
    params := []interface{}{minerID, nil}
    
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerFaults", params, &result); err != nil {
        return nil, err
    }

    return result, nil
}

// GetMinerDeadlines gets the deadlines information of a miner
func (c *Client) GetMinerDeadlines(ctx context.Context, minerID string) ([]map[string]interface{}, error) {
    var result []map[string]interface{}
    params := []interface{}{minerID, nil}
    
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerDeadlines", params, &result); err != nil {
        return nil, err
    }

    return result, nil
}

// GetMinerPartitions gets the partitions information of a miner
func (c *Client) GetMinerPartitions(ctx context.Context, minerID string, dlIdx uint64) ([]map[string]interface{}, error) {
    var result []map[string]interface{}
    params := []interface{}{minerID, dlIdx, nil}
    
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerPartitions", params, &result); err != nil {
        return nil, err
    }

    return result, nil
}

// GetMinerProvingDeadline gets the current proving deadline of a miner
func (c *Client) GetMinerProvingDeadline(ctx context.Context, minerID string) (map[string]interface{}, error) {
    var result map[string]interface{}
    params := []interface{}{minerID, nil}
    
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerProvingDeadline", params, &result); err != nil {
        return nil, err
    }

    return result, nil
}

// GetMinerPreCommitDeposit gets the pre-commit deposit required for a sector
func (c *Client) GetMinerPreCommitDeposit(ctx context.Context, minerID string, sectorNumber uint64) (string, error) {
    var result string
    params := []interface{}{minerID, sectorNumber, nil}
    
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerPreCommitDeposit", params, &result); err != nil {
        return "", err
    }

    return result, nil
}

// GetMinerInitialPledgeCollateral gets the initial pledge collateral required for a sector
func (c *Client) GetMinerInitialPledgeCollateral(ctx context.Context, minerID string, sectorNumber uint64) (string, error) {
    var result string
    params := []interface{}{minerID, sectorNumber, nil}
    
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerInitialPledgeCollateral", params, &result); err != nil {
        return "", fmt.Errorf("failed to get initial pledge: %v", err)
    }
    return result, nil
}

// GetMinerSectorAllocated checks if a sector number is allocated
func (c *Client) GetMinerSectorAllocated(ctx context.Context, minerID string, sectorNumber uint64) (bool, error) {
    var result bool
    params := []interface{}{minerID, sectorNumber, nil}
    
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerSectorAllocated", params, &result); err != nil {
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

type controlAddress struct {
    Address string `json:"address"`
    Balance string `json:"balance"`
    ID      string `json:"id,omitempty"`
}

// Response represents a standardized API response
type Response struct {
    Version   string      `json:"version"`
    Timestamp int64       `json:"timestamp"`
    Status    string      `json:"status"`
    Data      *MinerInfo  `json:"data"`
}

// MinerInfo represents comprehensive information about a miner
type MinerInfo struct {
    ID                string   `json:"id"`
    Robust            string   `json:"robust"`
    Actor             string   `json:"actor"`
    CreateHeight      uint64   `json:"createHeight"`
    CreateTimestamp   int64    `json:"createTimestamp"`
    LastSeenHeight    uint64   `json:"lastSeenHeight"`
    LastSeenTimestamp int64    `json:"lastSeenTimestamp"`
    Balance           string   `json:"balance"`
    MessageCount      uint64   `json:"messageCount"`
    TransferCount     uint64   `json:"transferCount"`
    TokenTransferCount uint64   `json:"tokenTransferCount"`
    Timestamp         int64    `json:"timestamp"`
    Tokens           uint64   `json:"tokens"`
    Miner            struct {
        Owner struct {
            Address string `json:"address"`
            Balance string `json:"balance"`
            ID      string `json:"id,omitempty"`
        } `json:"owner"`
        Worker struct {
            Address string `json:"address"`
            Balance string `json:"balance"`
            ID      string `json:"id,omitempty"`
        } `json:"worker"`
        Beneficiary struct {
            Address string `json:"address"`
            Balance string `json:"balance"`
            ID      string `json:"id,omitempty"`
        } `json:"beneficiary"`
        ControlAddresses []controlAddress `json:"controlAddresses"`
        PeerID          string   `json:"peerId"`
        MultiAddresses  []string `json:"multiAddresses"`
        SectorSize      uint64   `json:"sectorSize"`
        RawBytePower    string   `json:"rawBytePower"`
        QualityAdjPower string   `json:"qualityAdjPower"`
        NetworkRawBytePower     string   `json:"networkRawBytePower"`
        NetworkQualityAdjPower  string   `json:"networkQualityAdjPower"`
        BlocksMined             uint64   `json:"blocksMined"`
        WeightedBlocksMined     uint64   `json:"weightedBlocksMined"`
        TotalRewards           string   `json:"totalRewards"`
        Sectors struct {
            Live      uint64 `json:"live"`
            Active    uint64 `json:"active"`
            Faulty    uint64 `json:"faulty"`
            Recovering uint64 `json:"recovering"`
        } `json:"sectors"`
        PreCommitDeposits         string   `json:"preCommitDeposits"`
        VestingFunds             string   `json:"vestingFunds"`
        InitialPledgeRequirement string   `json:"initialPledgeRequirement"`
        AvailableBalance        string   `json:"availableBalance"`
        SectorPledgeBalance     string   `json:"sectorPledgeBalance"`
        PledgeBalance          string   `json:"pledgeBalance"`
        RawBytePowerRank       uint64   `json:"rawBytePowerRank"`
        QualityAdjPowerRank    uint64   `json:"qualityAdjPowerRank"`
    } `json:"miner"`
    OwnedMiners     []string `json:"ownedMiners"`
    WorkerMiners    []string `json:"workerMiners"`
    BenefitedMiners []string `json:"benefitedMiners"`
    Address         string   `json:"address"`
}

// GetComprehensiveMinerInfo gets all available information about a miner
func (c *Client) GetComprehensiveMinerInfo(ctx context.Context, minerID string) (*MinerInfo, error) {
    if minerID == "" {
        return nil, errors.New("miner ID cannot be empty")
    }

    // 创建基础信息结构
    info := &MinerInfo{
        ID:                minerID,
        Actor:             "storageminer",
        Address:          minerID,
        OwnedMiners:      make([]string, 0),
        WorkerMiners:     make([]string, 0),
        BenefitedMiners:  make([]string, 0),
        TokenTransferCount: 0,
        Tokens:            0,
        Timestamp:        time.Now().Unix(),
    }

    // 获取必需的基础信息（同步）
    basicInfo, err := c.GetMinerInfo(ctx, minerID)
    if err != nil {
        return nil, fmt.Errorf("failed to get basic info: %w", err)
    }

    if err := c.processBasicInfo(ctx, info, basicInfo); err != nil {
        return nil, err
    }

    actorInfo, err := c.GetActorInfo(ctx, minerID)
    if err != nil {
        return nil, fmt.Errorf("failed to get actor info: %w", err)
    }

    if err := c.processActorInfo(info, actorInfo); err != nil {
        return nil, err
    }

    // 创建带缓冲的通道控制并发
    maxConcurrency := 5
    semaphore := make(chan struct{}, maxConcurrency)
    defer close(semaphore)

    // 并发任务组
    tasks := []struct {
        name string
        fn   func() error
    }{
        {"robust_address", func() error {
            addr, err := c.GetRobustAddress(ctx, minerID)
            if err == nil {
                info.Robust = addr
            }
            return err
        }},
        {"actor_state", func() error {
            state, err := c.GetActorState(ctx, minerID)
            if err == nil {
                c.processActorState(ctx, info, state)
            }
            return err
        }},
        {"power_info", func() error {
            power, err := c.GetMinerPower(ctx, minerID)
            if err == nil {
                c.processPowerInfo(info, power)
            }
            return err
        }},
        {"sector_info", func() error {
            params := []interface{}{minerID, nil}
            var sectorsInfo []map[string]interface{}
            if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerSectors", params, &sectorsInfo); err == nil {
                info.Miner.Sectors.Live = uint64(len(sectorsInfo))
                info.Miner.Sectors.Active = uint64(len(sectorsInfo))
            }
            return err
        }},
        {"faults", func() error {
            faults, err := c.GetMinerFaults(ctx, minerID)
            if err == nil {
                info.Miner.Sectors.Faulty = uint64(len(faults))
            }
            return err
        }},
        {"financial_info", func() error {
            balance, err := c.GetMinerAvailableBalance(ctx, minerID)
            if err == nil {
                info.Miner.AvailableBalance = balance
            }
            pledge, err := c.GetMinerPledgeCollateral(ctx, minerID)
            if err == nil {
                info.Miner.InitialPledgeRequirement = pledge
                info.Miner.SectorPledgeBalance = pledge
                info.Miner.PledgeBalance = pledge
            }
            return err
        }},
        {"power_rank", func() error {
            rawRank, qualityRank, err := c.GetMinerPowerRank(ctx, minerID)
            if err == nil {
                info.Miner.RawBytePowerRank = rawRank
                info.Miner.QualityAdjPowerRank = qualityRank
            }
            return err
        }},
        {"chain_head", func() error {
            head, err := c.GetChainHead(ctx)
            if err == nil {
                c.processChainHead(info, head)
            }
            return err
        }},
    }

    // 并发执行任务
    var wg sync.WaitGroup
    for _, task := range tasks {
        wg.Add(1)
        go func(t struct {
            name string
            fn   func() error
        }) {
            defer wg.Done()
            
            // 获取并发信号量
            semaphore <- struct{}{}
            defer func() { <-semaphore }()

            c.asyncRetry(ctx, t.name, info, t.fn)
        }(task)
    }

    // 等待所有任务完成或超时
    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        // 所有任务完成
        return info, nil
    case <-time.After(30 * time.Second):
        // 超时处理
        return info, fmt.Errorf("miner info retrieval timed out after 30 seconds")
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

// asyncRetry 改进版异步重试函数
func (c *Client) asyncRetry(ctx context.Context, name string, info *MinerInfo, fn func() error) {
    go func() {
        retries := 3
        backoff := time.Second
        
        for i := 0; i < retries; i++ {
            // 检查上下文是否已取消
            select {
            case <-ctx.Done():
                return
            default:
            }

            err := fn()
            if err == nil {
                return
            }

            // 判断是否可重试的错误
            if !isRetryableError(err) {
                fmt.Printf("Non-retryable error for %s: %v\n", name, err)
                return
            }

            // 如果不是最后一次重试，等待后继续
            if i < retries-1 {
                fmt.Printf("Warning: %s failed, retrying in %v: %v\n", name, backoff, err)
                time.Sleep(backoff)
                backoff *= 2 // 指数退避
            } else {
                fmt.Printf("Error: %s failed after %d retries: %v\n", name, retries, err)
            }
        }
    }()
}

// isRetryableError 判断是否为可重试的错误
func isRetryableError(err error) bool {
    if err == nil {
        return false
    }
    
    // 网络和RPC相关的可重试错误
    retryableErrors := []string{
        "connection reset", 
        "connection refused", 
        "timeout", 
        "i/o timeout", 
        "EOF", 
        "network is unreachable", 
        "500", 
        "503", 
        "504",
    }
    
    errStr := strings.ToLower(err.Error())
    for _, retryErr := range retryableErrors {
        if strings.Contains(errStr, retryErr) {
            return true
        }
    }
    
    return false
}

// Process basic miner info
func (c *Client) processBasicInfo(ctx context.Context, info *MinerInfo, basicInfo map[string]interface{}) error {
    // Process owner info
    if owner, ok := basicInfo["Owner"].(string); ok {
        info.Miner.Owner.Address = owner
        if balance, err := c.GetBalance(ctx, owner); err == nil {
            info.Miner.Owner.Balance = balance
        }
    } else {
        return fmt.Errorf("owner address not found")
    }

    // Process worker info
    if worker, ok := basicInfo["Worker"].(string); ok {
        info.Miner.Worker.Address = worker
        if balance, err := c.GetBalance(ctx, worker); err == nil {
            info.Miner.Worker.Balance = balance
        }
    } else {
        return fmt.Errorf("worker address not found")
    }

    // Process beneficiary info
    if beneficiary, ok := basicInfo["Beneficiary"].(string); ok {
        info.Miner.Beneficiary.Address = beneficiary
        if balance, err := c.GetBalance(ctx, beneficiary); err == nil {
            info.Miner.Beneficiary.Balance = balance
        }
    }

    // Process control addresses
    if controlAddrs, ok := basicInfo["ControlAddresses"].([]interface{}); ok {
        info.Miner.ControlAddresses = make([]controlAddress, 0)
        for _, addr := range controlAddrs {
            if str, ok := addr.(string); ok {
                balance, _ := c.GetBalance(ctx, str)
                info.Miner.ControlAddresses = append(info.Miner.ControlAddresses, controlAddress{
                    Address: str,
                    Balance: balance,
                })
            }
        }
    }

    // Process sector size
    if sectorSize, ok := basicInfo["SectorSize"].(uint64); ok {
        info.Miner.SectorSize = sectorSize
    }

    return nil
}

// Process actor info
func (c *Client) processActorInfo(info *MinerInfo, actorInfo map[string]interface{}) error {
    if balance, ok := actorInfo["Balance"].(string); ok {
        info.Balance = balance
    }
    if nonce, ok := actorInfo["Nonce"].(float64); ok {
        info.MessageCount = uint64(nonce)
    }
    return nil
}

// Process actor state
func (c *Client) processActorState(ctx context.Context, info *MinerInfo, actorState map[string]interface{}) {
    if state, ok := actorState["State"].(map[string]interface{}); ok {
        if provingPeriodStart, ok := state["ProvingPeriodStart"].(float64); ok {
            info.CreateHeight = uint64(provingPeriodStart)
            // Get create timestamp
            if tipset, err := c.GetTipSetByHeight(ctx, uint64(provingPeriodStart)); err == nil {
                if blocks, ok := tipset["Blocks"].([]interface{}); ok && len(blocks) > 0 {
                    if block, ok := blocks[0].(map[string]interface{}); ok {
                        if timestamp, ok := block["Timestamp"].(float64); ok {
                            info.CreateTimestamp = int64(timestamp)
                        }
                    }
                }
            }
        }
        // Get transfer count from state
        if withdrawalBalance, ok := state["WithdrawBalance"].(string); ok && withdrawalBalance != "0" {
            info.TransferCount++
        }
    }
}

// Process power info
func (c *Client) processPowerInfo(info *MinerInfo, minerPower map[string]interface{}) {
    if rawPower, ok := minerPower["MinerPower"].(map[string]interface{}); ok {
        if raw, ok := rawPower["RawBytePower"].(string); ok {
            info.Miner.RawBytePower = raw
        }
        if quality, ok := rawPower["QualityAdjPower"].(string); ok {
            info.Miner.QualityAdjPower = quality
        }
    }
    if totalPower, ok := minerPower["TotalPower"].(map[string]interface{}); ok {
        if raw, ok := totalPower["RawBytePower"].(string); ok {
            info.Miner.NetworkRawBytePower = raw
        }
        if quality, ok := totalPower["QualityAdjPower"].(string); ok {
            info.Miner.NetworkQualityAdjPower = quality
        }
    }
}

// Process chain head info
func (c *Client) processChainHead(info *MinerInfo, chainHead map[string]interface{}) {
    if height, ok := chainHead["Height"].(float64); ok {
        info.LastSeenHeight = uint64(height)
        if blocks, ok := chainHead["Blocks"].([]interface{}); ok && len(blocks) > 0 {
            if block, ok := blocks[0].(map[string]interface{}); ok {
                if timestamp, ok := block["Timestamp"].(float64); ok {
                    info.LastSeenTimestamp = int64(timestamp)
                }
            }
        }
    }
}

// GetBalance gets the balance of an address
func (c *Client) GetBalance(ctx context.Context, address string) (string, error) {
    if address == "" {
        return "0", nil
    }

    var result string
    if err := c.callRPCWithRetry(ctx, "Filecoin.WalletBalance", []interface{}{address}, &result); err != nil {
        return "0", fmt.Errorf("failed to get balance: %w", err)
    }

    return result, nil
}

// calculatePowerShare calculates the percentage of network power
func calculatePowerShare(minerPower, totalPower string) float64 {
    if minerPower == "" || totalPower == "" {
        return 0.0
    }

    // Convert power strings to big integers
    minerBytes, ok := new(big.Int).SetString(minerPower, 10)
    if !ok {
        return 0.0
    }

    totalBytes, ok := new(big.Int).SetString(totalPower, 10)
    if !ok {
        return 0.0
    }

    if totalBytes.Sign() == 0 {
        return 0.0
    }

    // Calculate percentage: (minerBytes * 100) / totalBytes
    percentage := new(big.Float).SetInt(minerBytes)
    percentage.Mul(percentage, big.NewFloat(100))
    percentage.Quo(percentage, new(big.Float).SetInt(totalBytes))

    result, _ := percentage.Float64()
    return result
}

// formatBytes formats bytes to human readable string (KiB, MiB, GiB, TiB, PiB, EiB)
func formatBytes(bytes string) string {
    if bytes == "" {
        return "0 B"
    }

    b, ok := new(big.Int).SetString(bytes, 10)
    if !ok {
        return "0 B"
    }

    units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
    base := big.NewFloat(1024)
    size := new(big.Float).SetInt(b)

    var i int
    for i = 0; i < len(units)-1; i++ {
        next := new(big.Float).Quo(size, base)
        if next.Cmp(big.NewFloat(1)) < 0 {
            break
        }
        size = next
    }

    value, _ := size.Float64()
    return fmt.Sprintf("%.2f %s", value, units[i])
}

// formatAttoFil formats attoFIL to FIL with proper decimal places
func formatAttoFil(attoFil string) string {
    if attoFil == "" {
        return "0 FIL"
    }

    atto, ok := new(big.Int).SetString(attoFil, 10)
    if !ok {
        return "0 FIL"
    }

    // 1 FIL = 10^18 attoFIL
    fil := new(big.Float).SetInt(atto)
    fil.Quo(fil, big.NewFloat(1e18))

    value, _ := fil.Float64()
    return fmt.Sprintf("%.6f FIL", value)
}

// GetMinerStats returns formatted statistics about a miner
func (c *Client) GetMinerStats(ctx context.Context, minerID string) (map[string]string, error) {
    stats := make(map[string]string)

    info, err := c.GetComprehensiveMinerInfo(ctx, minerID)
    if err != nil {
        return nil, fmt.Errorf("failed to get miner info: %v", err)
    }

    // Format and add statistics
    stats["Raw Power"] = formatBytes(info.Miner.RawBytePower)
    stats["Quality-Adjusted Power"] = formatBytes(info.Miner.QualityAdjPower)
    stats["Network Power Share"] = fmt.Sprintf("%.4f%%", calculatePowerShare(info.Miner.RawBytePower, info.Miner.NetworkRawBytePower)*100)
    stats["Available Balance"] = formatAttoFil(info.Miner.AvailableBalance)
    stats["Total Sectors"] = fmt.Sprintf("%d", info.Miner.Sectors.Live+info.Miner.Sectors.Active+info.Miner.Sectors.Faulty+info.Miner.Sectors.Recovering)
    stats["Active Sectors"] = fmt.Sprintf("%d", info.Miner.Sectors.Active)
    stats["Faulty Sectors"] = fmt.Sprintf("%d", info.Miner.Sectors.Faulty)
    stats["Recovering Sectors"] = fmt.Sprintf("%d", info.Miner.Sectors.Recovering)
    stats["Live Sectors"] = fmt.Sprintf("%d", info.Miner.Sectors.Live)
    stats["Blocks Mined"] = fmt.Sprintf("%d", info.Miner.BlocksMined)
    stats["Weighted Blocks Mined"] = fmt.Sprintf("%d", info.Miner.WeightedBlocksMined)
    stats["Total Rewards"] = formatAttoFil(info.Miner.TotalRewards)
    stats["Quality-Adjusted Power/Sector"] = formatBytes(info.Miner.QualityAdjPower)

    return stats, nil
}

// GetActorInfo gets information about an actor
func (c *Client) GetActorInfo(ctx context.Context, address string) (map[string]interface{}, error) {
    var result map[string]interface{}
    params := []interface{}{address, nil}
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateGetActor", params, &result); err != nil {
        return nil, fmt.Errorf("failed to get actor info: %w", err)
    }
    return result, nil
}

// GetActorState gets the state of an actor
func (c *Client) GetActorState(ctx context.Context, address string) (map[string]interface{}, error) {
    var result map[string]interface{}
    params := []interface{}{address, nil}
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateReadState", params, &result); err != nil {
        return nil, fmt.Errorf("failed to get actor state: %w", err)
    }
    return result, nil
}

// GetAddressInfo gets information about an address
func (c *Client) GetAddressInfo(ctx context.Context, address string) (map[string]interface{}, error) {
    var result map[string]interface{}
    params := []interface{}{address, nil}
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateGetActor", params, &result); err != nil {
        return nil, fmt.Errorf("failed to get address info: %w", err)
    }
    return result, nil
}

// GetChainHead gets the current chain head
func (c *Client) GetChainHead(ctx context.Context) (map[string]interface{}, error) {
    var result map[string]interface{}
    if err := c.callRPCWithRetry(ctx, "Filecoin.ChainHead", nil, &result); err != nil {
        return nil, fmt.Errorf("failed to get chain head: %w", err)
    }
    return result, nil
}

// GetTipSetByHeight gets the tipset at the specified height
func (c *Client) GetTipSetByHeight(ctx context.Context, height uint64) (map[string]interface{}, error) {
    var result map[string]interface{}
    params := []interface{}{height, nil}
    if err := c.callRPCWithRetry(ctx, "Filecoin.ChainGetTipSetByHeight", params, &result); err != nil {
        return nil, fmt.Errorf("failed to get tipset: %w", err)
    }
    return result, nil
}

// GetRobustAddress gets the robust address for a given ID address
func (c *Client) GetRobustAddress(ctx context.Context, address string) (string, error) {
    var result string
    params := []interface{}{address, nil}
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateLookupRobustAddress", params, &result); err != nil {
        return "", fmt.Errorf("failed to get robust address: %w", err)
    }
    return result, nil
}

// GetMinerPledgeBalance gets the miner's sector pledge balance
func (c *Client) GetMinerPledgeBalance(ctx context.Context, minerAddr string) (string, error) {
    var result string
    params := []interface{}{minerAddr, nil}
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerPreCommitDeposit", params, &result); err != nil {
        return "", fmt.Errorf("failed to get miner pledge balance: %w", err)
    }
    return result, nil
}

// GetMinerPledgeCollateral gets the pledge collateral for a miner
func (c *Client) GetMinerPledgeCollateral(ctx context.Context, minerAddr string) (string, error) {
    var result string
    params := []interface{}{minerAddr, nil}
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerInitialPledgeCollateral", params, &result); err != nil {
        return "", fmt.Errorf("failed to get miner pledge collateral: %w", err)
    }
    return result, nil
}

// GetMinerPowerRank gets the miner's power rank information
func (c *Client) GetMinerPowerRank(ctx context.Context, minerAddr string) (uint64, uint64, error) {
    var result struct {
        MinerPower struct {
            RawBytePower    string
            QualityAdjPower string
        }
        TotalPower struct {
            RawBytePower    string
            QualityAdjPower string
        }
    }

    if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerPower", []interface{}{minerAddr, nil}, &result); err != nil {
        return 0, 0, fmt.Errorf("failed to get miner power: %w", err)
    }

    // Get all miners power
    var allMiners []string
    if err := c.callRPCWithRetry(ctx, "Filecoin.StateListMiners", []interface{}{nil}, &allMiners); err != nil {
        return 0, 0, fmt.Errorf("failed to list miners: %w", err)
    }

    type minerPower struct {
        miner    string
        rawBytes string
        adjPower string
    }

    var powers []minerPower
    for _, miner := range allMiners {
        var mp struct {
            MinerPower struct {
                RawBytePower    string
                QualityAdjPower string
            }
        }
        if err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerPower", []interface{}{miner, nil}, &mp); err != nil {
            continue
        }
        powers = append(powers, minerPower{
            miner:    miner,
            rawBytes: mp.MinerPower.RawBytePower,
            adjPower: mp.MinerPower.QualityAdjPower,
        })
    }

    // Sort by raw byte power
    sort.Slice(powers, func(i, j int) bool {
        return powers[i].rawBytes > powers[j].rawBytes
    })
    rawRank := uint64(1)
    for i, p := range powers {
        if p.miner == minerAddr {
            rawRank = uint64(i + 1)
            break
        }
    }

    // Sort by quality adjusted power
    sort.Slice(powers, func(i, j int) bool {
        return powers[i].adjPower > powers[j].adjPower
    })
    adjRank := uint64(1)
    for i, p := range powers {
        if p.miner == minerAddr {
            adjRank = uint64(i + 1)
            break
        }
    }

    return rawRank, adjRank, nil
}
