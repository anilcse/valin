package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

type Balance struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

func queryBalance(binary, account, node string) ([]Balance, error) {
	cmd := exec.Command(binary, "q", "bank", "balances", account, "--node", node)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var result map[string][]Balance
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, err
	}

	balances := result["balances"]
	return balances, nil
}

func calculateIncome(newBalance, oldBalance string) string {
	// Parse the balances as floats
	newBalanceFloat := parseBalanceToFloat(newBalance)
	oldBalanceFloat := parseBalanceToFloat(oldBalance)

	// Calculate income
	incomeFloat := newBalanceFloat - oldBalanceFloat
	return fmt.Sprintf("%.6f", incomeFloat)
}

func parseBalanceToFloat(balance string) float64 {
	balanceFloat, _ := strconv.ParseFloat(balance, 64)
	return balanceFloat
}
