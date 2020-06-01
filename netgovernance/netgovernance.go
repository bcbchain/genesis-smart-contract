package netgovernance

import (
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/types"
)

//NetGovernance This is struct of contract
//@:contract:netgovernance
//@:version:2.3
//@:organization:orgJgaGConUyK81zibntUBjQ33PKctpk1K1G
//@:author:5e8339cb1a5cce65602fd4f57e115905348f7e83bcbe38dd77694dbe1f8903c9
type NetGovernance struct {
	sdk       sdk.ISmartContract
	chainInfo *ChainInfo
}

//InitChain Constructor of this NetGovernance
//@:constructor
func (ng *NetGovernance) InitChain() {
	sdk.RequireMainChain()
}

//UpdateChain Constructor of this NetGovernance
//@:constructor
func (ng *NetGovernance) UpdateChain() {
	sdk.RequireMainChain()
}

//@:public:method:gas[50000]
func (ng *NetGovernance) RegisterSideChain(chainName, orgName string, ownerAddr types.Address) {
	// check sender
	sdk.RequireOwner()

	// check chainName
	sideChainID := ng.sdk.Helper().BlockChainHelper().CalcSideChainID(chainName)

	var ifInit, ifExits bool
	if ifExits = ng._chkChainInfo(sideChainID); ifExits {
		ifInit = ng.ifSideChainInit(sideChainID)
	}

	sdk.Require(!ifExits || ifInit,
		types.ErrInvalidParameter, "ChainName has been occupied ")

	// check organization
	ng.checkOrganization(orgName)

	// check ownerAddr
	ng.checkOwnerAddr(ownerAddr)

	// save chainInfo
	chainInfo := ChainInfo{
		SideChainName: chainName,
		ChainID:       sideChainID,
		OrgName:       orgName,
		Owner:         ownerAddr,
		Status:        Init,
		AddrVersion:   AddressVersion2,
	}
	ng._setChainInfo(chainInfo)

	//save sideChainID
	if !ifExits {
		sideChainIDList := ng._sideChainIDs()
		sideChainIDList = append(sideChainIDList, sideChainID)
		ng._setSideChainIDs(sideChainIDList)
	}

	// send receipt
	ng.emitRegisterSideChainReceipt(sideChainID, orgName, ownerAddr)
}

//@:public:method:gas[5000000]
func (ng *NetGovernance) GenesisSideChain(
	chainName string,
	nodeName string,
	nodePubKey types.PubKey,
	rewardAddr types.Address,
	openURL string,
	gasPriceRatio string) {

	mainOpenURLs := ng._openURLs(ng.sdk.Block().ChainID())
	sdk.Require(len(mainOpenURLs) != 0,
		types.ErrInvalidParameter, "main chain must set openURLs")

	// check chainName
	sideChainID, chainInfo := ng.checkSideChainInfo(chainName, Init, Disabled)

	// check nodeName
	ng.checkNodeName(nodeName)

	// check nodePubKey
	nodeAddr := ng.checkNodePubKey(nodePubKey, chainName)

	// check rewardAddr
	sdk.RequireAddressEx(sideChainID, rewardAddr)

	// check openurl
	openURLs := []string{openURL}
	ng.checkOpenUrls(openURLs)

	// check gasPriceRatio
	gasPriceRatio = ng.checkGasPriceRatio(gasPriceRatio)

	// 重置侧链环境
	ng.resetChainEnv(sideChainID)

	// save urls
	ng._setOpenURLs(sideChainID, openURLs)

	chainInfo.NodeNames = []string{nodeName}
	chainInfo.Status = Ready
	chainInfo.Height = ng.sdk.Block().Height()
	chainInfo.GasPriceRatio = gasPriceRatio
	ng._setChainInfo(chainInfo)

	// generate validator
	validator := Validator{
		PubKey:     nodePubKey,
		Power:      10,
		RewardAddr: rewardAddr,
		Name:       nodeName,
		NodeAddr:   nodeAddr,
	}

	//save Validator
	nodes := make(map[string]Validator)
	nodes[nodeAddr] = validator
	ng._setChainValidator(sideChainID, nodes)

	// send receipt
	genesisInfo, contractsData := ng.makeGenesisInfo(chainInfo, validator)
	ng.emitGenesisSideChainReceipt(
		sideChainID,
		openURLs,
		genesisInfo,
		contractsData,
		chainInfo.AddrVersion)
}

//@:public:method:gas[50000]
func (ng *NetGovernance) SetOpenURLs(chainName string, openURLs []string) {
	//check chainInfo
	var chainID string
	if chainName != ng.sdk.Block().ChainID() {
		chainID, _ = ng.checkSideChainInfo(chainName, Ready, "")
	} else {
		chainID = ng.checkMainChainInfo()
	}

	//check openURLs
	ng.checkOpenUrls(openURLs)

	//save openURLs
	ng._setOpenURLs(chainID, openURLs)

	// broadcast if chainID is mainChain
	if chainID == ng.sdk.Helper().BlockChainHelper().GetMainChainID() {
		ng.sdk.Helper().IBCHelper().Run(func() {
			ng.emitSetOpenURLReceipt(chainID, openURLs)
		}).Broadcast()
	} else {
		//send receipt
		ng.emitSetOpenURLReceipt(chainID, openURLs)
	}
}

//@:public:method:gas[50000]
func (ng *NetGovernance) SetStatus(chainName string, status string) {
	sdk.RequireOwner()

	// check chain name
	chainID := ng.checkChainName(chainName, true)

	// check status
	chainInfo := ng.checkStatus(chainID, status)

	//save chain status
	chainInfo.Status = status
	ng._setChainInfo(chainInfo)

	//send receipt
	ng.emitSetStatusReceipt(chainID, status)
}

//@:public:method:gas[50000]
func (ng *NetGovernance) SetGasPriceRatio(chainName string, gasPriceRatio string) {
	//check sideChainInfo
	chainID, chainInfo := ng.checkSideChainInfo(chainName, Ready, "")

	//check gasPriceRatio
	ng.checkGasPriceRatio(gasPriceRatio)

	// update chainInfo
	chainInfo.GasPriceRatio = gasPriceRatio
	ng._setChainInfo(chainInfo)

	toChainIDs := make([]string, 0)
	toChainIDs = append(toChainIDs, chainID)
	ng.sdk.Helper().IBCHelper().Run(func() {
		ng.emitSetGasPriceRatioReceipt(chainInfo.SideChainName, chainID, gasPriceRatio)
	}).Notify(toChainIDs)
}
