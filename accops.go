package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Balance struct {
	Denom  string `json:"denom"`
	Amount uint64 `json:"amount"`
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

// Calculate income based on slices of new and old balances
func calculateIncome(newBalances, oldBalances []Balance) []Balance {
	var incomes []Balance

	// Ensure that both slices have the same length
	if len(newBalances) != len(oldBalances) {
		fmt.Println("Error: Mismatched lengths of newBalances and oldBalances slices")
		return incomes
	}

	// Calculate income for each pair of balances
	for i := 0; i < len(newBalances); i++ {
		// Calculate income as the difference between new balance and old balance
		income := Balance{
			Denom:  newBalances[i].Denom,
			Amount: newBalances[i].Amount - oldBalances[i].Amount,
		}
		incomes = append(incomes, income)
	}

	return incomes
}

func parseBalanceToFloat(balance string) float64 {
	balanceFloat, _ := strconv.ParseFloat(balance, 64)
	return balanceFloat
}

// Withdraw rewards and execute the transaction
func withdrawRewards(binary, granter, chainID, node, feePayer, grantee string) {
	cmd := exec.Command(binary, "tx", "distribution", "withdraw-rewards", granter, "--from", granter, "--chain-id", chainID, "-y", "--generate-only")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error generating withdraw-rewards transaction: %v\n", err)
		return
	}

	// Execute the transaction using authz module
	cmd = exec.Command(binary, "tx", "authz", "exec", "-", "--chain-id", chainID, "--node", node, "--fees", "200uatom", "--fee-account", feePayer, "--from", grantee, "-y")
	cmd.Stdin = strings.NewReader(string(output))
	_, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error executing transaction: %v\n", err)
	}
}
