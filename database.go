package main

import (
	"database/sql"
	"fmt"
	"os"
)

func connectToDatabase(connString string) *sql.DB {
	db, err := sql.Open("mysql", connString)
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		os.Exit(1)
	}
	return db
}

func insertIncomeDetails(db *sql.DB, chainID, granter, oldBalance, income, newBalance string) error {
	insertStatement := `
        INSERT INTO income (chain_id, granter, old_balance, income, new_balance, date)
        VALUES (?, ?, ?, ?, ?, NOW())`
	_, err := db.Exec(insertStatement, chainID, granter, oldBalance, income, newBalance)
	if err != nil {
		fmt.Printf("Error inserting income details into SQL: %v\n", err)
		return err
	}
	return nil
}

// Get income data from the SQL table ordered by date
func getIncomeData(db *sql.DB) ([]IncomeData, error) {
	query := "SELECT chain_id, granter, old_balance, income, new_balance, date FROM income ORDER BY date"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incomeData []IncomeData
	for rows.Next() {
		var data IncomeData
		if err := rows.Scan(&data.ChainID, &data.Granter, &data.OldBalance, &data.Income, &data.NewBalance, &data.Date); err != nil {
			return nil, err
		}
		incomeData = append(incomeData, data)
	}

	return incomeData, nil
}
