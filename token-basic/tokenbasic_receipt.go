package tokenbasic

import (
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/types"
)

func (t *TokenBasic) emitLock(to, ibcAccountAddr types.Address, value bn.Number) {
	toChainID := t.sdk.Helper().BlockChainHelper().GetChainID(to)
	peerChainID := ""
	if !t.sdk.Helper().BlockChainHelper().IsSideChain() {
		peerChainID = toChainID
	} else {
		peerChainID = t.sdk.Helper().BlockChainHelper().GetMainChainID()
	}
	peerChainBalance := t._peerChainBalance(t.sdk.Message().Contract().Token(), peerChainID)

	receipt := AssetChange{
		Version: t.sdk.Message().Contract().Version(),
		Type:    Lock,
		Token:   t.sdk.Message().Contract().Token(),
		From:    t.sdk.Message().Sender().Address(),
		To:      to,
		Value:   value,
		IBCHash: t.sdk.Helper().IBCHelper().IbcHash(toChainID),
		ChangeItems: []ChangeItem{
			{
				ChainID:          t.sdk.Block().ChainID(),
				Address:          ibcAccountAddr,
				PeerChainID:      peerChainID,
				PeerChainBalance: peerChainBalance,
			},
		},
	}
	t.sdk.Helper().ReceiptHelper().Emit(receipt)
}

func (t *TokenBasic) emitRecast(lockReceipt *AssetChange, ibcAccountAddr types.Address, peerChainID string, peerChainBalance bn.Number) {
	receipt := AssetChange{
		Version: t.sdk.Message().Contract().Version(),
		Type:    Recast,
		Token:   t.sdk.Message().Contract().Token(),
		From:    lockReceipt.From,
		To:      lockReceipt.To,
		Value:   lockReceipt.Value,
		IBCHash: lockReceipt.IBCHash,
		ChangeItems: []ChangeItem{
			{
				ChainID:          t.sdk.Block().ChainID(),
				Address:          ibcAccountAddr,
				PeerChainID:      peerChainID,
				PeerChainBalance: peerChainBalance,
			},
		},
	}
	t.sdk.Helper().ReceiptHelper().Emit(receipt)
}

func (t *TokenBasic) emitConfirm(recastReceipt *AssetChange, ibcAccountAddr types.Address, peerChainID string, peerChainBalance bn.Number) {
	receipt := AssetChange{
		Version: t.sdk.Message().Contract().Version(),
		Type:    Destroy,
		Token:   t.sdk.Message().Contract().Token(),
		From:    recastReceipt.From,
		To:      recastReceipt.To,
		Value:   recastReceipt.Value,
		IBCHash: recastReceipt.IBCHash,
		ChangeItems: []ChangeItem{
			{
				ChainID:          t.sdk.Block().ChainID(),
				Address:          ibcAccountAddr,
				PeerChainID:      peerChainID,
				PeerChainBalance: peerChainBalance,
			},
		},
	}
	t.sdk.Helper().ReceiptHelper().Emit(receipt)
}

func (t *TokenBasic) emitUnlock(lockReceipt *AssetChange, ibcAccountAddr types.Address, peerChainID string, peerChainBalance bn.Number) {
	receipt := AssetChange{
		Version: t.sdk.Message().Contract().Version(),
		Type:    Unlock,
		Token:   t.sdk.Message().Contract().Token(),
		From:    lockReceipt.From,
		To:      lockReceipt.To,
		Value:   lockReceipt.Value,
		IBCHash: lockReceipt.IBCHash,
		ChangeItems: []ChangeItem{
			{
				ChainID:          t.sdk.Block().ChainID(),
				Address:          ibcAccountAddr,
				PeerChainID:      peerChainID,
				PeerChainBalance: peerChainBalance,
			},
		},
	}
	t.sdk.Helper().ReceiptHelper().Emit(receipt)
}

func (t *TokenBasic) emitTransferTypeReceipt(receipt *AssetChange, ibcAccountAddr types.Address, fromChainID, toChainID string,
	fromPeerChainBal, toPeerChainBal bn.Number) {

	newReceipt := AssetChange{
		Version: t.sdk.Message().Contract().Version(),
		Type:    Transfer,
		Token:   t.sdk.Message().Contract().Token(),
		From:    receipt.From,
		To:      receipt.To,
		Value:   receipt.Value,
		IBCHash: receipt.IBCHash,
		ChangeItems: []ChangeItem{
			{
				ChainID:          t.sdk.Block().ChainID(),
				Address:          ibcAccountAddr,
				PeerChainID:      fromChainID,
				PeerChainBalance: fromPeerChainBal,
			},
			{
				ChainID:          t.sdk.Block().ChainID(),
				Address:          ibcAccountAddr,
				PeerChainID:      toChainID,
				PeerChainBalance: toPeerChainBal,
			},
		},
	}
	t.sdk.Helper().ReceiptHelper().Emit(newReceipt)
}
