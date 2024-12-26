// Copyright (c) 2024 THCloud.AI
// Author: OC
// Last Updated: 2024-12-26
// Description: Lotus API types.

package lotus

// SectorInfo represents the information for a sector
type SectorInfo struct {
	SectorNumber uint64 `json:"sector_number"`
	SealProof    string `json:"seal_proof"`
	Activation   uint64 `json:"activation"`
	Expiration   uint64 `json:"expiration"`
}

// SectorStatus represents the status of a sector
type SectorStatus struct {
	Status string `json:"status"`
}
