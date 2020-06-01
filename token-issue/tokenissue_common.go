package tokenissue

import (
	"github.com/bcbchain/sdk/sdk/forx"
	"strings"
)

func (ti *TokenIssue) isCallByIBC() bool {
	origins := ti.sdk.Message().Origins()
	if len(origins) < 1 {
		return false
	}

	lastContract := ti.sdk.Helper().ContractHelper().ContractOfAddress(origins[len(origins)-1])
	if lastContract.Name() == "ibc" && lastContract.OrgID() == ti.sdk.Helper().GenesisHelper().OrgID() {
		return true
	}

	return false
}

func (ti *TokenIssue) getNotifyChainIDs() []string {
	toChainIDs := ti._supportSideChains(ti.sdk.Message().Contract().Token())

	newToChainIDs := []string{}
	forx.Range(toChainIDs, func(index int, toChainID string) bool {
		chainInfo := ti._chainInfo(toChainID)
		if chainInfo.Status != "" &&
			chainInfo.Status != "disabled" &&
			!inSlice(newToChainIDs, toChainID) {

			newToChainIDs = append(newToChainIDs, toChainID)
		}

		return forx.Continue
	})

	return newToChainIDs
}

func inSlice(slice []string, item string) bool {
	exist := false
	forx.Range(slice, func(index int, s string) bool {
		if item == s {
			exist = true
			return forx.Break
		}
		return forx.Continue
	})
	return exist
}

func (ti *TokenIssue) getChainName() string {
	chainID := ti.sdk.Block().ChainID()
	return chainID[strings.Index(chainID, "[")+1 : len(chainID)-1]
}

func (ti *TokenIssue) contractNameBRC20(name string) string {
	return "token-templet-" + name
}

func (ti *TokenIssue) contractNameBRC30(orgName, name string) string {
	return "token-templet-" + orgName + "-" + name
}
