package tokenbasic

import (
	"fmt"
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/forx"
	"github.com/bcbchain/sdk/sdk/jsoniter"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
	"strings"
)

func (t *TokenBasic) lock(to types.Address, value bn.Number) {
	sdk.Require(value.IsPositive(),
		types.ErrInvalidParameter, "")

	ibcAccountAddr := t.sdk.Helper().ContractHelper().ContractOfName("ibc").Account().Address()

	// 转账给当 IBC 合约账户，临时锁定资金
	t.sdk.Message().Sender().TransferWithNote(ibcAccountAddr, value, "lock:"+t.sdk.Tx().Note())

	t.emitLock(to, ibcAccountAddr, value)
}

func (t *TokenBasic) unlock(lockReceipt *AssetChange, transferReceipt *std.Transfer) {

	ibcAccount := t.sdk.Helper().ContractHelper().ContractOfName("ibc").Account()
	ibcAccount.TransferWithNote(lockReceipt.From, lockReceipt.Value, "unlock:"+transferReceipt.Note)

	peerChainID := ""
	if !t.sdk.Helper().BlockChainHelper().IsSideChain() {
		peerChainID = t.sdk.Helper().BlockChainHelper().GetChainID(lockReceipt.To)
	} else {
		peerChainID = t.sdk.Helper().BlockChainHelper().GetMainChainID()
	}
	peerChainBal := t._peerChainBalance(t.sdk.Message().Contract().Token(), peerChainID)
	t.emitUnlock(lockReceipt, ibcAccount.Address(), peerChainID, peerChainBal)
}

func (t *TokenBasic) checkReceipts(lockReceipt, recastReceipt *AssetChange, lockTransferReceipt, recastTransferReceipt *std.Transfer) bool {
	token := t.sdk.Message().Contract().Token()
	ibcAccount := t.sdk.Helper().ContractHelper().ContractOfName("ibc").Account()

	tokenAddrWithoutChainID := token
	if strings.Contains(token, "0") {
		tokenAddrWithoutChainID = token[:strings.Index(token, "0")]
	}
	tokenAddrWithoutChainID = strings.Replace(tokenAddrWithoutChainID, t.sdk.Helper().BlockChainHelper().GetMainChainID(), "", 1)

	ibcAccountAddr := ibcAccount.Address()
	IBCAccountWithoutChainID := ibcAccountAddr
	if strings.Contains(ibcAccountAddr, "0") {
		IBCAccountWithoutChainID = ibcAccount.Address()[:strings.Index(ibcAccount.Address(), "0")]
	}
	IBCAccountWithoutChainID = strings.Replace(IBCAccountWithoutChainID, t.sdk.Helper().BlockChainHelper().GetMainChainID(), "", 1)

	if lockReceipt == nil || lockTransferReceipt == nil {
		return false
	}

	lockOk := false
	// check lock receipt & lock transfer to ibc account receipt
	if lockReceipt.From == lockTransferReceipt.From &&
		lockReceipt.Value.IsEqual(lockTransferReceipt.Value) &&
		lockReceipt.Token == lockTransferReceipt.Token &&
		strings.Contains(lockTransferReceipt.To, IBCAccountWithoutChainID) &&
		strings.Contains(lockTransferReceipt.Token, tokenAddrWithoutChainID) {
		lockOk = true
	}

	// check recast & lock receipt
	if recastTransferReceipt != nil && recastReceipt != nil {
		if recastReceipt.From != lockReceipt.From ||
			recastReceipt.To != lockReceipt.To ||
			!recastReceipt.Value.IsEqual(lockReceipt.Value) {
			return false
		}

		// fix bug #https://dc.giblockchain.cn/zentao/bug-view-1646.html
		// transfer to contract when recast, recastTransferReceipt.To is contract's account address.
		tChainID := t.sdk.Helper().BlockChainHelper().GetChainID(recastTransferReceipt.To)
		rChainID := t.sdk.Helper().BlockChainHelper().GetChainID(recastReceipt.To)
		return lockOk &&
			tChainID == rChainID &&
			recastReceipt.Token == recastTransferReceipt.Token &&
			recastReceipt.Value.IsEqual(recastTransferReceipt.Value) &&
			strings.Contains(recastTransferReceipt.From, IBCAccountWithoutChainID) &&
			strings.Contains(recastTransferReceipt.Token, tokenAddrWithoutChainID)
	}

	return lockOk
}

func (t *TokenBasic) checkInputReceiptForRecast() (lockReceipt *AssetChange, transferReceipt *std.Transfer, ok bool) {

	receipts := t.getReceipt("std::transfer")
	if len(receipts) < 1 {
		return
	}

	transferReceipt = new(std.Transfer)
	err := jsoniter.Unmarshal(receipts[0].Bytes, transferReceipt)
	if err != nil {
		return
	}

	lockReceipt = t.getAssetChangeReceipt(Lock)
	if t.checkReceipts(lockReceipt, nil, transferReceipt, nil) {
		ok = true
		return
	}

	return
}

func (t *TokenBasic) checkInputReceiptForConfirm() (recastReceipt *AssetChange) {
	recastReceipt = t.getAssetChangeReceipt(Recast)
	sdk.Require(recastReceipt != nil,
		types.ErrInvalidParameter, "lose recast asset change receipt")

	receipts := t.getReceipt("std::transfer")
	sdk.Require(len(receipts) == 2,
		types.ErrInvalidParameter, fmt.Sprintf("expected transfer receipt count: 2, obtain: %d", len(receipts)))

	transferReceipt0 := new(std.Transfer)
	err := jsoniter.Unmarshal(receipts[0].Bytes, transferReceipt0)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	transferReceipt1 := new(std.Transfer)
	err = jsoniter.Unmarshal(receipts[1].Bytes, transferReceipt1)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	lockTransferReceipt := new(std.Transfer)
	recastTransferReceipt := new(std.Transfer)

	lockReceipt := t.getAssetChangeReceipt(Lock)
	if transferReceipt0.Token == lockReceipt.Token {
		lockTransferReceipt = transferReceipt0
		recastTransferReceipt = transferReceipt1

	} else {
		lockTransferReceipt = transferReceipt1
		recastTransferReceipt = transferReceipt0
	}

	sdk.Require(t.checkReceipts(lockReceipt, recastReceipt, lockTransferReceipt, recastTransferReceipt),
		types.ErrInvalidParameter, "invalid receipts")

	return
}

func (t *TokenBasic) checkInputReceiptForCancel() (lockReceipt *AssetChange, transferReceipt *std.Transfer) {
	lockReceipt = t.getAssetChangeReceipt(Lock)

	receipts := t.getReceipt("std::transfer")
	sdk.Require(len(receipts) == 1,
		types.ErrInvalidParameter, fmt.Sprintf("expected transfer receipt count: 1, obtain: %d", len(receipts)))

	transferReceipt = new(std.Transfer)
	err := jsoniter.Unmarshal(receipts[0].Bytes, transferReceipt)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	sdk.Require(t.checkReceipts(lockReceipt, nil, transferReceipt, nil),
		types.ErrInvalidParameter, "invalid receipts")

	return
}

func (t *TokenBasic) checkInputReceiptForTryRecast() (lockReceipt *AssetChange, ok bool) {
	loclTransferReceipt := new(std.Transfer)
	receipts := t.getReceipt("std::transfer")
	if len(receipts) != 1 {
		return
	}

	err := jsoniter.Unmarshal(receipts[0].Bytes, loclTransferReceipt)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	lockReceipt = t.getAssetChangeReceipt(Lock)
	if t.checkReceipts(lockReceipt, nil, loclTransferReceipt, nil) {
		ok = true
		return
	}

	return
}

func (t *TokenBasic) checkInputReceiptForConfirmRecast() (recastReceipt *AssetChange) {
	recastReceipt = t.getAssetChangeReceipt(Recast)
	sdk.Require(recastReceipt != nil,
		types.ErrInvalidParameter, "lose recast asset change receipt")

	lockReceipt := t.getAssetChangeReceipt(Lock)
	sdk.Require(lockReceipt != nil,
		types.ErrInvalidParameter, "lose lock asset change receipt")

	receipts := t.getReceipt("std::transfer")
	sdk.Require(len(receipts) == 2,
		types.ErrInvalidParameter, fmt.Sprintf("expected transfer receipt count: 2, obtain: %d", len(receipts)))

	// 解析两个给 IBC 合约账户转账的收据
	transferReceipt0 := new(std.Transfer)
	err := jsoniter.Unmarshal(receipts[0].Bytes, transferReceipt0)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	transferReceipt1 := new(std.Transfer)
	err = jsoniter.Unmarshal(receipts[1].Bytes, transferReceipt1)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	lockTransferReceipt := new(std.Transfer)
	recastTransferReceipt := new(std.Transfer)

	if transferReceipt0.Token == lockReceipt.Token {
		lockTransferReceipt = transferReceipt0
		recastTransferReceipt = transferReceipt1

	} else {
		lockTransferReceipt = transferReceipt1
		recastTransferReceipt = transferReceipt0
	}

	sdk.Require(t.checkReceipts(lockReceipt, recastReceipt, lockTransferReceipt, recastTransferReceipt),
		types.ErrInvalidParameter, "invalid receipts")

	return
}

func (t *TokenBasic) checkInputReceiptForCancelRecast() (lockReceipt *AssetChange) {
	lockTransferReceipt := new(std.Transfer)
	receipts := t.getReceipt("std::transfer")
	sdk.Require(len(receipts) == 1,
		types.ErrInvalidParameter, fmt.Sprintf("expected transfer receipt count: 1, obtain: %d", len(receipts)))

	err := jsoniter.Unmarshal(receipts[0].Bytes, lockTransferReceipt)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	lockReceipt = t.getAssetChangeReceipt(Lock)
	sdk.Require(t.checkReceipts(lockReceipt, nil, lockTransferReceipt, nil),
		types.ErrInvalidParameter, "invalid receipts")
	return
}

func (t *TokenBasic) getAssetChangeReceipt(receiptType string) *AssetChange {
	receipt := new(AssetChange)
	forx.Range(t.sdk.Message().InputReceipts(), func(i int, kvPair types.KVPair) bool {
		if strings.HasSuffix(string(kvPair.Key), "tokenbasic.AssetChange") {
			r := new(std.Receipt)
			if err := jsoniter.Unmarshal(kvPair.Value, r); err != nil {
				return forx.Continue
			}

			if r.Name == "tokenbasic.AssetChange" {
				temp := new(AssetChange)
				if err := jsoniter.Unmarshal(r.Bytes, temp); err != nil {
					return forx.Continue
				}
				if temp.Type == receiptType {
					receipt = temp
					return forx.Break
				}
			}
		}
		return true
	})
	return receipt
}

func (t *TokenBasic) getReceipt(name string) []*std.Receipt {
	var result []*std.Receipt
	forx.Range(t.sdk.Message().InputReceipts(), func(i int, kvPair types.KVPair) bool {
		if strings.HasSuffix(string(kvPair.Key), name) {
			r := new(std.Receipt)
			if err := jsoniter.Unmarshal(kvPair.Value, r); err != nil {
				return forx.Continue
			}

			if r.Name == name {
				result = append(result, r)
			}
		}
		return true
	})
	return result
}
