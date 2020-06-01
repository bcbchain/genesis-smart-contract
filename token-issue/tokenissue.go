package tokenissue

import (
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/forx"
	"github.com/bcbchain/sdk/sdk/types"
)

//TokenIssue a genesis contract for issuing token and token management
//@:contract:token-issue
//@:version:2.4
//@:organization:orgJgaGConUyK81zibntUBjQ33PKctpk1K1G
//@:author:5e8339cb1a5cce65602fd4f57e115905348f7e83bcbe38dd77694dbe1f8903c9
type TokenIssue struct {
	sdk sdk.ISmartContract
}

const (
	// Define the minimum of token supply for issuing new token.
	minTotalsupply = 1000000000
	// token name can be up to 40 characters
	maxNameLen = 40
	// token symbol can be up to 20 characters
	maxSymbolLen = 20
	// maximum gas price
	maxGasPrice = 1000000000
	// maximum number of accounts for batch transfer
	maxPerBatchTransfer = 1000
	// minimum name length for genesis organization issuing new token
	minNameLenForGenesisOrg = 2
	// minimum name length for other organization issuing new token
	minNameLenForOtherOrg = 3
	// minimum symbol length for genesis organization issuing new token
	minSymbolLenForGenesisOrg = 2
	// minimum symbol length for genesis organization issuing new token
	minSymbolLenForOtherOrg = 3
)

//InitChain Constructor of this TokenIssue
//@:constructor
func (ti *TokenIssue) InitChain() {

}

//UpdateChain UpdateChain of this TokenIssue
//@:constructor
func (ti *TokenIssue) UpdateChain() {

	// fix bug peerChainBalance for yy & jiuj side chain
	//ti.updatePeerBal()

	// upgrade token's contract
	ti.upgradeTokenContract()
}

//NewToken register a token
//Notes: Once a token is registered by NewToken() call, it will own a contract exactly same with this contract,
//       but its contract cannot be used to register a new token again.
//       That means only the NewToken() of genesis contract of "token-issue" can be executed.
//       And the genesis contract "token-issue" will never own a token.
//@:public:method:gas[500000]
func (ti *TokenIssue) NewToken(
	name string,
	symbol string,
	totalSupply bn.Number,
	addSupplyEnabled bool,
	burnEnabled bool,
	gasPrice int64) (address types.Address) {

	sdk.RequireMainChain()
	sdk.Require(ti.isValidNameAndSymbol("", name, symbol, totalSupply),
		types.ErrInvalidParameter, "Invalid name or symbol")

	address = ti.sdk.Helper().BlockChainHelper().CalcContractAddress(
		ti.contractNameBRC20(name),
		ti.sdk.Message().Contract().Version(),
		ti.sdk.Message().Contract().OrgID(),
	)

	ti.newToken(
		address,
		ti.sdk.Message().Sender().Address(),
		name,
		symbol,
		totalSupply,
		addSupplyEnabled,
		burnEnabled,
		gasPrice,
		"",
	)

	return
}

//@:public:method:gas[20000]
func (ti *TokenIssue) NewTokenBRC30(
	name string,
	symbol string,
	totalSupply bn.Number,
	addSupplyEnabled bool,
	burnEnabled bool,
	gasPrice int64,
	orgName string,
) (address types.Address) {
	//判断是否处于主链
	sdk.RequireMainChain()
	sdk.Require(ti.isValidNameAndSymbol(orgName, name, symbol, totalSupply),
		types.ErrInvalidParameter, "Invalid name or symbol")

	//计算代币地址。
	address = ti.sdk.Helper().BlockChainHelper().CalcContractAddress(
		ti.contractNameBRC30(orgName, name),
		ti.sdk.Message().Contract().Version(),
		ti.sdk.Message().Contract().OrgID(),
	)

	ti.newToken(
		address,
		ti.sdk.Message().Sender().Address(),
		name,
		symbol,
		totalSupply,
		addSupplyEnabled,
		burnEnabled,
		gasPrice,
		orgName,
	)

	return
}

//Transfer transfers token to an account
//@:public:method:gas[600]
//@:public:interface:gas[60]
func (ti *TokenIssue) Transfer(to types.Address, value bn.Number) {

	// If it is the genesis contract, cannot execute this function.
	sdk.Require(ti.sdk.Message().Contract().Token() != "",
		types.ErrNoAuthorization, "The contract has not a token")

	if ti.sdk.Helper().BlockChainHelper().IsPeerChainAddress(to) {
		err := ti.sdk.Helper().TokenHelper().CheckActivate(to)
		sdk.RequireNotError(err, types.ErrInvalidParameter)

		// cross chain transfer
		ti.sdk.Helper().IBCHelper().Run(func() {
			ti.lock(to, value)
		}).Register(ti.sdk.Helper().BlockChainHelper().GetChainID(to))
	} else {
		// Do transfer
		ti.sdk.Message().Sender().TransferWithNote(to, value, ti.sdk.Tx().Note())
	}
}

//@:public:interface:gas[60]
func (ti *TokenIssue) TransferWithNote(to types.Address, value bn.Number, note string) {

	// If it is the genesis contract, cannot execute this function.
	sdk.Require(ti.sdk.Message().Contract().Token() != "",
		types.ErrNoAuthorization, "The contract has not a token")
	// Judge whether the toAddress is in the other chain
	if ti.sdk.Helper().BlockChainHelper().IsPeerChainAddress(to) {
		err := ti.sdk.Helper().TokenHelper().CheckActivate(to)
		sdk.RequireNotError(err, types.ErrInvalidParameter)

		// cross chain transfer
		ti.sdk.Helper().IBCHelper().Run(func() {
			ti.lockWithNote(to, value, note)
		}).Register(ti.sdk.Helper().BlockChainHelper().GetChainID(to))
	} else {
		// Do transfer
		ti.sdk.Message().Sender().TransferWithNote(to, value, note)
	}
}

//BatchTransfer transfers token to multi accounts
//@:public:method:gas[6000]
func (ti *TokenIssue) BatchTransfer(toList []types.Address, value bn.Number) {
	sdk.RequireMainChain()

	// If it is the genesis contract, cannot execute this function.
	sdk.Require(ti.sdk.Message().Contract().Token() != "",
		types.ErrNoAuthorization, "The contract has not a token")

	sdk.Require(len(toList) > 0,
		types.ErrInvalidParameter, "Address list cannot be empty")
	sdk.Require(len(toList) <= maxPerBatchTransfer,
		types.ErrInvalidParameter, "Number of accounts is out of range")
	forx.Range(toList, func(i int, to types.Address) bool {
		sdk.Require(!ti.sdk.Helper().BlockChainHelper().IsPeerChainAddress(to),
			types.ErrInvalidAddress, "Can not Do Inter-Blockchain-Transfer")
		return true
	})

	forx.Range(toList, func(i int, to types.Address) bool {
		ti.sdk.Message().Sender().Transfer(to, value)
		return true
	})
}

//AddSupply add token's supply after it's issued
//@:public:method:gas[2400]
func (ti *TokenIssue) AddSupply(value bn.Number) {
	sdk.RequireMainChain()

	// If it is the genesis contract, cannot execute this function.
	sdk.Require(ti.sdk.Message().Contract().Token() != "",
		types.ErrNoAuthorization, "The contract has not a token")

	sdk.Require(value.IsGreaterThanI(0),
		types.ErrInvalidParameter, "Value must greater than zero")

	newTotalSupply := ti.sdk.Helper().TokenHelper().Token().TotalSupply().Add(value)
	ti.sdk.Helper().TokenHelper().Token().SetTotalSupply(newTotalSupply)
}

//Burn burn token's supply after it's issued
//@:public:method:gas[2400]
func (ti *TokenIssue) Burn(value bn.Number) {
	sdk.RequireMainChain()

	// If it is the genesis contract, cannot execute this function.
	sdk.Require(ti.sdk.Message().Contract().Token() != "",
		types.ErrNoAuthorization, "The contract has not a token")

	sdk.Require(value.IsGreaterThanI(0),
		types.ErrInvalidParameter, "Value must greater than zero")

	newTotalSupply := ti.sdk.Helper().TokenHelper().Token().TotalSupply().Sub(value)
	sdk.Require(newTotalSupply.IsGEI(0),
		types.ErrInvalidParameter, "New totalsupply cannot be negative")

	ti.sdk.Helper().TokenHelper().Token().SetTotalSupply(newTotalSupply)
}

//SetOwner set a new owner to token after it's issued
//@:public:method:gas[2400]
func (ti *TokenIssue) SetOwner(newOnwer types.Address) {
	sdk.RequireMainChain()

	toChainIDs := ti.getNotifyChainIDs()
	if len(toChainIDs) > 0 {
		ti.sdk.Helper().IBCHelper().Run(func() {
			ti.sdk.Message().Contract().SetOwner(newOnwer)
		}).Notify(toChainIDs)
	} else {
		ti.sdk.Message().Contract().SetOwner(newOnwer)
	}
}

//SetGasPrice set token's gasprice after it's issued
//@:public:method:gas[2400]
func (ti *TokenIssue) SetGasPrice(value int64) {
	sdk.RequireMainChain()

	// If it is the genesis contract, cannot execute this function.
	sdk.Require(ti.sdk.Message().Contract().Token() != "",
		types.ErrNoAuthorization, "The contract has not a token")

	toChainIDs := ti.getNotifyChainIDs()
	if len(toChainIDs) > 0 {
		ti.sdk.Helper().IBCHelper().Run(func() {
			ti.sdk.Helper().TokenHelper().Token().SetGasPrice(value)
		}).Notify(toChainIDs)
	} else {
		ti.sdk.Helper().TokenHelper().Token().SetGasPrice(value)
	}
}

//@:public:method:gas[2400]
func (ti *TokenIssue) Activate(tokenName, chainName string) {
	ti.activate("", tokenName, chainName)
}

//@:public:method:gas[2400]
func (ti *TokenIssue) ActivateBRC30(orgName, tokenName, chainName string) {
	ti.activate(orgName, tokenName, chainName)
}
