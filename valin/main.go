package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"regexp"

	"github.com/robfig/cron" // Import the cron library
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	lens "github.com/strangelove-ventures/lens/client"
	registry "github.com/strangelove-ventures/lens/client/chain_registry"
)

var config Configuration
var zapLog *zap.Logger

var ChainClients map[string]*lens.ChainClient

func main() {
	configFileName := "config.json"

	var err error
	zapConfig := zap.NewProductionConfig()
	enccoderConfig := zap.NewProductionEncoderConfig()
	zapcore.TimeEncoderOfLayout("Jan _2 15:04:05.000000000")
	enccoderConfig.StacktraceKey = "" // to hide stacktrace info
	zapConfig.EncoderConfig = enccoderConfig

	zapLog, err = zapConfig.Build(zap.AddCallerSkip(1))

	if err != nil {
		panic(err)
	}

	config, err = ReadConfig(configFileName)
	if err != nil {
		zapLog.Error(err.Error())
		return
	}

	ChainClients = make(map[string]*lens.ChainClient, len(config.Networks))

	// Extract database details from the config structure
	// dbURL := config.DB.SQLURL
	// dbDriver := config.DB.Driver
	// dbTable := config.DB.Table

	// // Connect to the database using the extracted details
	// db := connectToDatabase(dbURL, dbDriver)

	// // Create the 'income' table if it doesn't exist
	// if err := createTableIfNotExists(db, dbTable); err != nil {
	// 	fmt.Println("Error creating income table:", err)
	// 	return
	// }

	// initChains
	initChains()

	WithdrawRewardsAndStore()

	// TODO
	// // Schedule rewards withdrawal based on the cronJobTime
	// startRewardsWithdrawalCron(db, dbTable, config, config.CronJobTime)

	// // Run the HTTP server or other main program logic
	// startHTTPServer(db, config)
}

func startRewardsWithdrawalCron(db *sql.DB, dbTable string, config Configuration, cronSchedule string) {
	// Create a new cron job
	c := cron.New()

	// Add a scheduled task to run the rewards withdrawal
	c.AddFunc(cronSchedule, func() {
		WithdrawRewardsAndStore()
	})

	// Start the cron job
	c.Start()

	// Run the cron job indefinitely
	select {}
}

func initChains() {
	// We should use this chain-specific client while executing transactions and queries
	networkConfigs := config.Networks

	for _, network := range networkConfigs {
		//	Fetches chain info from chain registry
		chainInfo, err := registry.DefaultChainRegistry(zapLog).GetChain(context.Background(), network.ChainName)
		if err != nil {
			log.Fatalf("Failed to get chain info. Err: %v \n", err)
		}

		//	Use Chain info to select random endpoint
		rpc, err := chainInfo.GetRandomRPCEndpoint(context.Background())
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
		chainClient, err := lens.NewChainClient(zapLog, &chainConfig_1, key_dir, os.Stdin, os.Stdout)
		if err != nil {
			log.Fatalf("Failed to build new chain client for %s. Err: %v \n", chainInfo.ChainID, err)
		}

		// TODO: check if key exists, if doesn't, restore
		// Lets restore a key with funds and name it "source_key", this is the wallet we'll use to send tx.

		if _, err := chainClient.GetKeyAddress(); err != nil {
			_, err = chainClient.RestoreKey(keyname, config.Mnemonic, uint32(chainInfo.Slip44))
			if err != nil {
				log.Fatalf("\n \n Failed to restore key. Err: %v %v \n", err, config)
			}
		}

		ChainClients[chainInfo.ChainID] = chainClient
	}
}
