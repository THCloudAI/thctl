package miner

import (
    "encoding/json"
    "fmt"
    "math/big"
    "strings"
    "time"
    "github.com/jedib0t/go-pretty/v6/table"
    "github.com/spf13/cobra"
    "github.com/THCloudAI/thctl/internal/lotus"
    "gopkg.in/yaml.v3"
)

func NewMinerCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "miner [minerID]",
        Short: "Get miner information",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            minerID := args[0]
            output, _ := cmd.Flags().GetString("output")

            client, err := lotus.NewFromEnv()
            if err != nil {
                return fmt.Errorf("❌ failed to create Lotus client: %v", err)
            }

            info, err := client.GetComprehensiveMinerInfo(cmd.Context(), minerID)
            if err != nil {
                return fmt.Errorf("❌ error getting miner info: %v", err)
            }

            // If any required fields are missing, return an error
            if info.Miner.Owner.Address == "" || info.Miner.Worker.Address == "" {
                return fmt.Errorf("❌ failed to get required miner information")
            }

            // Create standardized response
            resp := &lotus.Response{
                Version:   "1.0",
                Timestamp: time.Now().Unix(),
                Status:    "success",
                Data:     info,
            }

            switch output {
            case "json":
                jsonBytes, err := json.MarshalIndent(resp, "", "  ")
                if err != nil {
                    return fmt.Errorf("❌ error marshaling JSON: %v", err)
                }
                fmt.Println(string(jsonBytes))
            case "yaml":
                yamlBytes, err := yaml.Marshal(resp)
                if err != nil {
                    return fmt.Errorf("❌ error marshaling YAML: %v", err)
                }
                fmt.Println(string(yamlBytes))
            case "table":
                printMinerInfoTable(minerID, info)
            default:
                return fmt.Errorf("❌ unsupported output format: %s", output)
            }

            return nil
        },
    }

    cmd.Flags().StringP("output", "o", "json", "Output format: json, yaml, or table")
    return cmd
}

func printMinerInfoTable(minerID string, info *lotus.MinerInfo) {
    fmt.Printf("\n🔍 Miner Information for %s\n", minerID)
    fmt.Println(strings.Repeat("-", 50))

    // Basic Information
    fmt.Println("\n📋 Basic Information:")
    t := table.NewWriter()
    t.AppendHeader(table.Row{"Attribute", "Value"})
    t.AppendRow(table.Row{"ID", info.ID})
    t.AppendRow(table.Row{"Robust Address", info.Robust})
    t.AppendRow(table.Row{"Actor Type", info.Actor})
    t.AppendRow(table.Row{"Balance", formatFIL(info.Balance)})
    t.AppendRow(table.Row{"Create Height", fmt.Sprintf("%d", info.CreateHeight)})
    t.AppendRow(table.Row{"Create Time", time.Unix(info.CreateTimestamp, 0).Format(time.RFC3339)})
    t.AppendRow(table.Row{"Last Seen Height", fmt.Sprintf("%d", info.LastSeenHeight)})
    t.AppendRow(table.Row{"Last Seen Time", time.Unix(info.LastSeenTimestamp, 0).Format(time.RFC3339)})
    t.AppendRow(table.Row{"Message Count", fmt.Sprintf("%d", info.MessageCount)})
    t.AppendRow(table.Row{"Transfer Count", fmt.Sprintf("%d", info.TransferCount)})
    t.AppendRow(table.Row{"Token Transfer Count", fmt.Sprintf("%d", info.TokenTransferCount)})
    t.AppendRow(table.Row{"Tokens", fmt.Sprintf("%d", info.Tokens)})
    fmt.Println(t.Render())

    // Address Information
    fmt.Println("\n📫 Address Information:")
    t = table.NewWriter()
    t.AppendHeader(table.Row{"Role", "Address", "Balance"})
    t.AppendRow(table.Row{"Owner", info.Miner.Owner.Address, formatFIL(info.Miner.Owner.Balance)})
    t.AppendRow(table.Row{"Worker", info.Miner.Worker.Address, formatFIL(info.Miner.Worker.Balance)})
    t.AppendRow(table.Row{"Beneficiary", info.Miner.Beneficiary.Address, formatFIL(info.Miner.Beneficiary.Balance)})
    for i, ctrl := range info.Miner.ControlAddresses {
        t.AppendRow(table.Row{fmt.Sprintf("Control %d", i+1), ctrl.Address, formatFIL(ctrl.Balance)})
    }
    fmt.Println(t.Render())

    // Power Statistics
    fmt.Println("\n💪 Power Statistics:")
    t = table.NewWriter()
    t.AppendHeader(table.Row{"Attribute", "Value"})
    t.AppendRow(table.Row{"Raw Power", formatBytes(info.Miner.RawBytePower)})
    t.AppendRow(table.Row{"Quality Adjusted Power", formatBytes(info.Miner.QualityAdjPower)})
    t.AppendRow(table.Row{"Network Raw Power", formatBytes(info.Miner.NetworkRawBytePower)})
    t.AppendRow(table.Row{"Network Quality Power", formatBytes(info.Miner.NetworkQualityAdjPower)})
    t.AppendRow(table.Row{"Network Power Share", fmt.Sprintf("%.4f%%", calculatePowerShare(info.Miner.RawBytePower, info.Miner.NetworkRawBytePower)*100)})
    t.AppendRow(table.Row{"Raw Power Rank", fmt.Sprintf("%d", info.Miner.RawBytePowerRank)})
    t.AppendRow(table.Row{"Quality Power Rank", fmt.Sprintf("%d", info.Miner.QualityAdjPowerRank)})
    fmt.Println(t.Render())

    // Financial Information
    fmt.Println("\n💰 Financial Information:")
    t = table.NewWriter()
    t.AppendHeader(table.Row{"Attribute", "Value"})
    t.AppendRow(table.Row{"Available Balance", formatFIL(info.Miner.AvailableBalance)})
    t.AppendRow(table.Row{"Initial Pledge", formatFIL(info.Miner.InitialPledgeRequirement)})
    t.AppendRow(table.Row{"Vesting Funds", formatFIL(info.Miner.VestingFunds)})
    t.AppendRow(table.Row{"Pre-Commit Deposits", formatFIL(info.Miner.PreCommitDeposits)})
    t.AppendRow(table.Row{"Total Rewards", formatFIL(info.Miner.TotalRewards)})
    t.AppendRow(table.Row{"Sector Pledge Balance", formatFIL(info.Miner.SectorPledgeBalance)})
    t.AppendRow(table.Row{"Pledge Balance", formatFIL(info.Miner.PledgeBalance)})
    fmt.Println(t.Render())

    // Sector Statistics
    fmt.Println("\n📊 Sector Statistics:")
    t = table.NewWriter()
    t.AppendHeader(table.Row{"Attribute", "Value"})
    t.AppendRow(table.Row{"Live Sectors", fmt.Sprintf("%d", info.Miner.Sectors.Live)})
    t.AppendRow(table.Row{"Active Sectors", fmt.Sprintf("%d", info.Miner.Sectors.Active)})
    t.AppendRow(table.Row{"Faulty Sectors", fmt.Sprintf("%d", info.Miner.Sectors.Faulty)})
    t.AppendRow(table.Row{"Recovering Sectors", fmt.Sprintf("%d", info.Miner.Sectors.Recovering)})
    fmt.Println(t.Render())

    // Mining Statistics
    fmt.Println("\n⛏️ Mining Statistics:")
    t = table.NewWriter()
    t.AppendHeader(table.Row{"Attribute", "Value"})
    t.AppendRow(table.Row{"Blocks Mined", fmt.Sprintf("%d", info.Miner.BlocksMined)})
    t.AppendRow(table.Row{"Weighted Blocks", fmt.Sprintf("%d", info.Miner.WeightedBlocksMined)})
    fmt.Println(t.Render())

    fmt.Println(strings.Repeat("-", 50))
}

func calculatePowerShare(power, networkPower string) float64 {
    // Convert string to big.Int
    p := new(big.Int)
    p.SetString(power, 10)

    // Convert string to big.Int
    np := new(big.Int)
    np.SetString(networkPower, 10)

    // Calculate power share
    share := new(big.Float).SetInt(p)
    share.Quo(share, new(big.Float).SetInt(np))

    // Return power share as float64
    result, _ := share.Float64()
    return result
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
func formatFIL(attoFil string) string {
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