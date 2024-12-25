// Copyright (c) 2024 THCloud.AI
// Author: OC
// Last Updated: 2024-12-25
// Description: Lotus API types.

package lotus

// Config represents Lotus client configuration
type Config struct {
	APIAddress string `mapstructure:"api_address"`
	AuthToken  string `mapstructure:"auth_token"`
}

// SectorPenaltyResponse represents the response for sector penalty calculation
type SectorPenaltyResponse struct {
	Penalty string `json:"penalty"`
}

// VestedFundsResponse represents the response for vested funds query
type VestedFundsResponse struct {
	Vested string `json:"vested"`
}
