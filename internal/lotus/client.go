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
    "strings"
    "time"

    ma "github.com/multiformats/go-multiaddr"
    "golang.org/x/sync/errgroup"
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
    if cfg.APIURL == "" {
        cfg.APIURL = "http://127.0.0.1:1234/rpc/v0"
    }
    if cfg.Timeout == 0 {
        cfg.Timeout = defaultTimeout
    }

    return &Client{
        apiURL: cfg.APIURL,
        token:  cfg.AuthToken,
        httpClient: &http.Client{
            Timeout: cfg.Timeout,
        },
    }
}

// NewFromEnv creates a new Lotus client from environment variables
func NewFromEnv() *Client {
    config := &Config{}
    if url := os.Getenv("LOTUS_API_URL"); url != "" {
        config.APIURL = url
    }
    if token := os.Getenv("LOTUS_API_TOKEN"); token != "" {
        config.AuthToken = token
    }

    return New(*config)
}

// Client represents a Lotus API client
type Client struct {
    apiURL     string
    token      string
    httpClient *http.Client
}

// callRPC makes a JSON-RPC call to the Lotus API
func (c *Client) callRPC(ctx context.Context, method string, params interface{}, result interface{}) error {
    // Convert multiaddr URL to HTTP URL if necessary
    apiURL := c.apiURL
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
    if c.token != "" {
        req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
    }

    // Make request
    resp, err := c.httpClient.Do(req)
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

// callRPCWithRetry makes a JSON-RPC call to the Lotus API with retry
func (c *Client) callRPCWithRetry(ctx context.Context, method string, params interface{}, result interface{}) error {
    var err error
    for i := 0; i < 3; i++ {
        err = c.callRPC(ctx, method, params, result)
        if err == nil {
            break
        }
    }
    return err
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
    err := c.callRPCWithRetry(ctx, "Filecoin.StateMinerInfo", []interface{}{minerID, nil}, &result)
    if err != nil {
        return nil, fmt.Errorf("failed to get miner info: %w", err)
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
func (c *Client) GetMinerDeadlines(ctx context.Context, minerID string) (map[string]interface{}, error) {
    var result map[string]interface{}
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
        return "", err
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

// MinerInfo represents comprehensive information about a miner
type MinerInfo struct {
    // Basic Information
    BasicInfo         map[string]interface{} `json:"basic_info"`
    SectorSize       uint64                 `json:"sector_size"`
    WindowPoStProofType uint64              `json:"window_post_proof_type"`
    
    // Addresses
    WorkerAddress    string                 `json:"worker_address"`
    OwnerAddress     string                 `json:"owner_address"`
    Beneficiary      string                 `json:"beneficiary"`
    ControlAddresses []string               `json:"control_addresses"`
    
    // Power Information
    Power            map[string]interface{} `json:"power"`
    RawBytePower     string                 `json:"raw_byte_power"`
    QualityAdjPower  string                 `json:"quality_adj_power"`
    NetworkPowerShare float64               `json:"network_power_share"`
    
    // Financial Information
    AvailableBalance string                 `json:"available_balance"`
    InitialPledge    string                 `json:"initial_pledge"`
    PreCommitDeposits string                `json:"pre_commit_deposits"`
    VestingFunds     string                 `json:"vesting_funds"`
    TotalLocked      string                 `json:"total_locked"`
    
    // Sector Information
    TotalSectors     uint64                 `json:"total_sectors"`
    ActiveSectors    uint64                 `json:"active_sectors"`
    FaultySectors    []uint64               `json:"faulty_sectors"`
    RecoveringSectors []uint64              `json:"recovering_sectors"`
    LiveSectors      []uint64               `json:"live_sectors"`
    
    // Deadline Information
    CurrentDeadline  uint64                 `json:"current_deadline"`
    CurrentEpoch     uint64                 `json:"current_epoch"`
    ProvingPeriodStart uint64              `json:"proving_period_start"`
    Deadlines        map[string]interface{} `json:"deadlines"`
    ProvingDeadline  map[string]interface{} `json:"proving_deadline"`
    
    // Performance Metrics
    QualityAdjPowerPerSector string         `json:"quality_adj_power_per_sector"`
    ConsensusMiners         uint64          `json:"consensus_miners"`
    MinerUptime            float64         `json:"miner_uptime"`
}

// GetComprehensiveMinerInfo gets all available information about a miner
func (c *Client) GetComprehensiveMinerInfo(ctx context.Context, minerID string) (*MinerInfo, error) {
    if minerID == "" {
        return nil, errors.New("miner ID cannot be empty")
    }

    var info MinerInfo
    g, ctx := errgroup.WithContext(ctx)

    // Get basic miner info
    g.Go(func() error {
        basicInfo, err := c.GetMinerInfo(ctx, minerID)
        if err != nil {
            return fmt.Errorf("failed to get basic info: %w", err)
        }
        info.BasicInfo = basicInfo
        
        // Extract specific fields from basic info
        if sectorSize, ok := basicInfo["SectorSize"].(uint64); ok {
            info.SectorSize = sectorSize
        }
        if proofType, ok := basicInfo["WindowPoStProofType"].(uint64); ok {
            info.WindowPoStProofType = proofType
        }
        if controlAddrs, ok := basicInfo["ControlAddresses"].([]string); ok {
            info.ControlAddresses = controlAddrs
        }
        return nil
    })

    // Get power information
    g.Go(func() error {
        power, err := c.GetMinerPower(ctx, minerID)
        if err != nil {
            return fmt.Errorf("failed to get power info: %w", err)
        }
        info.Power = power
        
        // Extract and calculate power metrics
        if minerPower, ok := power["MinerPower"].(map[string]interface{}); ok {
            info.RawBytePower = minerPower["RawBytePower"].(string)
            info.QualityAdjPower = minerPower["QualityAdjPower"].(string)
            
            // Calculate network power share
            if totalPower, ok := power["TotalPower"].(map[string]interface{}); ok {
                if totalRaw, ok := totalPower["RawBytePower"].(string); ok {
                    info.NetworkPowerShare = calculatePowerShare(info.RawBytePower, totalRaw)
                }
            }
        }
        return nil
    })

    // Get financial information
    g.Go(func() error {
        balance, err := c.GetMinerAvailableBalance(ctx, minerID)
        if err != nil {
            return fmt.Errorf("failed to get available balance: %w", err)
        }
        info.AvailableBalance = balance

        // Get initial pledge
        var initialPledge interface{}
        err = c.callRPCWithRetry(ctx, "Filecoin.StateMinerInitialPledgeCollateral", []interface{}{minerID, nil}, &initialPledge)
        if err != nil {
            return fmt.Errorf("failed to get initial pledge: %w", err)
        }
        if pledge, ok := initialPledge.(string); ok {
            info.InitialPledge = pledge
        }

        // Get pre-commit deposits
        var preCommitDeposits interface{}
        err = c.callRPCWithRetry(ctx, "Filecoin.StateMinerPreCommitDeposit", []interface{}{minerID, nil}, &preCommitDeposits)
        if err != nil {
            return fmt.Errorf("failed to get pre-commit deposits: %w", err)
        }
        if deposits, ok := preCommitDeposits.(string); ok {
            info.PreCommitDeposits = deposits
        }

        // Get vesting funds
        var vestingFunds interface{}
        err = c.callRPCWithRetry(ctx, "Filecoin.StateMinerVestingFunds", []interface{}{minerID, nil}, &vestingFunds)
        if err != nil {
            return fmt.Errorf("failed to get vesting funds: %w", err)
        }
        if vesting, ok := vestingFunds.(string); ok {
            info.VestingFunds = vesting
        }

        return nil
    })

    // Get sector information
    g.Go(func() error {
        // Get faulty sectors
        faults, err := c.GetMinerFaults(ctx, minerID)
        if err != nil {
            return fmt.Errorf("failed to get faults: %w", err)
        }
        info.FaultySectors = faults

        // Get recovering sectors
        var recoveringSectors interface{}
        err = c.callRPCWithRetry(ctx, "Filecoin.StateMinerRecoveringSectors", []interface{}{minerID, nil}, &recoveringSectors)
        if err != nil {
            return fmt.Errorf("failed to get recovering sectors: %w", err)
        }
        if recovering, ok := recoveringSectors.([]uint64); ok {
            info.RecoveringSectors = recovering
        }

        // Get live sectors
        var liveSectors interface{}
        err = c.callRPCWithRetry(ctx, "Filecoin.StateMinerActiveSectors", []interface{}{minerID, nil}, &liveSectors)
        if err != nil {
            return fmt.Errorf("failed to get live sectors: %w", err)
        }
        if live, ok := liveSectors.([]uint64); ok {
            info.LiveSectors = live
            info.TotalSectors = uint64(len(live))
            info.ActiveSectors = info.TotalSectors - uint64(len(info.FaultySectors))
        }

        return nil
    })

    // Get deadline information
    g.Go(func() error {
        deadlines, err := c.GetMinerDeadlines(ctx, minerID)
        if err != nil {
            return fmt.Errorf("failed to get deadlines: %w", err)
        }
        info.Deadlines = deadlines

        provingDeadline, err := c.GetMinerProvingDeadline(ctx, minerID)
        if err != nil {
            return fmt.Errorf("failed to get proving deadline: %w", err)
        }
        info.ProvingDeadline = provingDeadline

        if epoch, ok := provingDeadline["CurrentEpoch"].(uint64); ok {
            info.CurrentEpoch = epoch
        }
        if deadline, ok := provingDeadline["DeadlineIndex"].(uint64); ok {
            info.CurrentDeadline = deadline
        }
        if start, ok := provingDeadline["PeriodStart"].(uint64); ok {
            info.ProvingPeriodStart = start
        }

        return nil
    })

    // Get addresses
    g.Go(func() error {
        worker, err := c.GetMinerWorkerAddress(ctx, minerID)
        if err != nil {
            return fmt.Errorf("failed to get worker address: %w", err)
        }
        info.WorkerAddress = worker

        owner, err := c.GetMinerOwnerAddress(ctx, minerID)
        if err != nil {
            return fmt.Errorf("failed to get owner address: %w", err)
        }
        info.OwnerAddress = owner

        beneficiary, err := c.GetMinerBeneficiaryAddress(ctx, minerID)
        if err != nil {
            return fmt.Errorf("failed to get beneficiary address: %w", err)
        }
        info.Beneficiary = beneficiary

        return nil
    })

    if err := g.Wait(); err != nil {
        return &info, fmt.Errorf("failed to get comprehensive miner info: %w", err)
    }

    return &info, nil
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
    stats["Raw Power"] = formatBytes(info.RawBytePower)
    stats["Quality-Adjusted Power"] = formatBytes(info.QualityAdjPower)
    stats["Network Power Share"] = fmt.Sprintf("%.4f%%", info.NetworkPowerShare*100)
    stats["Available Balance"] = formatAttoFil(info.AvailableBalance)
    stats["Initial Pledge"] = formatAttoFil(info.InitialPledge)
    stats["Pre-Commit Deposits"] = formatAttoFil(info.PreCommitDeposits)
    stats["Vesting Funds"] = formatAttoFil(info.VestingFunds)
    stats["Total Locked"] = formatAttoFil(info.TotalLocked)
    stats["Total Sectors"] = fmt.Sprintf("%d", info.TotalSectors)
    stats["Active Sectors"] = fmt.Sprintf("%d", info.ActiveSectors)
    stats["Faulty Sectors"] = fmt.Sprintf("%d", len(info.FaultySectors))
    stats["Recovering Sectors"] = fmt.Sprintf("%d", len(info.RecoveringSectors))
    stats["Live Sectors"] = fmt.Sprintf("%d", len(info.LiveSectors))
    stats["Current Deadline"] = fmt.Sprintf("%d", info.CurrentDeadline)
    stats["Current Epoch"] = fmt.Sprintf("%d", info.CurrentEpoch)
    stats["Proving Period Start"] = fmt.Sprintf("%d", info.ProvingPeriodStart)
    stats["Quality-Adjusted Power/Sector"] = formatBytes(info.QualityAdjPowerPerSector)
    stats["Consensus Miners"] = fmt.Sprintf("%d", info.ConsensusMiners)
    stats["Miner Uptime"] = fmt.Sprintf("%.2f%%", info.MinerUptime*100)

    return stats, nil
}
