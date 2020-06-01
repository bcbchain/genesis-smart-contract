package netgovernance

import (
	"encoding/hex"
	"fmt"
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/forx"
	"github.com/bcbchain/sdk/sdk/jsoniter"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
	"regexp"
	"sort"
)

func (ng *NetGovernance) makeGenesisInfo(
	chainInfo ChainInfo,
	validator Validator) (string, []ContractData) {

	genesisTokenAddr := ng.sdk.Helper().GenesisHelper().Token().Address()
	baseToken := ng.sdk.Helper().TokenHelper().TokenOfAddress(genesisTokenAddr)
	contracts, contractsData := ng.getContractsForGenesis(chainInfo)
	mainChainValidators := ng.getMainChainValidators()
	chainVersion := ng._chainVersion()

	var mainChainInfo MainChainInfo
	mainChainUrls := ng._openURLs(ng.sdk.Block().ChainID())
	mainChainInfo.OpenUrls = mainChainUrls
	mainChainInfo.Validators = mainChainValidators

	chainName := chainInfo.SideChainName
	bcHelper := ng.sdk.Helper().BlockChainHelper()
	genesis := GenesisInfo{
		ChainID:      chainInfo.ChainID,
		ChainVersion: fmt.Sprintf("%d", chainVersion),
		GenesisTime:  ng.sdk.Helper().BlockChainHelper().FormatTime(ng.sdk.Block().Time(), "2006-01-02T15:04:05.999999999Z07:00"),
		AppHash:      "",
		AppState: AppState{
			Organization:  "genesis",
			GasPriceRatio: chainInfo.GasPriceRatio,
			Token: std.Token{
				Address:          bcHelper.RecalcAddressEx(genesisTokenAddr, chainName),
				Owner:            bcHelper.RecalcAddressEx(baseToken.Owner().Address(), chainName),
				Name:             baseToken.Name(),
				Symbol:           baseToken.Symbol(),
				TotalSupply:      bn.N(0),
				AddSupplyEnabled: baseToken.AddSupplyEnabled(),
				BurnEnabled:      baseToken.BurnEnabled(),
				GasPrice:         baseToken.GasPrice(),
			},
			RewardStrategy: []Reward{
				{
					Name:          "validators",
					RewardPercent: "100.00",
					Address:       "",
				},
			},
			Contracts: contracts,
			OrgBind: OrgBind{
				OrgName: chainInfo.OrgName,
				Owner:   bcHelper.RecalcAddressEx(chainInfo.Owner, chainName),
			},
			MainChain: mainChainInfo,
		},
		Validators: []Validator{validator},
	}

	info, err := jsoniter.Marshal(genesis)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	return string(info), contractsData
}

func (ng *NetGovernance) getContractsForGenesis(chainInfo ChainInfo) (contracts []Contract, contractData []ContractData) {

	genesisContracts := []string{"token-basic", "token-issue", "governance",
		"organization", "smartcontract", "ibc"}

	forx.Range(genesisContracts, func(i int, genesisContract string) bool {
		c := ng.sdk.Helper().ContractHelper().ContractOfName(genesisContract)

		contractOwner := ""
		if genesisContract == "governance" {
			contractOwner = ng.sdk.Helper().BlockChainHelper().RecalcAddressEx(chainInfo.Owner, chainInfo.SideChainName)
		}

		codeDevSig, codeOrgSig := ng.getConCodeSig(c.Address())
		code := c.Name() + "-" + c.Version() + ".tar.gz"
		con := Contract{
			Name:       c.Name(),
			Version:    c.Version(),
			Code:       code,
			CodeHash:   hex.EncodeToString(c.CodeHash()),
			Owner:      contractOwner,
			CodeDevSig: codeDevSig,
			CodeOrgSig: codeOrgSig,
		}
		contracts = append(contracts, con)

		conData := ContractData{
			Name:     c.Name(),
			Version:  c.Version(),
			CodeByte: ng._contractCode(c.Address()),
		}
		contractData = append(contractData, conData)

		return forx.Continue
	})

	return contracts, contractData
}

// 公钥对应的节点不能是主链的节点
func (ng *NetGovernance) checkNodePubKey(nodePubKey types.PubKey, sideChainName string) types.Address {
	mainChainAddr := ng.sdk.Helper().BlockChainHelper().CalcAccountFromPubKey(nodePubKey)

	sdk.Require(!ng._chkBCBValidator(mainChainAddr),
		types.ErrInvalidParameter, "Can not use mainChain validator")

	sideChainNodeAddr := ng.sdk.Helper().BlockChainHelper().RecalcAddressEx(mainChainAddr, sideChainName)

	return sideChainNodeAddr
}

func (ng *NetGovernance) checkOrganization(orgName string) {

	sdk.Require(len(orgName) > 0,
		types.ErrInvalidParameter, "Invalid orgName")

	genesisOrgID := ng.sdk.Helper().GenesisHelper().OrgID()
	orgID := ng.sdk.Helper().BlockChainHelper().CalcOrgID(orgName)
	sdk.Require(orgID != genesisOrgID,
		types.ErrInvalidParameter, "SideChain organization could not be genesis organization")

	sdk.Require(ng._chkOrganization(orgID),
		types.ErrInvalidParameter, "There is no organization with name "+orgName)
}

func (ng *NetGovernance) getConCodeSig(conAddr types.Address) (codeDevSig, codeOrgSig Signature) {

	type contractMeta struct {
		Name         string        `json:"name"`
		ContractAddr types.Address `json:"contractAddr"`
		OrgID        string        `json:"orgID"`
		Version      string        `json:"version"`
		EffectHeight int64         `json:"effectHeight"`
		LoseHeight   int64         `json:"loseHeight"`
		CodeData     []byte        `json:"codeData"`
		CodeHash     []byte        `json:"codeHash"`
		CodeDevSig   []byte        `json:"codeDevSig"`
		CodeOrgSig   []byte        `json:"codeOrgSig"`
	}

	key := std.KeyOfContractCode(conAddr)
	conMeta := ng.sdk.Helper().StateHelper().Get(key, new(contractMeta)).(*contractMeta)

	var codeDevSigStr, codeOrgSigStr string
	err := jsoniter.Unmarshal(conMeta.CodeDevSig, &codeDevSigStr)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	err = jsoniter.Unmarshal(conMeta.CodeOrgSig, &codeOrgSigStr)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	codeDevSigByte := []byte(codeDevSigStr)
	codeOrgSigByte := []byte(codeOrgSigStr)

	err = jsoniter.Unmarshal(codeDevSigByte, &codeDevSig)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	err = jsoniter.Unmarshal(codeOrgSigByte, &codeOrgSig)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	return codeDevSig, codeOrgSig
}

func (ng *NetGovernance) getMainChainValidators() map[string]Validator {

	mainChainValidatorAddrList := ng._bcbValidator()
	mainChainValidators := make(map[string]Validator)
	forx.Range(mainChainValidatorAddrList, func(i int, addr string) {
		valInfo := *ng.sdk.Helper().StateHelper().GetEx("/validator/"+addr,
			&Validator{}).(*Validator)

		if valInfo.Power != 0 {
			mainChainValidators[addr] = valInfo
		}
	})

	return mainChainValidators
}

func (ng *NetGovernance) checkChainName(chainName string, occupiedWanted bool) (chainID string) {
	sdk.Require(chainName != ng.sdk.Block().ChainID(),
		types.ErrInvalidParameter, "ChainName should not be mainChain")

	chainID = ng.sdk.Helper().BlockChainHelper().CalcSideChainID(chainName)

	var errInfo string
	if occupiedWanted == false {
		errInfo = "ChainName has been occupied "
	} else {
		errInfo = "ChainName does not exits "
	}
	sdk.Require(ng._chkChainInfo(chainID) == occupiedWanted,
		types.ErrInvalidParameter, errInfo)

	return
}

func (ng *NetGovernance) checkOwnerAddr(ownerAddr types.Address) {
	//检查地址格式
	sdk.RequireAddress(ownerAddr)

	//侧链Owner地址不能为主链委员会地址
	genesisOwnerAddr := ng.sdk.Message().Contract().Owner().Address()
	sdk.Require(ownerAddr != genesisOwnerAddr,
		types.ErrInvalidParameter, "Invalid owner Address")
}

func (ng *NetGovernance) checkOpenUrls(openURLs []string) {

	sdk.Require(len(openURLs) > 0,
		types.ErrInvalidParameter, "OpenUrls should not be empty")

	urlExpr := `^(https|http)://`
	forx.Range(openURLs, func(i int, openURL string) bool {
		match, _ := regexp.MatchString(urlExpr, openURL)
		sdk.Require(match,
			types.ErrInvalidParameter, fmt.Sprintf("Invalid openURL: %v ", openURL))

		return forx.Continue
	})
}

func (ng *NetGovernance) checkSideChainInfo(chainName, statusWanted, otherStatus string) (chainID string, chainInfo ChainInfo) {
	// check chainName
	chainID = ng.checkChainName(chainName, true)

	// check sender
	sdk.Require(ng.sdk.Message().Sender().Address() == ng._chainInfo(chainID).Owner,
		types.ErrNoAuthorization, "Only SideChain Owner can do this ")

	// check chain status
	chainInfo = ng._chainInfo(chainID)
	sdk.Require(chainInfo.Status == statusWanted || chainInfo.Status == otherStatus,
		types.ErrInvalidParameter, "The status of sideChain is error ")

	return
}

func (ng *NetGovernance) checkMainChainInfo() string {
	//check Sender
	sdk.RequireOwner()

	return ng.sdk.Helper().BlockChainHelper().GetMainChainID()
}

func (ng *NetGovernance) checkNodeName(nodeName string) {
	sdk.Require(len(nodeName) > 0 && len(nodeName) <= MaxNameLen,
		types.ErrInvalidParameter, fmt.Sprintf("The length of nodeName should be within (0,%d]", MaxNameLen))
}

func (ng *NetGovernance) checkStatus(chainID, status string) ChainInfo {
	chainInfo := ng._chainInfo(chainID)
	sdk.Require((status == Ready && chainInfo.Status == Clear) ||
		(status == Clear && chainInfo.Status == Ready) ||
		(status == Disabled && chainInfo.Status == Clear),
		types.ErrInvalidParameter, "Status value error!")

	return chainInfo
}

func (ng *NetGovernance) checkGasPriceRatio(gasPriceRatio string) string {
	//检查为空
	if gasPriceRatio == "" {
		return "1.000"
	}

	//检查格式精确到小数点后三位
	urlExpr := `^[0-9]+(.[0-9]{3})?$`
	match, _ := regexp.MatchString(urlExpr, gasPriceRatio)
	sdk.Require(match,
		types.ErrInvalidParameter, fmt.Sprintf("Invalid gasPriceRatio: %v ", gasPriceRatio))

	return gasPriceRatio
}

func (ng *NetGovernance) resetChainEnv(chainID string) {
	queueID := ng.sdk.Block().ChainID() + "->" + chainID
	ng.delLastQueueHashAndSeq(queueID)

	queueID1 := chainID + "->" + ng.sdk.Block().ChainID()
	ng.delLastQueueHashAndSeq(queueID1)

	ng.updateChainTokenInfo(chainID)
}

func (ng *NetGovernance) updateChainTokenInfo(chainID string) {
	tokens := ng._sideChainSupportTokens(chainID)

	forx.Range(tokens, func(i int, addr string) bool {

		// delete token peer chain balance
		ng._delPeerChainBal(addr, chainID)

		scList := ng._supportSideChains(addr)
		forx.Range(scList, func(i int, scId string) bool {
			if scId == chainID {
				toChainIDs := make([]string, 0, len(scList)-1)
				index := sort.SearchStrings(scList, scId)

				temp1 := scList[:index]
				temp2 := scList[index-1:]
				toChainIDs = append(temp1, temp2...)

				// update token support side chain ID list
				ng._setSupportSideChains(addr, toChainIDs)
				return forx.Break
			}

			return forx.Continue
		})

		return forx.Continue
	})

	if len(tokens) > 0 {
		ng.sdk.Helper().IBCHelper().Run(func() {
			ng.emitRemoveSideChainToken(chainID, tokens)
		}).Broadcast()
	}

	// delete side chain support tokens
	ng._delSideChainSupportTokens(chainID)

	// delete genesis token peer chain balance
	ng._delPeerChainBal(ng.sdk.Helper().GenesisHelper().Token().Address(), chainID)
}

func (ng *NetGovernance) delLastQueueHashAndSeq(queueID string) {
	ng._delLastQueueHash(queueID)

	seq := ng._sequence(queueID)
	forx.Range(int(seq), func(index int) {
		ng._delSequenceHeight(queueID, uint64(index)+1)
	})

	ng._delSequence(queueID)
}

func (ng *NetGovernance) ifSideChainInit(sideChainID string) bool {
	chainInfo := ng._chainInfo(sideChainID)
	if chainInfo.Status != Init {
		return false
	}

	return true
}
