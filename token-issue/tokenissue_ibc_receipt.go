package tokenissue

import (
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/types"
)

func (ti *TokenIssue) emitLock(to, ibcAccount types.Address, value bn.Number) {
	token := ti.sdk.Message().Contract().Token()
	peerChainBal := bn.N(0)
	peerChainID := ti.sdk.Helper().BlockChainHelper().GetChainID(to)

	if ti.sdk.Helper().BlockChainHelper().IsSideChain() {
		peerChainBal = ti._peerChainBalance(token, ti.sdk.Helper().BlockChainHelper().GetMainChainID())
	} else {
		peerChainBal = ti._peerChainBalance(token, peerChainID)
	}

	receipt := AssetChange{
		Version: ti.sdk.Message().Contract().Version(),
		Type:    Lock,
		Token:   token,
		From:    ti.sdk.Message().Sender().Address(),
		To:      to,
		Value:   value,
		IBCHash: ti.sdk.Helper().IBCHelper().IbcHash(peerChainID),
		ChangeItems: []ChangeItem{
			{
				ChainID:          ti.sdk.Block().ChainID(),
				Address:          ibcAccount,
				PeerChainID:      peerChainID,
				PeerChainBalance: peerChainBal,
			},
		},
	}
	ti.sdk.Helper().ReceiptHelper().Emit(receipt)
}

func (ti *TokenIssue) emitUnlock(lockReceipt *AssetChange, ibcAccount types.Address, peerChainID string, balance bn.Number) {
	receipt := AssetChange{
		Version: ti.sdk.Message().Contract().Version(),
		Type:    Unlock,
		Token:   ti.sdk.Message().Contract().Token(),
		From:    lockReceipt.From,
		To:      lockReceipt.To,
		Value:   lockReceipt.Value,
		IBCHash: lockReceipt.IBCHash,
		ChangeItems: []ChangeItem{
			{
				ChainID:          ti.sdk.Block().ChainID(),
				Address:          ibcAccount,
				PeerChainID:      peerChainID,
				PeerChainBalance: balance,
			},
		},
	}
	ti.sdk.Helper().ReceiptHelper().Emit(receipt)
}

func (ti *TokenIssue) emitRecast(lockReceipt *AssetChange, ibcAccount types.Address, peerChainID string, balance bn.Number) {
	receipt := AssetChange{
		Version: ti.sdk.Message().Contract().Version(),
		Type:    Recast,
		Token:   ti.sdk.Message().Contract().Token(),
		From:    lockReceipt.From,
		To:      lockReceipt.To,
		Value:   lockReceipt.Value,
		IBCHash: lockReceipt.IBCHash,
		ChangeItems: []ChangeItem{
			{
				ChainID:          ti.sdk.Block().ChainID(),
				Address:          ibcAccount,
				PeerChainID:      peerChainID,
				PeerChainBalance: balance,
			},
		},
	}
	ti.sdk.Helper().ReceiptHelper().Emit(receipt)
}

func (ti *TokenIssue) emitConfirm(recastReceipt *AssetChange, balance bn.Number, ibcAccount types.Address, peerChainID string) {
	receipt := AssetChange{
		Version: ti.sdk.Message().Contract().Version(),
		Type:    Destroy,
		Token:   ti.sdk.Message().Contract().Token(),
		From:    recastReceipt.From,
		To:      recastReceipt.To,
		Value:   recastReceipt.Value,
		IBCHash: recastReceipt.IBCHash,
		ChangeItems: []ChangeItem{
			{
				ChainID:          ti.sdk.Block().ChainID(),
				Address:          ibcAccount,
				PeerChainID:      peerChainID,
				PeerChainBalance: balance,
			},
		},
	}
	ti.sdk.Helper().ReceiptHelper().Emit(receipt)
}

func (ti *TokenIssue) emitTransferTypeReceipt(receipt *AssetChange, ibcAccount types.Address, fromChainID, toChainID string,
	fromPeerChainBal, toPeerChainBal bn.Number) {

	chainID := ti.sdk.Block().ChainID()
	newReceipt := AssetChange{
		Version: ti.sdk.Message().Contract().Version(),
		Type:    Transfer,
		Token:   ti.sdk.Message().Contract().Token(),
		From:    receipt.From,
		To:      receipt.To,
		Value:   receipt.Value,
		IBCHash: receipt.IBCHash,
		ChangeItems: []ChangeItem{
			{
				ChainID:          chainID,
				Address:          ibcAccount,
				PeerChainID:      fromChainID,
				PeerChainBalance: fromPeerChainBal,
			},
			{
				ChainID:          chainID,
				Address:          ibcAccount,
				PeerChainID:      toChainID,
				PeerChainBalance: toPeerChainBal,
			},
		},
	}
	ti.sdk.Helper().ReceiptHelper().Emit(newReceipt)
}

func (ti *TokenIssue) emitActivate(chainName string, token sdk.IToken, gasPrice int64, orgName string) {
	tokenAddr := token.Address()
	contract := ti.sdk.Helper().ContractHelper().ContractOfToken(tokenAddr)
	sideChainTokenAddr := ti.sdk.Helper().BlockChainHelper().RecalcAddressEx(tokenAddr, chainName)
	sideChainOwner := ti.sdk.Helper().BlockChainHelper().RecalcAddressEx(token.Owner().Address(), chainName)

	receipt := Activate{
		ChainName:    chainName,
		Address:      sideChainTokenAddr,
		Owner:        sideChainOwner,
		OrgID:        contract.OrgID(),
		ContractName: contract.Name(),
		Name:         token.Name(),
		Symbol:       token.Symbol(),
		GasPrice:     gasPrice,
		OrgName:      orgName,
	}
	ti.sdk.Helper().ReceiptHelper().Emit(receipt)
}
