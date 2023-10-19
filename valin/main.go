package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/robfig/cron" // Import the cron library

	lens "github.com/strangelove-ventures/lens/client"
	registry "github.com/strangelove-ventures/lens/client/chain_registry"
)

var config Configuration

func main() {
	configFileName := "config.json"

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

	// initChains
	initChains()

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
			fmt.Printf("Processing network: %s\n", network.ChainName)

			// Query the initial balance
			initialBalance, err := queryBalance(network)
			if err != nil {
				fmt.Printf("Error querying initial balance: %v\n", err)
				continue
			}
			fmt.Printf("Initial Balance: %v\n", initialBalance)

			// Withdraw rewards and execute on the network
			withdrawRewards(network.Binary, network.Granter, network.ChainID, network.Node, network.FeePayer, network.Grantee)

			// Query the new balance
			newBalance, err := queryBalance(network)
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

func initChains() {
	// We should use this chain-specific client while executing transactions and queries
	networkConfigs := config.Networks
	var granteeKeyMnemonic string

	for i, network := range networkConfigs {
		//	Fetches chain info from chain registry
		chainInfo, err := registry.DefaultChainRegistry().GetChain(network.ChainName)
		if err != nil {
			log.Fatalf("Failed to get chain info. Err: %v \n", err)
		}

		//	Use Chain info to select random endpoint
		rpc, err := chainInfo.GetRandomRPCEndpoint()
		if err != nil {
			log.Fatalf("Failed to get random RPC endpoint on chain %s. Err: %v \n", chainInfo.ChainName, err)
		}

		// For this example, lets place the key directory in your PWD.
		pwd, _ := os.Getwd()
		key_dir := pwd + "/keys"

		// Define a regular expression pattern to match non-alphanumeric characters and whitespace
		regex := regexp.MustCompile("[^a-zA-Z0-9]+")
		// Replace all matches with an empty string
		keyname := regex.ReplaceAllString(network.ChainName, "")

		// Build chain config
		chainConfig_1 := lens.ChainClientConfig{
			Key:     keyname,
			ChainID: chainInfo.ChainID,
			RPCAddr: rpc,
			// GRPCAddr       string,
			AccountPrefix:  chainInfo.Bech32Prefix,
			KeyringBackend: "test",
			GasAdjustment:  1.2,
			GasPrices:      "0.01uosmo",
			KeyDirectory:   key_dir,
			Debug:          true,
			Timeout:        "20s",
			OutputFormat:   "json",
			SignModeStr:    "direct",
			Modules:        lens.ModuleBasics,
		}

		// Creates client object to pull chain info
		chainClient, err := lens.NewChainClient(&chainConfig_1, key_dir, os.Stdin, os.Stdout)
		if err != nil {
			log.Fatalf("Failed to build new chain client for %s. Err: %v \n", chainInfo.ChainID, err)
		}

		// TODO: check if key exists, if doesn't, restore
		// Lets restore a key with funds and name it "source_key", this is the wallet we'll use to send tx.
		srcWalletAddress, err := chainClient.RestoreKey(chainInfo.ChainID, granteeKeyMnemonic)
		if err != nil {
			log.Fatalf("Failed to restore key. Err: %v \n", err)
		}

		ChainClients[chainInfo.chain_id] = chainClient
	}
}
