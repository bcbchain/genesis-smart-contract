package tokenissue

import (
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/forx"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
)

func (ti *TokenIssue) _chkSideChainID(chainID string) bool {
	return ti.sdk.Helper().StateHelper().Check(keyOfChainInfo(chainID))
}

func (ti *TokenIssue) _setAccountBalance(account, token types.Address, balance bn.Number) {
	key := std.KeyOfAccountToken(account, token)
	info := std.AccountInfo{
		Address: ti.sdk.Message().Contract().Token(),
		Balance: balance,
	}
	ti.sdk.Helper().StateHelper().Set(key, &info)
}

func (ti *TokenIssue) _accountBalance(account, token types.Address) bn.Number {
	key := std.KeyOfAccountToken(account, token)
	return ti.sdk.Helper().StateHelper().GetEx(key, new(std.AccountInfo)).(*std.AccountInfo).Balance
}

func (ti *TokenIssue) _peerChainBalance(tokenAddr types.Address, chainID string) bn.Number {
	key := keyOfPeerChainBal(tokenAddr, chainID)
	bn0 := bn.N(0)
	return *ti.sdk.Helper().StateHelper().GetEx(key, &bn0).(*bn.Number)
}

func (ti *TokenIssue) _setPeerChainBalance(tokenAddr types.Address, chainID string, balance bn.Number) {
	key := keyOfPeerChainBal(tokenAddr, chainID)
	ti.sdk.Helper().StateHelper().Set(key, &balance)
}

func (ti *TokenIssue) _chainInfo(chainID string) *ChainInfo {
	return ti.sdk.Helper().StateHelper().GetEx(keyOfChainInfo(chainID), new(ChainInfo)).(*ChainInfo)
}

func (ti *TokenIssue) _supportSideChains(tokenAddr types.Address) []string {
	return *ti.sdk.Helper().StateHelper().GetEx(keyOfSupportSCList(tokenAddr), new([]string)).(*[]string)
}

func (ti *TokenIssue) _setSupportSideChains(tokenAddr types.Address, sideChainIDs []string) {
	ti.sdk.Helper().StateHelper().Set(keyOfSupportSCList(tokenAddr), &sideChainIDs)
}

// _contract get contract information with orgID and name
func (ti *TokenIssue) _contract(orgID, name string) *std.Contract {
	contractVersions := ti._contractVersions(orgID, name)

	var contract *std.Contract
	forx.RangeReverse(contractVersions.ContractAddrList, func(index int, contractAddr types.Address) bool {
		if ti.sdk.Block().Height() >= contractVersions.EffectHeights[index] {
			key := std.KeyOfContract(contractAddr)
			tempCon := ti.sdk.Helper().StateHelper().GetEx(key, new(std.Contract)).(*std.Contract)
			if tempCon.LoseHeight != 0 && tempCon.LoseHeight < ti.sdk.Block().Height() {
				return forx.Continue
			}

			contract = tempCon
		}

		return true
	})

	return contract
}

// _contractVersions get contract version list with orgID and name
func (ti *TokenIssue) _contractVersions(orgID, name string) *std.ContractVersionList {
	key := std.KeyOfContractsWithName(orgID, name)
	defaultValue := std.ContractVersionList{
		Name:             name,
		ContractAddrList: make([]types.Address, 0),
		EffectHeights:    make([]int64, 0),
	}
	contractVersions := ti.sdk.Helper().StateHelper().GetEx(key, &defaultValue)

	return contractVersions.(*std.ContractVersionList)
}

func (ti *TokenIssue) _allToken() (value []types.Address) {
	return *ti.sdk.Helper().StateHelper().McGetEx(std.KeyOfAllToken(), new([]types.Address)).(*[]types.Address)
}

func (ti *TokenIssue) _sideChainSupportTokens(sideChainID string) []string {
	return *ti.sdk.Helper().StateHelper().GetEx(keyOfSideChainSupportTokens(sideChainID), new([]string)).(*[]string)
}

func (ti *TokenIssue) _setSideChainSupportTokens(sideChainID string, tokens []types.Address) {
	ti.sdk.Helper().StateHelper().Set(keyOfSideChainSupportTokens(sideChainID), &tokens)
}

func (ti *TokenIssue) _BRC30TokenAddr(orgName, tokenName string) types.Address {
	return *ti.sdk.Helper().StateHelper().Get(std.KeyOfBRC30TokenWithName(orgName, tokenName), new(types.Address)).(*types.Address)
}

// Organization
func (ti *TokenIssue) _chkOrganization(orgID string) bool {
	return ti.sdk.Helper().StateHelper().Check(keyOfOrganization(orgID))
}
