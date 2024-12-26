package miner

import (
    "context"
    "fmt"
    "math/big"
    "strings"
    "time"

    "github.com/spf13/cobra"
    "github.com/THCloudAI/thctl/internal/lotus"
)

func NewMinerCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "miner [minerID]",
        Short: "Get miner information",
        Long: `Get comprehensive information about a Filecoin miner, including:
- Basic information (addresses, sector size, proof type)
- Power statistics (raw power, quality adjusted power, network share)
- Financial information (balance, pledges, deposits)
- Sector information (total, active, faulty sectors)
- Deadline information and schedule
- Performance metrics`,
        Args: cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            minerID := args[0]
            return getMinerInfo(minerID)
        },
    }

    // Add flags
    cmd.PersistentFlags().StringP("output", "o", "json", "Output format: json, yaml, or table")

    return cmd
}

func getMinerInfo(minerID string) error {
    // Create Lotus client from environment
    client := lotus.NewFromEnv()
    if client == nil {
        return fmt.Errorf("❌ failed to create Lotus client")
    }

    // Get comprehensive miner info
    info, err := client.GetComprehensiveMinerInfo(context.Background(), minerID)
    if err != nil {
        // Check specific error types
        switch {
        case lotus.IsNotFound(err):
            return fmt.Errorf("❌ miner %s not found", minerID)
        case lotus.IsConnectionError(err):
            return fmt.Errorf("❌ connection error: failed to connect to Lotus node")
        case lotus.IsAuthError(err):
            return fmt.Errorf("❌ authentication error: invalid or missing API token")
        default:
            return fmt.Errorf("❌ error getting miner info: %v", err)
        }
    }

    // Display miner information in sections
    fmt.Printf("\n🔍 Miner Information for %s\n", minerID)
    fmt.Println(strings.Repeat("-", 50))

    // Basic Information
    fmt.Println("\n📋 Basic Information:")
    fmt.Printf("   Sector Size: %s\n", formatBytes(fmt.Sprintf("%d", info.SectorSize)))
    fmt.Printf("   Window PoSt Proof Type: %d\n", info.WindowPoStProofType)
    fmt.Printf("   Owner Address: %s\n", info.OwnerAddress)
    fmt.Printf("   Worker Address: %s\n", info.WorkerAddress)
    fmt.Printf("   Beneficiary: %s\n", info.Beneficiary)
    if len(info.ControlAddresses) > 0 {
        fmt.Println("   Control Addresses:")
        for _, addr := range info.ControlAddresses {
            fmt.Printf("      - %s\n", addr)
        }
    }

    // Power Information
    fmt.Println("\n💪 Power Statistics:")
    fmt.Printf("   Raw Power: %s\n", formatBytes(info.RawBytePower))
    fmt.Printf("   Quality Adjusted Power: %s\n", formatBytes(info.QualityAdjPower))
    fmt.Printf("   Network Power Share: %.4f%%\n", info.NetworkPowerShare*100)
    if info.QualityAdjPowerPerSector != "" {
        fmt.Printf("   QAP Per Sector: %s\n", formatBytes(info.QualityAdjPowerPerSector))
    }
    if info.ConsensusMiners > 0 {
        fmt.Printf("   Total Consensus Miners: %d\n", info.ConsensusMiners)
    }

    // Financial Information
    fmt.Println("\n💰 Financial Information:")
    fmt.Printf("   Available Balance: %s\n", formatAttoFil(info.AvailableBalance))
    fmt.Printf("   Initial Pledge: %s\n", formatAttoFil(info.InitialPledge))
    fmt.Printf("   Pre-Commit Deposits: %s\n", formatAttoFil(info.PreCommitDeposits))
    fmt.Printf("   Vesting Funds: %s\n", formatAttoFil(info.VestingFunds))
    fmt.Printf("   Total Locked: %s\n", formatAttoFil(info.TotalLocked))

    // Sector Information
    fmt.Println("\n📊 Sector Statistics:")
    fmt.Printf("   Total Sectors: %d\n", info.TotalSectors)
    fmt.Printf("   Active Sectors: %d\n", info.ActiveSectors)
    fmt.Printf("   Faulty Sectors: %d\n", len(info.FaultySectors))
    fmt.Printf("   Recovering Sectors: %d\n", len(info.RecoveringSectors))
    if len(info.LiveSectors) > 0 {
        fmt.Printf("   Live Sectors: %d\n", len(info.LiveSectors))
    }
    if len(info.FaultySectors) > 0 {
        faultRate := float64(len(info.FaultySectors)) / float64(info.TotalSectors) * 100
        fmt.Printf("   ⚠️  Fault Rate: %.2f%%\n", faultRate)
        fmt.Println("   Faulty Sector Numbers:")
        for i, sector := range info.FaultySectors {
            if i < 5 { // Show only first 5 faulty sectors to avoid cluttering
                fmt.Printf("      - %d\n", sector)
            } else {
                fmt.Printf("      ... and %d more\n", len(info.FaultySectors)-5)
                break
            }
        }
    }

    // Deadline Information
    fmt.Println("\n⏰ Deadline Information:")
    fmt.Printf("   Current Epoch: %d\n", info.CurrentEpoch)
    fmt.Printf("   Current Deadline: %d\n", info.CurrentDeadline)
    fmt.Printf("   Proving Period Start: %d\n", info.ProvingPeriodStart)
    
    // Calculate and display time-based information
    epochDuration := 30 * time.Second // Filecoin epoch duration
    currentTime := time.Now()
    
    deadlineOpen := currentTime.Add(time.Duration(info.CurrentDeadline) * epochDuration)
    deadlineClose := deadlineOpen.Add(30 * time.Minute) // Default deadline window is 30 minutes
    
    fmt.Printf("   Current Time: %s\n", currentTime.Format(time.RFC3339))
    fmt.Printf("   Deadline Window Opens: %s\n", deadlineOpen.Format(time.RFC3339))
    fmt.Printf("   Deadline Window Closes: %s\n", deadlineClose.Format(time.RFC3339))
    
    timeUntilOpen := deadlineOpen.Sub(currentTime)
    if timeUntilOpen > 0 {
        fmt.Printf("   Time Until Window Opens: %.1f minutes\n", timeUntilOpen.Minutes())
    }

    // Performance Metrics
    if info.MinerUptime > 0 {
        fmt.Println("\n📈 Performance Metrics:")
        fmt.Printf("   Miner Uptime: %.2f%%\n", info.MinerUptime*100)
        
        // Calculate and display additional metrics
        successRate := float64(info.ActiveSectors) / float64(info.TotalSectors) * 100
        fmt.Printf("   Sector Success Rate: %.2f%%\n", successRate)
        
        if len(info.FaultySectors) > 0 {
            avgRecoveryTime := float64(len(info.RecoveringSectors)) / float64(len(info.FaultySectors)) * 100
            fmt.Printf("   Recovery Progress: %.2f%%\n", avgRecoveryTime)
        }
    }

    fmt.Println(strings.Repeat("-", 50))
    return nil
}

// formatBytes formats bytes into human readable format
func formatBytes(bytes string) string {
    // Convert string to big.Int
    b := new(big.Int)
    b.SetString(bytes, 10)

    // Define units
    units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
    unitSize := new(big.Int).SetInt64(1024)
    unit := 0
    zero := new(big.Int).SetInt64(0)
    threshold := new(big.Int).SetInt64(1024)

    // Find appropriate unit
    value := new(big.Int).Set(b)
    for value.Cmp(threshold) >= 0 && unit < len(units)-1 {
        value.Div(value, unitSize)
        unit++
    }

    if value.Cmp(zero) == 0 {
        return "0 B"
    }

    return fmt.Sprintf("%s %s", value.String(), units[unit])
}

// formatAttoFil formats attoFIL to FIL with proper decimal places
func formatAttoFil(attoFil string) string {
    // Convert string to big.Int
    atto := new(big.Int)
    atto.SetString(attoFil, 10)

    // Define FIL = 10^18 attoFIL
    fil := new(big.Int).SetInt64(1000000000000000000)

    // Calculate FIL amount
    filAmount := new(big.Float).SetInt(atto)
    filAmount.Quo(filAmount, new(big.Float).SetInt(fil))

    // Format with 6 decimal places
    return fmt.Sprintf("%.6f FIL", filAmount)
}