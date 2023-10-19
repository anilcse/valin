package main

import (
	"time"

	lens "github.com/strangelove-ventures/lens/client"
)

// NetworkConfig represents the configuration for each network
type NetworkConfig struct {
	Binary    string `json:"binary"`
	Granter   string `json:"granter"`
	Grantee   string `json:"grantee"`
	ChainName string `json:"chain_name"`
	ChainID   string `json:"chain_id"`
	Node      string `json:"node"`
	FeePayer  string `json:"feepayer"`
	Validator string `json:"validator"`
}

// IncomeData represents income details
type IncomeData struct {
	ChainID    string    `json:"chain_id"`
	Granter    string    `json:"granter"`
	OldBalance string    `json:"old_balance"`
	Income     string    `json:"income"`
	NewBalance string    `json:"new_balance"`
	Date       time.Time `json:"date"`
}

// Configuration represents the structure of the config.json file
type Configuration struct {
	CronJobTime string `json:"cronJobTime"`
	Mnemonic    string `json:"mnemonic"`
	DB          struct {
		Driver string `json:"driver"`
		Table  string `json:"table"`
		SQLURL string `json:"sqlUrl"`
	} `json:"db"`
	Networks []NetworkConfig `json:"networks"`
}

var ChainClients map[string]lens.ChainClients
