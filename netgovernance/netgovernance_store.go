package netgovernance

import (
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
)

// ChainInfo
func (ng *NetGovernance) _setChainInfo(ci ChainInfo) {
	ng.sdk.Helper().StateHelper().Set(keyOfChainInfo(ci.ChainID), &ci)
}

func (ng *NetGovernance) _chainInfo(chainID string) ChainInfo {
	return *ng.sdk.Helper().StateHelper().Get(keyOfChainInfo(chainID), new(ChainInfo)).(*ChainInfo)
}

func (ng *NetGovernance) _chkChainInfo(chainID string) bool {
	return ng.sdk.Helper().StateHelper().Check(keyOfChainInfo(chainID))
}

// Organization
func (ng *NetGovernance) _chkOrganization(orgID string) bool {
	return ng.sdk.Helper().StateHelper().Check(keyOfOrganization(orgID))
}

// ChainValidatorPubKeys
func (ng *NetGovernance) _setChainValidator(chainID string, cvp map[string]Validator) {
	ng.sdk.Helper().StateHelper().Set("/ibc/"+chainID, &cvp)
}

// Check BCB validator
func (ng *NetGovernance) _chkBCBValidator(nodeAddr types.Address) bool {
	return ng.sdk.Helper().StateHelper().Check(keyOfBCBValidator(nodeAddr))
}

// Get BCB validator
func (ng *NetGovernance) _bcbValidator() []string {
	return *ng.sdk.Helper().StateHelper().GetEx("/validators/all/0", new([]string)).(*[]string)
}

// set open URL
func (ng *NetGovernance) _setOpenURLs(chainID string, urls []string) {
	ng.sdk.Helper().StateHelper().Set(keyOfOpenURLs(chainID), &urls)
}

// Get open URL
func (ng *NetGovernance) _openURLs(chainID string) []string {
	return *ng.sdk.Helper().StateHelper().GetEx(keyOfOpenURLs(chainID), new([]string)).(*[]string)
}

// contract code
func (ng *NetGovernance) _contractCode(contractAddr types.Address) []byte {
	key := keyOfContractCode(contractAddr)
	contractMeta := ng.sdk.Helper().StateHelper().GetEx(key, new(std.ContractMeta)).(*std.ContractMeta)
	return contractMeta.CodeData
}

// validator
func (ng *NetGovernance) _setSideChainIDs(sideChainIDs []string) {
	ng.sdk.Helper().StateHelper().Set(keyOfSideChainIDs(), &sideChainIDs)
}

func (ng *NetGovernance) _sideChainIDs() []string {
	return *ng.sdk.Helper().StateHelper().GetEx(keyOfSideChainIDs(), new([]string)).(*[]string)
}

func (ng *NetGovernance) _chainVersion() int64 {
	type appState struct {
		ChainVersion int64 `json:"chain_version,omitempty"` //当前链版本
	}

	return ng.sdk.Helper().StateHelper().Get(keyOfWorldAppState(), new(appState)).(*appState).ChainVersion
}

func (ng *NetGovernance) _sequence(queueID string) uint64 {
	return ng.sdk.Helper().StateHelper().GetUint64(keyOfSequence(queueID))
}

func (ng *NetGovernance) _delSequence(queueID string) {
	ng.sdk.Helper().StateHelper().Delete(keyOfSequence(queueID))
}

func (ng *NetGovernance) _delSequenceHeight(queueID string, seq uint64) {
	ng.sdk.Helper().StateHelper().Delete(keyOfSequenceHeight(queueID, seq))
}

func (ng *NetGovernance) _delLastQueueHash(queueID string) {
	ng.sdk.Helper().StateHelper().Delete(keyOfLastQueueHash(queueID))
}

func (ng *NetGovernance) _delPeerChainBal(token, chainID string) {
	ng.sdk.Helper().StateHelper().Delete(keyOfPeerChainBal(token, chainID))
}

func (ng *NetGovernance) _allToken() []types.Address {
	return *ng.sdk.Helper().StateHelper().GetEx(std.KeyOfAllToken(), new([]types.Address)).(*[]types.Address)
}

func (ng *NetGovernance) _sideChainSupportTokens(sideChainID string) []string {
	return *ng.sdk.Helper().StateHelper().GetEx(keyOfSideChainSupportTokens(sideChainID), new([]string)).(*[]string)
}

func (ng *NetGovernance) _delSideChainSupportTokens(sideChainID string) {
	ng.sdk.Helper().StateHelper().Delete(keyOfSideChainSupportTokens(sideChainID))
}

func (ng *NetGovernance) _setSideChainSupportTokens(sideChainID string, tokens []types.Address) {
	ng.sdk.Helper().StateHelper().Set(keyOfSideChainSupportTokens(sideChainID), &tokens)
}

func (ng *NetGovernance) _supportSideChains(tokenAddr types.Address) []string {
	return *ng.sdk.Helper().StateHelper().GetEx(keyOfSupportSCList(tokenAddr), new([]string)).(*[]string)
}

func (ng *NetGovernance) _setSupportSideChains(tokenAddr types.Address, sideChainIDs []string) {
	ng.sdk.Helper().StateHelper().Set(keyOfSupportSCList(tokenAddr), &sideChainIDs)
}
