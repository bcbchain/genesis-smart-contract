package netgovernance

import (
	"github.com/bcbchain/sdk/sdk/types"
)

func (ng *NetGovernance) emitRegisterSideChainReceipt(sideChainID, orgName string, ownerAddr types.Address) {
	type registerSideChain struct {
		SideChainID string        `json:"sideChainID"`
		OrgName     string        `json:"orgName"`
		OwnerAddr   types.Address `json:"ownerAddress"`
	}
	receipt := registerSideChain{
		SideChainID: sideChainID,
		OrgName:     orgName,
		OwnerAddr:   ownerAddr,
	}
	ng.sdk.Helper().ReceiptHelper().Emit(receipt)
}

func (ng *NetGovernance) emitGenesisSideChainReceipt(sideChainID string,
	openURLs []string, genesisInfo string,
	contractsData []ContractData, AddrVersion AddressVersion) {
	type genesisSideChain struct {
		SideChainID  string         `json:"sideChainID"`
		OpenURLs     []string       `json:"openURLs"`
		GenesisInfo  string         `json:"genesisInfo"`
		ContractData []ContractData `json:"contractData"`
		AddrVersion  AddressVersion `json:"addrVersion"`
	}

	receipt := genesisSideChain{
		SideChainID:  sideChainID,
		OpenURLs:     openURLs,
		GenesisInfo:  genesisInfo,
		ContractData: contractsData,
		AddrVersion:  AddrVersion,
	}
	ng.sdk.Helper().ReceiptHelper().Emit(receipt)
}

func (ng *NetGovernance) emitSetOpenURLReceipt(sideChainID string, openURLs []string) {
	type setOpenURL struct {
		SideChainID string   `json:"sideChainID"`
		OpenURLs    []string `json:"openURLs"`
	}
	receipt := setOpenURL{
		SideChainID: sideChainID,
		OpenURLs:    openURLs,
	}
	ng.sdk.Helper().ReceiptHelper().Emit(receipt)
}

func (ng *NetGovernance) emitSetStatusReceipt(sideChainID string, status string) {
	type setStatus struct {
		SideChainID string `json:"sideChainID"`
		Status      string `json:"status"`
	}
	receipt := setStatus{
		SideChainID: sideChainID,
		Status:      status,
	}
	ng.sdk.Helper().ReceiptHelper().Emit(receipt)
}

func (ng *NetGovernance) emitSetGasPriceRatioReceipt(chainName, chainID, gasPriceRatio string) {
	type setGasPriceRatio struct {
		ChainName     string `json:"chainName"`
		ChainID       string `json:"chainID"`
		GasPriceRatio string `json:"gasPriceRatio"`
	}
	receipt := setGasPriceRatio{
		ChainName:     chainName,
		ChainID:       chainID,
		GasPriceRatio: gasPriceRatio,
	}
	ng.sdk.Helper().ReceiptHelper().Emit(receipt)
}

func (ng *NetGovernance) emitRemoveSideChainToken(chainID string, tokenAddrs []types.Address) {
	type removeSideChainToken struct {
		ChainID    string          `json:"chainID"`
		TokenAddrs []types.Address `json:"tokenAddrs"`
	}
	receipt := removeSideChainToken{
		ChainID:    chainID,
		TokenAddrs: tokenAddrs,
	}
	ng.sdk.Helper().ReceiptHelper().Emit(receipt)
}
