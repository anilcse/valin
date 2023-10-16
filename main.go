package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/robfig/cron" // Import the cron library
)

// NetworkConfig represents the configuration for each network
type NetworkConfig struct {
	Binary   string `json:"binary"`
	Granter  string `json:"granter"`
	Grantee  string `json:"grantee"`
	ChainID  string `json:"chainId"`
	Node     string `json:"node"`
	FeePayer string `json:"feepayer"`
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
	DB          struct {
		Driver string `json:"driver"`
		Table  string `json:"table"`
		SQLURL string `json:"sqlUrl"`
	} `json:"db"`
	Networks []NetworkConfig `json:"networks"`
}

func main() {
	configFileName := "config.json"

	var config Configuration
	if err := readJSONConfig(configFileName, &config); err != nil {
		fmt.Println("Error reading config:", err)
		return
	}

	// Extract database details from the config structure
	dbURL := config.DB.SQLURL
	dbDriver := config.DB.Driver
	dbTable := config.DB.Table

	// Connect to the database using the extracted details
	db := connectToDatabase(dbURL, dbDriver)

	// Create the 'income' table if it doesn't exist
	if err := createTableIfNotExists(db, dbTable); err != nil {
		fmt.Println("Error creating income table:", err)
		return
	}

	// Schedule rewards withdrawal based on the cronJobTime
	startRewardsWithdrawalCron(db, dbTable, config, config.CronJobTime)

	// Run the HTTP server or other main program logic
	startHTTPServer(db, config)
}

func readJSONConfig(filename string, config interface{}) error {
	configFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer configFile.Close()

	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(config); err != nil {
		return err
	}

	return nil
}

func startRewardsWithdrawalCron(db *sql.DB, dbTable string, config Configuration, cronSchedule string) {
	// Create a new cron job
	c := cron.New()

	// Add a scheduled task to run the rewards withdrawal
	c.AddFunc(cronSchedule, func() {
		networkConfigs := config.Networks

		for _, network := range networkConfigs {
			fmt.Printf("Processing network: %s\n", network.ChainID)

			// Query the initial balance
			initialBalance, err := queryBalance(network.Binary, network.Granter, network.Node)
			if err != nil {
				fmt.Printf("Error querying initial balance: %v\n", err)
				continue
			}
			fmt.Printf("Initial Balance: %v\n", initialBalance)

			// Withdraw rewards and execute on the network
			withdrawRewards(network.Binary, network.Granter, network.ChainID, network.Node, network.FeePayer, network.Grantee)

			// Query the new balance
			newBalance, err := queryBalance(network.Binary, network.Granter, network.Node)
			if err != nil {
				fmt.Printf("Error querying new balance: %v\n", err)
				continue
			}
			fmt.Printf("New Balance: %v\n", newBalance)

			// Calculate income (new balance - initial balance)
			income := calculateIncome(newBalance, initialBalance)
			fmt.Printf("Income: %v\n", income)

			// Insert income details into 'income' table
			insertIncomeDetails(db, dbTable, network.ChainID, network.Granter, initialBalance, income, newBalance)
		}
	})

	// Start the cron job
	c.Start()

	// Run the cron job indefinitely
	select {}
}
