package main

import (
	"database/sql"
	"fmt"
	"os"
)

func connectToDatabase(dbURL, dbDriver string) *sql.DB {
	db, err := sql.Open(dbDriver, dbURL)
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		os.Exit(1)
	}
	return db
}

func createTableIfNotExists(db *sql.DB, tableName string) error {
	createTableSQL := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            id INT AUTO_INCREMENT PRIMARY KEY,
            chain_id VARCHAR(255),
            granter VARCHAR(255),
            old_balance VARCHAR(255),
            income VARCHAR(255),
            new_balance VARCHAR(255),
            date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`, tableName)

	_, err := db.Exec(createTableSQL)
	if err != nil {
		fmt.Printf("Error creating table: %v\n", err)
		return err
	}
	return nil
}

func insertIncomeDetails(db *sql.DB, dbTable, chainID, granter string, oldBalance, income, newBalance sdk.Coins) error {
	insertStatement := fmt.Sprintf(`
        INSERT INTO income (chain_id, granter, old_balance, income, new_balance, date)
        VALUES (?, ?, ?, ?, ?, NOW())`, dbTable)
	_, err := db.Exec(insertStatement, chainID, granter, oldBalance, income, newBalance)
	if err != nil {
		fmt.Printf("Error inserting income details into SQL: %v\n", err)
		return err
	}
	return nil
}

// Get income data from the SQL table ordered by date
func getIncomeData(db *sql.DB, dbTable string) ([]IncomeData, error) {
	query := fmt.Sprintf(`SELECT chain_id, granter, old_balance, income, new_balance, date FROM income ORDER BY date`, dbTable)
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
