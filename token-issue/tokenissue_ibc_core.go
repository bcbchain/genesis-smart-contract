package tokenissue

import (
	"fmt"
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/forx"
	"github.com/bcbchain/sdk/sdk/jsoniter"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
	"github.com/bcbchain/sdk/sdkimpl/object"
	"strings"
)

/////////////////////////// tcctx core function  //////////////////////////////
func (ti *TokenIssue) lock(to types.Address, value bn.Number) {
	ibcAccount := ti.sdk.Helper().ContractHelper().ContractOfName("ibc").Account().Address()

	// 转账给当 IBC 合约账户，临时锁定资金
	ti.sdk.Message().Sender().TransferWithNote(ibcAccount, value, "lock:"+ti.sdk.Tx().Note())

	ti.emitLock(to, ibcAccount, value)
}

func (ti *TokenIssue) lockWithNote(to types.Address, value bn.Number, note string) {
	ibcAccount := ti.sdk.Helper().ContractHelper().ContractOfName("ibc").Account().Address()

	ti.sdk.Message().Sender().TransferWithNote(ibcAccount, value, "lock:"+note)

	ti.emitLock(to, ibcAccount, value)
}

func (ti *TokenIssue) unlock(lockAssetChange *AssetChange, lockTransfer *std.Transfer) {
	ibcAccount := ti.sdk.Helper().ContractHelper().ContractOfName("ibc").Account()
	ibcAccount.TransferWithNote(lockAssetChange.From, lockAssetChange.Value, "unlock:"+lockTransfer.Note)
	peerChainID := ""
	if ti.sdk.Helper().BlockChainHelper().IsSideChain() {
		// 如果是侧链，修改主链的 peerChainBalance
		peerChainID = ti.sdk.Helper().BlockChainHelper().GetMainChainID()
	} else {
		// 如果是主链，修改源侧链的 peerChainBalance
		peerChainID = ti.sdk.Helper().BlockChainHelper().GetChainID(lockAssetChange.To)
	}
	peerChainBal := ti._peerChainBalance(ti.sdk.Message().Contract().Token(), peerChainID)
	ti.emitUnlock(lockAssetChange, ibcAccount.Address(), peerChainID, peerChainBal)
}

func (ti *TokenIssue) recast(lockAssetChange *AssetChange, lockTransfer *std.Transfer) bool {
	ibcAccount := ti.sdk.Helper().ContractHelper().ContractOfName("ibc").Account()

	peerChainBal := bn.N(0)
	peerChainID := ""
	if ti.sdk.Helper().BlockChainHelper().IsSideChain() {
		sdk.RequireNewFormatAddress(lockAssetChange.To)
		// 如果是侧链，修改主链的 peerChainBalance
		peerChainID = ti.sdk.Helper().BlockChainHelper().GetMainChainID()
	} else {
		// 如果是主链，修改源侧链的 peerChainBalance
		peerChainID = ti.sdk.Helper().BlockChainHelper().GetChainID(lockAssetChange.From)
	}

	b := ti._peerChainBalance(ti.sdk.Message().Contract().Token(), peerChainID)
	if !ti.sdk.Helper().BlockChainHelper().IsSideChain() &&
		lockTransfer.Value.IsGreaterThan(b) {
		return false
	}

	// transfer from ibcAccount to toAccount
	err := ibcAccount.TransferWithNoteEx(lockAssetChange.To, lockAssetChange.Value, strings.TrimPrefix(lockTransfer.Note, "lock:"))
	if err.ErrorCode != types.CodeOK {
		return false
	}

	// update peer chain balance
	peerChainBal = b.Sub(lockAssetChange.Value)
	ti._setPeerChainBalance(ti.sdk.Message().Contract().Token(), peerChainID, peerChainBal)

	ti.emitRecast(lockAssetChange, ibcAccount.Address(), peerChainID, peerChainBal)

	return true
}

func (ti *TokenIssue) confirm(recastAssetChange *AssetChange) {
	peerChainBal := bn.N(0)
	peerChainID := ""
	if ti.sdk.Helper().BlockChainHelper().IsSideChain() {
		// 如果是侧链，修改主链的 peerChainBalance
		peerChainID = ti.sdk.Helper().BlockChainHelper().GetMainChainID()
	} else {
		// 如果是主链，修改源侧链的 peerChainBalance
		peerChainID = ti.sdk.Helper().BlockChainHelper().GetChainID(recastAssetChange.To)
	}
	b := ti._peerChainBalance(ti.sdk.Message().Contract().Token(), peerChainID)
	peerChainBal = b.Add(recastAssetChange.Value)
	ti._setPeerChainBalance(ti.sdk.Message().Contract().Token(), peerChainID, peerChainBal)

	ti.emitConfirm(recastAssetChange, peerChainBal, ti.sdk.Helper().ContractHelper().ContractOfName("ibc").Account().Address(), peerChainID)
}

func (ti *TokenIssue) tryRecast(lockAssetChange *AssetChange) bool {
	fromChainID := ti.sdk.Helper().BlockChainHelper().GetChainID(lockAssetChange.From)
	peerChainBal := ti._peerChainBalance(ti.sdk.Message().Contract().Token(), fromChainID)

	//errMsg := fmt.Sprintf("expected fromChain balance great or equal than: %v obtain: %v", lockAssetChange.Value, peerChainBal)
	if !peerChainBal.IsGE(lockAssetChange.Value) {
		return false
	}

	return true
}

func (ti *TokenIssue) confirmRecast(recastAssetChange *AssetChange) {
	token := ti.sdk.Message().Contract().Token()
	ibcAccount := ti.sdk.Helper().ContractHelper().ContractOfName("ibc").Account()

	// update from chain balance
	fromChainID := ti.sdk.Helper().BlockChainHelper().GetChainID(recastAssetChange.From)
	fromPeerChainBal := ti._peerChainBalance(token, fromChainID).Sub(recastAssetChange.Value)
	ti._setPeerChainBalance(token, fromChainID, fromPeerChainBal)

	// update to chain balance
	toChainID := ti.sdk.Helper().BlockChainHelper().GetChainID(recastAssetChange.To)
	toPeerChainBal := ti._peerChainBalance(token, toChainID).Add(recastAssetChange.Value)
	ti._setPeerChainBalance(token, toChainID, toPeerChainBal)

	// emit receipt
	ti.emitTransferTypeReceipt(recastAssetChange, ibcAccount.Address(), fromChainID, toChainID,
		fromPeerChainBal, toPeerChainBal)
}

func (ti *TokenIssue) cancelRecast(lockAssetChange *AssetChange) {
	// update from chain balance
	token := ti.sdk.Message().Contract().Token()
	ibcAccount := ti.sdk.Helper().ContractHelper().ContractOfName("ibc").Account()
	fromChainID := ti.sdk.Helper().BlockChainHelper().GetChainID(lockAssetChange.From)
	fromPeerChainBal := ti._peerChainBalance(token, fromChainID)

	// update to chain balance
	toChainID := ti.sdk.Helper().BlockChainHelper().GetChainID(lockAssetChange.To)
	toPeerChainBal := ti._peerChainBalance(token, toChainID)

	// emit receipt
	ti.emitTransferTypeReceipt(lockAssetChange, ibcAccount.Address(), fromChainID, toChainID,
		fromPeerChainBal, toPeerChainBal)
}

///////////////////////////  notify core function  //////////////////////////////
func (ti *TokenIssue) setGasPriceForNotify(inputReceipt std.Receipt) {
	var sgp std.SetGasPrice
	err := jsoniter.Unmarshal(inputReceipt.Bytes, &sgp)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	tokenAddr := ti.sdk.Helper().BlockChainHelper().RecalcAddressEx(sgp.Token, ti.getChainName())
	//token := ti.sdk.Helper().TokenHelper().TokenOfAddress(tokenAddr)
	key := std.KeyOfToken(tokenAddr)
	token, _ := ti.sdk.Helper().StateHelper().Get(key, new(std.Token)).(*std.Token)

	stdToken := std.Token{
		Address:          token.Address,
		Owner:            token.Owner,
		Name:             token.Name,
		Symbol:           token.Symbol,
		TotalSupply:      token.TotalSupply,
		AddSupplyEnabled: token.AddSupplyEnabled,
		BurnEnabled:      token.BurnEnabled,
		GasPrice:         sgp.GasPrice,
		OrgName:          token.OrgName,
		Proto:            token.Proto,
	}
	ti.sdk.Helper().StateHelper().McSet(key, &stdToken)

	// fire event of setGasPrice
	ti.sdk.Helper().ReceiptHelper().Emit(std.SetGasPrice{
		Token:    tokenAddr,
		GasPrice: sgp.GasPrice,
	})
}

func (ti *TokenIssue) setOwnerForNotify(inputReceipt std.Receipt) {
	var so std.SetOwner
	err := jsoniter.Unmarshal(inputReceipt.Bytes, &so)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	contractAddr := ti.sdk.Helper().BlockChainHelper().RecalcAddressEx(so.ContractAddr, ti.getChainName())
	newOwnerAddr := ti.sdk.Helper().BlockChainHelper().RecalcAddressEx(so.NewOwner, ti.getChainName())
	token := ti.sdk.Helper().TokenHelper().Token()

	// add contract to new owner and delete contract from old owner
	ti.addContractToNewOwner(newOwnerAddr, contractAddr)
	ti.delContractFromOldOwner(token.Owner().Address(), contractAddr)
	if token != nil {
		token.(*object.Token).SetOwner(newOwnerAddr)
	}

	key := std.KeyOfContract(contractAddr)
	// set new contract
	contract := ti.sdk.Helper().ContractHelper().ContractOfToken(contractAddr)
	stdContract := std.Contract{
		Address:      contract.Address(),
		Account:      contract.Account().Address(),
		Owner:        newOwnerAddr,
		Name:         contract.Name(),
		Version:      contract.Version(),
		CodeHash:     contract.CodeHash(),
		EffectHeight: contract.EffectHeight(),
		LoseHeight:   contract.LoseHeight(),
		KeyPrefix:    contract.KeyPrefix(),
		Methods:      contract.Methods(),
		Interfaces:   contract.Interfaces(),
		Mine:         contract.Mine(),
		IBCs:         contract.IBCs(),
		Token:        contract.Token(),
		OrgID:        contract.OrgID(),
		ChainVersion: contract.ChainVersion(),
	}
	ti.sdk.Helper().StateHelper().McSet(key, &stdContract)

	// fire event
	ti.sdk.Helper().ReceiptHelper().Emit(std.SetOwner{
		ContractAddr: contractAddr,
		NewOwner:     newOwnerAddr,
	})
}

func (ti *TokenIssue) addContractToNewOwner(newOwner types.Address, contractAddress types.Address) {
	key := std.KeyOfAccountContracts(newOwner)
	addrList := ti.sdk.Helper().StateHelper().GetStrings(key)
	addrList = append(addrList, contractAddress)

	ti.sdk.Helper().StateHelper().McSet(key, &addrList)
}

func (ti *TokenIssue) delContractFromOldOwner(oldOwner types.Address, contractAddress types.Address) {
	key := std.KeyOfAccountContracts(oldOwner)
	addrList := ti.sdk.Helper().StateHelper().GetStrings(key)

	forx.Range(addrList, func(index int, addr types.Address) bool {
		if addr == contractAddress {
			addrList = append(addrList[:index], addrList[index+1:]...)
			return forx.Break
		}

		return true
	})

	ti.sdk.Helper().StateHelper().McSet(key, &addrList)
}

func (ti *TokenIssue) activateForNotify(inputReceipt std.Receipt) {

	var activate Activate
	err := jsoniter.Unmarshal(inputReceipt.Bytes, &activate)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	// 非目的链时需要计算本地的tokenAddress
	toChainID := ti.sdk.Helper().BlockChainHelper().CalcSideChainID(activate.ChainName)
	tokenAddr := ti.sdk.Helper().BlockChainHelper().RecalcAddressEx(activate.Address, ti.getChainName())
	ti.addSupportSideChain(tokenAddr, toChainID) // 此处tokenAddr不能使用activate中的Address

	if toChainID == ti.sdk.Block().ChainID() {
		// 已激活则不重复激活
		contract := ti.sdk.Helper().ContractHelper().ContractOfAddress(activate.Address)
		if contract != nil && contract.Token() != "" {
			return
		}

		if activate.OrgID == ti.sdk.Message().Contract().OrgID() &&
			strings.HasPrefix(activate.ContractName, "token-templet") {

			ti.newToken(
				activate.Address,
				activate.Owner,
				activate.Name,
				activate.Symbol,
				bn.N(0),
				false,
				false,
				activate.GasPrice,
				activate.OrgName,
			)
		} else {
			contract := ti._contract(activate.OrgID, activate.ContractName)
			sdk.Require(contract != nil,
				types.ErrInvalidParameter,
				fmt.Sprintf("cannot get contract with orgID:%s contractName:%s", activate.OrgID, activate.ContractName))

			ti.newTokenForThirdOrg(
				contract,
				activate.Address,
				activate.Name,
				activate.Symbol,
				activate.GasPrice,
				activate.OrgName,
				"")
		}
	}
}

///////////////////////////  check receipts function  //////////////////////////////
func (ti *TokenIssue) checkReceiptsForRecast() (*AssetChange, *std.Transfer, bool) {
	// get lock receipt
	lockAssetChange := ti.getAssetChangeReceipt(Lock)
	receipts := ti.getTransferReceipts()
	if len(receipts) != 1 {
		return nil, nil, false
	}
	lockTransfer := receipts[0]

	bCheck := ti.checkReceipts(lockAssetChange, nil, lockTransfer, nil)

	return lockAssetChange, lockTransfer, bCheck
}

func (ti *TokenIssue) checkReceiptsForConfirm() *AssetChange {
	// get lock receipt
	recastAssetChange := ti.getAssetChangeReceipt(Recast)
	lockAssetChange := ti.getAssetChangeReceipt(Lock)
	sdk.Require(recastAssetChange != nil && lockAssetChange != nil,
		types.ErrInvalidParameter, "lose asset change receipt")

	receipts := ti.getTransferReceipts()
	sdk.Require(len(receipts) == 2,
		types.ErrInvalidParameter, fmt.Sprintf("expected transfer receipt count: 2, obtain: %d", len(receipts)))

	// 解析两个给 IBC 合约账户转账的收据
	lockTransfer := receipts[0]
	recastTransfer := receipts[1]

	// 源链上给 IBC 合约账户转账收据
	sdk.Require(ti.checkReceipts(lockAssetChange, recastAssetChange, lockTransfer, recastTransfer),
		types.ErrInvalidParameter, "invalid receipts")

	return recastAssetChange
}

func (ti *TokenIssue) checkReceiptsForCancel() (*AssetChange, *std.Transfer) {
	// get lock receipt
	lockAssetChange := ti.getAssetChangeReceipt(Lock)

	receipts := ti.getTransferReceipts()
	sdk.Require(len(receipts) == 1,
		types.ErrInvalidParameter, fmt.Sprintf("expected transfer receipt count: 1, obtain: %d", len(receipts)))
	lockTransfer := receipts[0]

	sdk.Require(ti.checkReceipts(lockAssetChange, nil, lockTransfer, nil),
		types.ErrInvalidParameter, "invalid receipts")

	return lockAssetChange, lockTransfer
}

func (ti *TokenIssue) checkReceiptsForTryRecast() (*AssetChange, bool) {
	// get lock receipt
	lockAssetChange := ti.getAssetChangeReceipt(Lock)

	receipts := ti.getTransferReceipts()
	if len(receipts) != 1 {
		return nil, false
	}
	lockTransfer := receipts[0]

	bCheck := ti.checkReceipts(lockAssetChange, nil, lockTransfer, nil)

	return lockAssetChange, bCheck
}

func (ti *TokenIssue) checkReceiptsForConfirmRecast() *AssetChange {
	// get lock receipt
	recastAssetChange := ti.getAssetChangeReceipt(Recast)
	lockAssetChange := ti.getAssetChangeReceipt(Lock)
	sdk.Require(recastAssetChange != nil && lockAssetChange != nil,
		types.ErrInvalidParameter, "lose asset change receipt")

	receipts := ti.getTransferReceipts()
	sdk.Require(len(receipts) == 2,
		types.ErrInvalidParameter, fmt.Sprintf("expected transfer receipt count: 2, obtain: %d", len(receipts)))

	// 解析两个给 IBC 合约账户转账的收据
	lockTransfer := receipts[0]
	recastTransfer := receipts[1]

	// 源链上给 IBC 合约账户转账收据
	sdk.Require(ti.checkReceipts(lockAssetChange, recastAssetChange, lockTransfer, recastTransfer),
		types.ErrInvalidParameter, "invalid receipts")

	return lockAssetChange
}

func (ti *TokenIssue) checkReceiptsForCancelRecast() *AssetChange {
	// get lock receipt
	lockAssetChange := ti.getAssetChangeReceipt(Lock)
	receipts := ti.getTransferReceipts()
	sdk.Require(len(receipts) == 1,
		types.ErrInvalidParameter, fmt.Sprintf("expected transfer receipt count: 1, obtain: %d", len(receipts)))
	lockTransfer := receipts[0]

	// 检查锁定收据是否有效
	sdk.Require(ti.checkReceipts(lockAssetChange, nil, lockTransfer, nil),
		types.ErrInvalidParameter, "invalid receipts")

	return lockAssetChange
}

func (ti *TokenIssue) checkReceipts(lockAssetChange, recastAssetChange *AssetChange, lockTransfer, recastTransfer *std.Transfer) bool {
	ibcAccountAddr := ti.sdk.Helper().ContractHelper().ContractOfName("ibc").Account().Address()
	tokenAddr := ti.sdk.Message().Contract().Token()

	tokenAddrNoChainID := tokenAddr
	if strings.Contains(tokenAddr, "0") {
		tokenAddrNoChainID = tokenAddr[:strings.Index(tokenAddr, "0")]
	}
	ibcAccountNoChainID := ibcAccountAddr
	if strings.Contains(ibcAccountAddr, "0") {
		ibcAccountNoChainID = ibcAccountAddr[:strings.Index(ibcAccountAddr, "0")]
	}

	if lockAssetChange == nil || lockTransfer == nil {
		return false
	}

	// check lock receipt & lock transfer to ibc account receipt
	if lockAssetChange.From != lockTransfer.From ||
		lockAssetChange.Token != lockTransfer.Token ||
		!lockAssetChange.Value.IsEqual(lockTransfer.Value) ||
		!strings.HasPrefix(lockTransfer.To, ibcAccountNoChainID) ||
		!strings.HasPrefix(lockTransfer.Token, tokenAddrNoChainID) {
		return false
	}

	// check recast & lock receipt
	if recastAssetChange != nil && recastTransfer != nil {
		if recastAssetChange.From != lockAssetChange.From ||
			recastAssetChange.To != lockAssetChange.To ||
			!recastAssetChange.Value.IsEqual(lockAssetChange.Value) {
			return false
		}

		if recastAssetChange.To != recastTransfer.To ||
			recastAssetChange.Token != recastTransfer.Token ||
			!recastAssetChange.Value.IsEqual(recastTransfer.Value) ||
			!strings.HasPrefix(recastTransfer.From, ibcAccountNoChainID) ||
			!strings.HasPrefix(recastTransfer.Token, tokenAddrNoChainID) {
			return false
		}
	}

	return true
}

func (ti *TokenIssue) getAssetChangeReceipt(typeValue string) *AssetChange {
	var receipt *AssetChange
	forx.Range(ti.sdk.Message().InputReceipts(), func(i int, kvPair types.KVPair) bool {
		if strings.HasSuffix(string(kvPair.Key), "tokenissue.AssetChange") {
			var r std.Receipt
			err := jsoniter.Unmarshal(kvPair.Value, &r)
			sdk.RequireNotError(err, types.ErrInvalidParameter)

			var temp AssetChange
			err = jsoniter.Unmarshal(r.Bytes, &temp)
			sdk.RequireNotError(err, types.ErrInvalidParameter)

			if temp.Type == typeValue {
				receipt = &temp
				return forx.Break
			}
		}
		return true
	})

	return receipt
}

func (ti *TokenIssue) getTransferReceipts() []*std.Transfer {
	result := make([]*std.Transfer, 0)
	name := "std::transfer"

	forx.Range(ti.sdk.Message().InputReceipts(), func(i int, kvPair types.KVPair) bool {
		if strings.HasSuffix(string(kvPair.Key), name) {
			var r std.Receipt
			err := jsoniter.Unmarshal(kvPair.Value, &r)
			sdk.RequireNotError(err, types.ErrInvalidParameter)

			var t std.Transfer
			err = jsoniter.Unmarshal(r.Bytes, &t)
			sdk.RequireNotError(err, types.ErrInvalidParameter)

			result = append(result, &t)
		}
		return true
	})

	return result
}
