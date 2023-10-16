package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/robfig/cron" // Import the cron library
)

// NetworkConfig represents the configuration for each network
type NetworkConfig struct {
	Binary    string `json:"binary"`
	Granter   string `json:"granter"`
	Grantee   string `json:"grantee"`
	ChainID   string `json:"chainId"`
	Node      string `json:"node"`
	FeePayer  string `json:"feepayer"`
	SQLTable  string `json:"sqlTable"`
	SQLDriver string `json:"sqlDriver"`
	SQLURL    string `json:"sqlURL"`
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

func main() {
	configFileName := "config.json"
	db := connectToDatabase("username:password@tcp(database-server:port)/database-name")
	startHTTPServer(db, configFileName)
	// Schedule the rewards withdrawal to run every month on the 1st at 9 PM
	cronSchedule := "0 21 1 * *"
	startRewardsWithdrawalCron(db, configFileName, cronSchedule)
}

func startRewardsWithdrawalCron(db *sql.DB, configFileName, cronSchedule string) {
	// Create a new cron job
	c := cron.New()

	// Add a scheduled task to run the rewards withdrawal
	c.AddFunc(cronSchedule, func() {
		networkConfigs, err := ReadConfig(configFileName)
		if err != nil {
			fmt.Println("Error reading config:", err)
			return
		}

		for _, network := range networkConfigs {
			fmt.Printf("Processing network: %s\n", network.ChainID)

			// Query the initial balance
			initialBalance, err := queryBalance(network.Binary, network.Granter, network.Node)
			if err != nil {
				fmt.Printf("Error querying initial balance: %v\n", err)
				continue
			}
			fmt.Printf("Initial Balance: %s\n", initialBalance)

			// Withdraw rewards and execute on the network
			withdrawRewards(network.Binary, network.Granter, network.ChainID, network.Node, network.FeePayer, network.Grantee)

			// Query the new balance
			newBalance, err := queryBalance(network.Binary, network.Granter, network.Node)
			if err != nil {
				fmt.Printf("Error querying new balance: %v\n", err)
				continue
			}
			fmt.Printf("New Balance: %s\n", newBalance)

			// Calculate income (new balance - initial balance)
			income := calculateIncome(newBalance, initialBalance)
			fmt.Printf("Income: %s\n", income)

			// Insert income details into 'income' table
			insertIncomeDetails(db, network.ChainID, network.Granter, initialBalance, income, newBalance)
		}
	})

	// Start the cron job
	c.Start()

	// Run the cron job indefinitely
	select {}
}
