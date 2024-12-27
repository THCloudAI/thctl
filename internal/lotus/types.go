package lotus

import (
	"encoding/json"
)

// RPCRequest represents a JSON-RPC request
type RPCRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

// RPCResponse represents a JSON-RPC response
type RPCResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *RPCError       `json:"error"`
	ID      int            `json:"id"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Response represents a standardized API response
type Response struct {
	Version   string      `json:"version"`
	Timestamp int64      `json:"timestamp"`
	Status    string     `json:"status"`
	Data      interface{} `json:"data"`
}

// MinerInfo represents comprehensive information about a miner
type MinerInfo struct {
	ID                 string   `json:"id"`
	Robust             string   `json:"robust"`
	Actor              string   `json:"actor"`
	CreateHeight       uint64   `json:"createHeight"`
	CreateTimestamp    int64    `json:"createTimestamp"`
	LastSeenHeight     uint64   `json:"lastSeenHeight"`
	LastSeenTimestamp  int64    `json:"lastSeenTimestamp"`
	Balance            string   `json:"balance"`
	MessageCount       uint64   `json:"messageCount"`
	TransferCount      uint64   `json:"transferCount"`
	TokenTransferCount uint64   `json:"tokenTransferCount"`
	Tokens            uint64   `json:"tokens"`
	Miner             struct {
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
		ControlAddresses     []ControlAddress `json:"controlAddresses"`
		PeerID              string           `json:"peerId"`
		MultiAddresses      []string         `json:"multiAddresses"`
		SectorSize          uint64           `json:"sectorSize"`
		RawBytePower        string           `json:"rawBytePower"`
		QualityAdjPower     string           `json:"qualityAdjPower"`
		NetworkRawBytePower string           `json:"networkRawBytePower"`
		NetworkQualityAdjPower string        `json:"networkQualityAdjPower"`
		BlocksMined          uint64          `json:"blocksMined"`
		WeightedBlocksMined  uint64          `json:"weightedBlocksMined"`
		TotalRewards        string           `json:"totalRewards"`
		Sectors             struct {
			Live       uint64 `json:"live"`
			Active     uint64 `json:"active"`
			Faulty     uint64 `json:"faulty"`
			Recovering uint64 `json:"recovering"`
		} `json:"sectors"`
		PreCommitDeposits        string `json:"preCommitDeposits"`
		VestingFunds            string `json:"vestingFunds"`
		InitialPledgeRequirement string `json:"initialPledgeRequirement"`
		AvailableBalance        string `json:"availableBalance"`
		SectorPledgeBalance     string `json:"sectorPledgeBalance"`
		PledgeBalance           string `json:"pledgeBalance"`
		RawBytePowerRank        uint64 `json:"rawBytePowerRank"`
		QualityAdjPowerRank     uint64 `json:"qualityAdjPowerRank"`
	} `json:"miner"`
	OwnedMiners     []string `json:"ownedMiners"`
	WorkerMiners    []string `json:"workerMiners"`
	BenefitedMiners []string `json:"benefitedMiners"`
	Address         string   `json:"address"`
}

// ControlAddress represents a control address with its balance
type ControlAddress struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
	ID      string `json:"id,omitempty"`
}

// SectorInfo represents information about a sector
type SectorInfo struct {
	SectorNumber uint64 `json:"sectorNumber"`
	State        string `json:"state"`
	SealedCID    string `json:"sealedCid"`
	Deals        []struct {
		DealID      uint64 `json:"dealId"`
		PieceCID    string `json:"pieceCid"`
		PieceSize   uint64 `json:"pieceSize"`
		Client      string `json:"client"`
		StartEpoch  uint64 `json:"startEpoch"`
		EndEpoch    uint64 `json:"endEpoch"`
		StorageFee  string `json:"storageFee"`
		PieceStatus string `json:"pieceStatus"`
	} `json:"deals"`
	CreationTime      int64  `json:"creationTime"`
	ExpirationTime    int64  `json:"expirationTime"`
	DealWeight        string `json:"dealWeight"`
	VerifiedWeight    string `json:"verifiedWeight"`
	InitialPledge     string `json:"initialPledge"`
	ExpectedDayReward string `json:"expectedDayReward"`
	ExpectedStoragePledge string `json:"expectedStoragePledge"`
}

// SectorPenalty represents penalty information for a sector
type SectorPenalty struct {
	SectorNumber uint64 `json:"sectorNumber"`
	Penalty      string `json:"penalty"`
	Reason       string `json:"reason"`
	Timestamp    int64  `json:"timestamp"`
}

// SectorVested represents vesting information for a sector
type SectorVested struct {
	SectorNumber uint64 `json:"sectorNumber"`
	VestedFunds  string `json:"vestedFunds"`
	Timestamp    int64  `json:"timestamp"`
}
