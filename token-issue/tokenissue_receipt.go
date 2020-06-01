package tokenissue

import (
	"github.com/bcbchain/sdk/sdk/std"
)

func (ti *TokenIssue) emitNewToken(token *std.Token) {
	//Receipts -- token issue, transfer
	ti.sdk.Helper().ReceiptHelper().Emit(
		std.NewToken{
			TokenAddress:     token.Address,
			ContractAddress:  token.Address,
			Owner:            token.Owner,
			Name:             token.Name,
			Symbol:           token.Symbol,
			TotalSupply:      token.TotalSupply,
			AddSupplyEnabled: token.AddSupplyEnabled,
			BurnEnabled:      token.BurnEnabled,
			GasPrice:         token.GasPrice,
			OrgName:          token.OrgName,
			Proto:            token.Proto,
		},
	)
}

func (ti *TokenIssue) emitTransfer(token *std.Token) {
	ti.sdk.Helper().ReceiptHelper().Emit(
		std.Transfer{
			Token: token.Address,
			From:  "",
			To:    token.Owner,
			Value: token.TotalSupply,
		},
	)
}
