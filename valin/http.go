package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func startHTTPServer(db *sql.DB, config Configuration) {
	router := mux.NewRouter()
	router.HandleFunc("/income", func(w http.ResponseWriter, r *http.Request) {
		incomeData, err := getIncomeData(db, config.DB.Table)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(incomeData); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}).Methods("GET")

	http.Handle("/", router)
	http.ListenAndServe(":8080", nil)
}

func WithdrawRewardsAndStore() {
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

		// Withdraw rewards and commission
		// and execute on the network via authz
		res, err := withdrawRewardsAndCommission(network)
		if err != nil {
			fmt.Printf("Error Withdraw Rewards: %v\n", err)
			continue
		}

		fmt.Printf("Withdraw Rewards successful: %v\n", res)

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

		// TODO
		// Insert income details into 'income' table
		// insertIncomeDetails(db, dbTable, network.ChainID, network.Granter, initialBalance, income, newBalance)
	}
}
