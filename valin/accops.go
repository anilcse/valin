package main

import (
	"context"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func queryBalance(network NetworkConfig) (sdk.Coins, error) {
	chainClient := ChainClients[network.ChainID]
	balance, err := chainClient.QueryBalanceWithAddress(context.Background(), network.Granter)

	return balance, err
}

// Calculate income based on slices of new and old balances
func calculateIncome(newBalances, oldBalances sdk.Coins) sdk.Coins {
	var incomes sdk.Coins

	// // Ensure that both slices have the same length
	// if len(newBalances) != len(oldBalances) {
	// 	fmt.Println("Error: Mismatched lengths of newBalances and oldBalances slices")
	// 	return incomes
	// }

	// Calculate income for each pair of balances
	for i := 0; i < len(newBalances); i++ {
		// Calculate income as the difference between new balance and old balance
		income := sdk.Coin{
			Denom:  newBalances[i].Denom,
			Amount: newBalances[i].Amount.Sub(oldBalances[i].Amount),
		}
		incomes = append(incomes, income)
	}

	return incomes
}

func parseBalanceToFloat(balance string) float64 {
	balanceFloat, _ := strconv.ParseFloat(balance, 64)
	return balanceFloat
}

// Withdraw rewards and commission
// via authz
func withdrawRewardsAndCommission(network NetworkConfig) (*sdk.TxResponse, error) {
	chainClient := ChainClients[network.ChainID]

	//	Build transaction message
	delAddr, err := sdk.AccAddressFromBech32(network.Granter)
	if err != nil {
		return nil, err
	}
	valAddr, err := sdk.ValAddressFromBech32(network.Validator)
	if err != nil {
		return nil, err
	}
	grantee, err := sdk.AccAddressFromBech32(network.Grantee)
	if err != nil {
		return nil, err
	}

	// TODO
	// feepayer, err := sdk.AccAddressFromBech32(network.FeePayer)
	// if err != nil {
	// 	return err
	// }

	withdrwaDelegationMsg := distrtypes.NewMsgWithdrawDelegatorReward(delAddr, valAddr)
	withdrwaCommissionMsg := distrtypes.NewMsgWithdrawValidatorCommission(valAddr)
	msgs := []sdk.Msg{
		withdrwaDelegationMsg,
		withdrwaCommissionMsg,
	}

	authzMsg := authztypes.NewMsgExec(grantee, msgs)
	if err := authzMsg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Send message and get response
	return chainClient.SendMsgs(context.Background(), msgs)
}
