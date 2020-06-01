package tokenbasic

import (
	"fmt"
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/jsoniter"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
	"strings"
)

//@:public:ibc:gas[0]
func (t *TokenBasic) Recast(ibcHash types.Hash) bool {
	if !t.isCallByIBC() {
		return false
	}

	lockReceipt, transferReceipt, ok := t.checkInputReceiptForRecast()
	if !ok {
		return false
	}

	peerChainID := ""
	if !t.sdk.Helper().BlockChainHelper().IsSideChain() {
		peerChainID = t.sdk.Helper().BlockChainHelper().GetChainID(lockReceipt.From)
	} else {
		sdk.RequireNewFormatAddress(lockReceipt.To)
		peerChainID = t.sdk.Helper().BlockChainHelper().GetMainChainID()
	}

	peerChainBalance := t._peerChainBalance(t.sdk.Message().Contract().Token(), peerChainID)
	if !t.sdk.Helper().BlockChainHelper().IsSideChain() &&
		lockReceipt.Value.IsGreaterThan(peerChainBalance) {
		return false
	}

	// recast
	ibcAccount := t.sdk.Helper().ContractHelper().ContractOfName("ibc").Account()
	err := ibcAccount.TransferWithNoteEx(lockReceipt.To, lockReceipt.Value, strings.TrimPrefix(transferReceipt.Note, "lock:"))
	if err.ErrorCode != types.CodeOK {
		return false
	}

	peerChainBalance = peerChainBalance.Sub(lockReceipt.Value)
	t._setPeerChainBalance(t.sdk.Message().Contract().Token(), peerChainID, peerChainBalance)

	t.emitRecast(lockReceipt, t.sdk.Helper().ContractHelper().ContractOfName("ibc").Account().Address(), peerChainID, peerChainBalance)
	return true
}

//@:public:ibc:gas[0]
func (t *TokenBasic) Confirm(ibcHash types.Hash) {
	sdk.Require(t.isCallByIBC(),
		types.ErrNoAuthorization, "")

	recastReceipt := t.checkInputReceiptForConfirm()

	peerChainID := ""
	if !t.sdk.Helper().BlockChainHelper().IsSideChain() {
		peerChainID = t.sdk.Helper().BlockChainHelper().GetChainID(recastReceipt.To)
	} else {
		peerChainID = t.sdk.Helper().BlockChainHelper().GetMainChainID()
	}

	peerChainBalance := t._peerChainBalance(t.sdk.Message().Contract().Token(), peerChainID)
	peerChainBalance = peerChainBalance.Add(recastReceipt.Value)
	t._setPeerChainBalance(t.sdk.Message().Contract().Token(), peerChainID, peerChainBalance)

	t.emitConfirm(recastReceipt, t.sdk.Helper().ContractHelper().ContractOfName("ibc").Account().Address(), peerChainID, peerChainBalance)
}

//@:public:ibc:gas[0]
func (t *TokenBasic) Cancel(ibcHash types.Hash) {
	sdk.Require(t.isCallByIBC(),
		types.ErrNoAuthorization, "")

	lockReceipt, transferReceipt := t.checkInputReceiptForCancel()
	t.unlock(lockReceipt, transferReceipt)
}

//@:public:ibc:gas[0]
func (t *TokenBasic) TryRecast(ibcHash types.Hash) bool {
	if !t.isCallByIBC() {
		return false
	}

	lockReceipt, ok := t.checkInputReceiptForTryRecast()
	if !ok {
		return false
	}

	peerChainBal := t._peerChainBalance(t.sdk.Message().Contract().Token(), t.sdk.Helper().BlockChainHelper().GetChainID(lockReceipt.From))
	if lockReceipt.Value.IsGreaterThan(peerChainBal) {
		return false
	}
	return true
}

//@:public:ibc:gas[0]
func (t *TokenBasic) ConfirmRecast(ibcHash types.Hash) {
	sdk.Require(t.isCallByIBC(), types.ErrNoAuthorization, "")

	recastReceipt := t.checkInputReceiptForConfirmRecast()

	// update from chain balance
	fromChainID := t.sdk.Helper().BlockChainHelper().GetChainID(recastReceipt.From)
	fromPeerChainBal := t._peerChainBalance(t.sdk.Message().Contract().Token(), fromChainID).Sub(recastReceipt.Value)
	t._setPeerChainBalance(t.sdk.Message().Contract().Token(), fromChainID, fromPeerChainBal)

	// update to chain balance
	toChainID := t.sdk.Helper().BlockChainHelper().GetChainID(recastReceipt.To)
	toPeerChainBal := t._peerChainBalance(t.sdk.Message().Contract().Token(), toChainID).Add(recastReceipt.Value)
	t._setPeerChainBalance(t.sdk.Message().Contract().Token(), toChainID, toPeerChainBal)

	// emit receipt
	t.emitTransferTypeReceipt(recastReceipt, t.sdk.Helper().ContractHelper().ContractOfName("ibc").Account().Address(),
		fromChainID, toChainID, fromPeerChainBal, toPeerChainBal)
}

//@:public:ibc:gas[0]
func (t *TokenBasic) CancelRecast(ibcHash types.Hash) {
	sdk.Require(t.isCallByIBC(),
		types.ErrNoAuthorization, "")

	lockReceipt := t.checkInputReceiptForCancelRecast()

	fromChainID := t.sdk.Helper().BlockChainHelper().GetChainID(lockReceipt.From)
	fromPeerChainBal := t._peerChainBalance(t.sdk.Message().Contract().Token(), fromChainID)

	// update to chain balance
	toChainID := t.sdk.Helper().BlockChainHelper().GetChainID(lockReceipt.To)
	toPeerChainBal := t._peerChainBalance(t.sdk.Message().Contract().Token(), toChainID)

	// emit receipt
	t.emitTransferTypeReceipt(lockReceipt, t.sdk.Helper().ContractHelper().ContractOfName("ibc").Account().Address(),
		fromChainID, toChainID, fromPeerChainBal, toPeerChainBal)
}

//@:public:ibc:gas[0]
func (t *TokenBasic) Notify(ibcHash types.Hash) {
	sdk.Require(t.isCallByIBC(),
		types.ErrNoAuthorization, "")

	inputReceiptsLen := len(t.sdk.Message().InputReceipts())
	sdk.Require(inputReceiptsLen == 1,
		types.ErrInvalidParameter, fmt.Sprintf("expected receipt's count: 1, obtain: %d", inputReceiptsLen))

	inputReceipt := t.sdk.Message().InputReceipts()[0]
	splitKey := strings.Split(string(inputReceipt.Key), "/")
	sdk.Require(splitKey[2] == "std::setGasPrice",
		types.ErrInvalidParameter, "")

	var receipt std.Receipt
	err := jsoniter.Unmarshal(inputReceipt.Value, &receipt)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	var sgp std.SetGasPrice
	err = jsoniter.Unmarshal(receipt.Bytes, &sgp)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	chainID := t.sdk.Block().ChainID()
	chainName := chainID[strings.Index(chainID, "[")+1 : len(chainID)-1]
	tokenAddr := t.sdk.Helper().BlockChainHelper().RecalcAddressEx(sgp.Token, chainName)
	token := t.sdk.Helper().TokenHelper().TokenOfAddress(tokenAddr)
	key := std.KeyOfToken(token.Address())
	stdToken := std.Token{
		Address:          token.Address(),
		Owner:            token.Owner().Address(),
		Name:             token.Name(),
		Symbol:           token.Symbol(),
		TotalSupply:      token.TotalSupply(),
		AddSupplyEnabled: token.AddSupplyEnabled(),
		BurnEnabled:      token.BurnEnabled(),
		GasPrice:         sgp.GasPrice,
	}
	t.sdk.Helper().StateHelper().McSet(key, &stdToken)

	// fire event of setGasPrice
	t.sdk.Helper().ReceiptHelper().Emit(std.SetGasPrice{
		Token:    tokenAddr,
		GasPrice: sgp.GasPrice,
	})
}
