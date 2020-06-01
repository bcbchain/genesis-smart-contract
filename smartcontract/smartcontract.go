package smartcontract

import (
	"github.com/bcbchain/sdk/sdk/forx"
	"strconv"
	"strings"

	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/jsoniter"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
)

//SmartContract manage deploy and forbid contract
//@:contract:smartcontract
//@:version:2.4
//@:organization:orgJgaGConUyK81zibntUBjQ33PKctpk1K1G
//@:author:5e8339cb1a5cce65602fd4f57e115905348f7e83bcbe38dd77694dbe1f8903c9
type SmartContract struct {
	sdk sdk.ISmartContract

	contract              map[types.Address]std.Contract                //key=contractAddress
	contractName          map[types.Address]bool                        //key=contractAddress
	organization          map[string]std.Organization                   //key=orgId
	orgAuthDeployContract map[string]types.Address                      //key=orgId
	contractVersionList   map[string]map[string]std.ContractVersionList //key1=orgId, key2=contractName
	contractMeta          map[types.Address]std.ContractMeta            //key=contractAddress
}

//InitChain Constructor of this SmcManager
//@:constructor
func (s *SmartContract) InitChain() {

}

//UpdateChain Constructor of this SmcManager
//@:constructor
func (s *SmartContract) UpdateChain() {
	// 添加 contract all 0 key，
	// 合约：tokenbasic_cancellation，tokenbasic_foundation，tokenbasic_team，transferAgency，tokenbyb，yuebao-dc 1.0，yuebao-dc 2.0，yuebao-usdy
	// bcbtest & bcb

	s.addContractAllKey()
}

//Authorize authorize someone to deploy contracts for specific organization
//@:public:method:gas[50000]
func (s *SmartContract) Authorize(deployer types.Address, orgID string) {

	sdk.RequireOwner()
	sdk.RequireAddress(deployer)
	sdk.Require(orgID != "",
		types.ErrInvalidParameter, "Invalid orgID.")
	sdk.Require(s._chkOrganization(orgID) == true,
		types.ErrInvalidParameter, "OrgID is not exist.")
	sdk.Require(orgID != s.sdk.Helper().GenesisHelper().OrgID(),
		types.ErrInvalidParameter, "OrgID can not be genesis orgID.")

	s._setOrgAuthDeployContract(orgID, deployer)
	s.emitAuthorize(deployer, orgID)
}

//DeployContract deploy or upgrade a smart contract
//@:public:method:gas[50000]
func (s *SmartContract) DeployContract(
	name string,
	version string,
	orgID string,
	codeHash types.Hash,
	codeData []byte,
	codeDevSig string,
	codeOrgSig string,
	effectHeight int64,
	owner types.Address,
) (contractAddr types.Address) {

	// check basic params
	sdk.Require(name != "",
		types.ErrInvalidParameter, "Invalid contract name.")
	sdk.Require(orgID != "",
		types.ErrInvalidParameter, "Invalid orgID.")
	sdk.Require(s._chkOrganization(orgID) == true,
		types.ErrInvalidParameter, "OrgID is not exist.")
	sdk.Require(s.checkName(name, orgID) == true,
		types.ErrInvalidParameter, "Invalid contract name.")

	// check auth for org
	sdk.Require(s._orgAuthDeployContract(orgID) == s.sdk.Message().Sender().Address(),
		types.ErrNoAuthorization, "No authorization to deploy contract.")

	// continue check other params
	sdk.Require(version != "",
		types.ErrInvalidParameter, "Invalid version.")
	sdk.Require(len(codeHash) == 32,
		types.ErrInvalidParameter, "Invalid codeHash.")
	sdk.Require(len(codeData) != 0,
		types.ErrInvalidParameter, "Invalid codeData.")
	sdk.Require(codeDevSig != "",
		types.ErrInvalidParameter, "Invalid codeDevSig.")
	sdk.Require(codeOrgSig != "",
		types.ErrInvalidParameter, "Invalid codeOrgSig.")
	sdk.Require(effectHeight > s.sdk.Block().Height() || effectHeight == 0,
		types.ErrInvalidParameter, "Invalid effectHeight.")
	sdk.Require(owner != "",
		types.ErrInvalidParameter, "Invalid owner.")
	sdk.RequireAddress(owner)

	// check effectHeight
	if effectHeight == 0 {
		effectHeight = s.sdk.Block().Height() + 1 //next block deploy contract
	}
	contractAddr = s.sdk.Helper().BlockChainHelper().CalcContractAddress(name, version, orgID)
	codeDevSigBytes, err1 := jsoniter.Marshal(codeDevSig)
	codeOrgSigBytes, err2 := jsoniter.Marshal(codeOrgSig)
	sdk.RequireNotError(err1, types.ErrInvalidParameter)
	sdk.RequireNotError(err2, types.ErrInvalidParameter)

	// save contract
	orgCodeHash := s.checkAndForbidContractInfo(
		contractAddr,
		name,
		version,
		orgID,
		codeHash,
		codeData,
		codeDevSigBytes,
		codeOrgSigBytes,
		effectHeight,
		owner,
	)

	// save contract metadata
	contractMeta := std.ContractMeta{
		Name:         name,
		ContractAddr: contractAddr,
		OrgID:        orgID,
		Version:      version,
		EffectHeight: effectHeight,
		LoseHeight:   0,
		CodeData:     codeData,
		CodeHash:     codeHash,
		CodeDevSig:   codeDevSigBytes,
		CodeOrgSig:   codeOrgSigBytes,
	}
	s._setContractMeta(contractAddr, contractMeta)

	// update org info and delete losed effective contract address
	orgInfo := s._organization(orgID)
	newContractAddrList := make([]types.Address, 0)

	forx.Range(orgInfo.ContractAddrList, func(i int, addr types.Address) bool {
		contract := s._contract(addr)
		if contract.LoseHeight == 0 || contract.LoseHeight >= s.sdk.Block().Height() {
			newContractAddrList = append(newContractAddrList, addr)
		}
		return true
	})

	orgInfo.ContractAddrList = newContractAddrList
	orgInfo.ContractAddrList = append(orgInfo.ContractAddrList, contractAddr)
	orgInfo.OrgCodeHash = orgCodeHash
	s._setOrganization(orgInfo)

	s.emitDeployContract(
		contractAddr,
		name,
		version,
		orgID,
		codeHash,
		codeData,
		codeDevSig,
		codeOrgSig,
		effectHeight,
		owner,
	)
	return
}

//ForbidContract forbid a contract by contract address
//@:public:method:gas[50000]
func (s *SmartContract) ForbidContract(contractAddr types.Address, loseHeight int64) {

	sdk.Require(loseHeight > s.sdk.Block().Height(),
		types.ErrInvalidParameter, "Invalid loseHeight.")
	sdk.RequireAddress(contractAddr)

	sdk.Require(s._chkContract(contractAddr) == true,
		types.ErrInvalidParameter, "Contract does not exist.")
	contract := s._contract(contractAddr)

	// check auth for org
	sdk.Require(s._orgAuthDeployContract(contract.OrgID) == s.sdk.Message().Sender().Address(),
		types.ErrNoAuthorization, "No authorization to forbid contract.")

	sdk.Require(contract.LoseHeight == 0,
		types.ErrInvalidParameter, "Contract is already forbidden.")
	sdk.Require(loseHeight > contract.EffectHeight,
		types.ErrInvalidParameter, "Contract is not effective with loseHeight.")

	contract.LoseHeight = loseHeight
	s._setContract(contract)

	allMineContract := s._getMineContract()
	forx.Range(allMineContract, func(i int, v std.MineContract) bool {
		if v.Address == contractAddr {
			allMineContract = append(allMineContract[:i], allMineContract[i+1:]...)
		}
		return true
	})
	s._setMineContract(allMineContract)

	s.emitForbidContract(
		contractAddr,
		loseHeight,
	)
}

func (s *SmartContract) checkVersionAndEffectHeight(
	oldContract std.ContractVersionList, version string,
	effectHeight int64) {

	var oldVersion string
	forx.Range(oldContract.ContractAddrList, func(i int, v types.Address) bool {
		contract := s._contract(v)
		sdk.Require(contract.Address == v,
			types.ErrInvalidParameter, "Invalid contract address.")
		sdk.Require(contract.EffectHeight < effectHeight,
			types.ErrInvalidParameter, "Invalid effect height.")
		sdk.Require(compareVersion(contract.Version, version) < 0,
			types.ErrInvalidParameter, "Invalid version.")
		oldVersion = contract.Version
		return true
	})

	oldVersionSplit := strings.Split(oldVersion, ".")
	newVersionSplit := strings.Split(version, ".")
	sdk.Require(len(oldVersionSplit) == len(newVersionSplit),
		types.ErrInvalidParameter, "Invalid version.")
}

func compareVersion(v1, v2 string) int {
	if v1 == "" {
		return -1
	}
	v1s := strings.Split(v1, ".")
	v2s := strings.Split(v2, ".")
	sdk.Require(len(v1s) > 0 && len(v1s) == len(v2s),
		types.ErrInvalidParameter, "Invalid version.")

	code := 0
	forx.Range(v1s, func(i int, v string) bool {
		v1, err := strconv.Atoi(v1s[i])
		sdk.Require(err == nil, types.ErrInvalidParameter, "Invalid version.")
		v2, err := strconv.Atoi(v2s[i])
		sdk.Require(err == nil, types.ErrInvalidParameter, "Invalid version.")

		if v1 > v2 {
			code = 1
			return forx.Break
		} else if v1 < v2 {
			code = -1
			return forx.Break
		}
		return true
	})

	return code
}

func (s *SmartContract) checkAndForbidContractInfo(
	contractAddr, name, version, orgID string,
	codeHash, codeData, codeDevSigBytes, codeOrgSigBytes []byte,
	effectHeight int64,
	owner types.Address,
) (orgCodeHash []byte) {

	contractVersionInfo := s._contractVersionList(orgID, name)
	contractInfo := std.Contract{
		Address:      contractAddr,
		Account:      s.sdk.Helper().BlockChainHelper().CalcAccountFromName(name, orgID),
		Owner:        owner,
		Name:         name,
		Version:      version,
		CodeHash:     codeHash,
		EffectHeight: effectHeight,
		LoseHeight:   0,
		KeyPrefix:    "",
		Methods:      nil,
		Interfaces:   nil,
		Mine:         nil,
		IBCs:         nil,
		Token:        "",
		OrgID:        orgID,
		ChainVersion: 2,
	}

	height := strconv.FormatInt(effectHeight, 10)
	isUpgrade := false
	if s._chkContractVersionList(orgID, name) == false {
		//new contract
		if orgID == s.sdk.Helper().GenesisHelper().OrgID() {
			contractInfo.KeyPrefix = ""
		} else {
			s._setContractName(name)
			contractInfo.KeyPrefix = "/" + orgID + "/" + name
		}
	} else {
		// contract exist, upgrade it
		isUpgrade = true
		sdk.Require(contractVersionInfo.Name == name,
			types.ErrInvalidParameter, "Invalid name or orgID.")

		s.checkVersionAndEffectHeight(contractVersionInfo, version, effectHeight)

		lastContractAddr := contractVersionInfo.ContractAddrList[len(contractVersionInfo.ContractAddrList)-1]
		lastContract := s._contract(lastContractAddr)
		sdk.Require(lastContract.LoseHeight == 0,
			types.ErrInvalidParameter, "The current contract has expired and cannot be upgraded.")
		sdk.Require(lastContract.ChainVersion == 2,
			types.ErrInvalidParameter, "Only upgrade v2 chain's contract.")

		lastContract.LoseHeight = effectHeight
		s._setContract(lastContract)

		sdk.Require(owner == lastContract.Owner,
			types.ErrInvalidParameter, "Internal owner.")
		sdk.Require(contractInfo.Account == lastContract.Account,
			types.ErrInvalidParameter, "Internal contract account.")

		contractInfo.Token = lastContract.Token
		contractInfo.Owner = lastContract.Owner
		contractInfo.KeyPrefix = lastContract.KeyPrefix

		lastContractMeta := s._contractMeta(lastContractAddr)
		lastContractMeta.LoseHeight = effectHeight
		s._setContractMeta(lastContractAddr, lastContractMeta)

		s.emitForbidContract(
			lastContract.Address,
			lastContract.LoseHeight,
		)
	}

	// set contract address and effect height, when block height == effect height, bcchain will init this smart contract
	newConWithHeight := std.ContractWithEffectHeight{
		ContractAddr: contractAddr,
		IsUpgrade:    isUpgrade,
	}
	conWithHeight := s._effectHeightContractAddrs(height)
	conWithHeight = append(conWithHeight, newConWithHeight)
	s._setEffectHeightContractAddrs(height, conWithHeight)

	s.sdk.Helper().StateHelper().Flush()
	// build contract code
	buildRes := s.sdk.Helper().BuildHelper().Build(
		std.ContractMeta{
			Name:         name,
			ContractAddr: contractAddr,
			OrgID:        orgID,
			Version:      version,
			EffectHeight: effectHeight,
			LoseHeight:   0,
			CodeData:     codeData,
			CodeHash:     codeHash,
			CodeDevSig:   codeDevSigBytes,
			CodeOrgSig:   codeOrgSigBytes,
		})
	sdk.Require(buildRes.Code == types.CodeOK,
		buildRes.Code, buildRes.Error)

	orgCodeHash = buildRes.OrgCodeHash
	contractInfo.Interfaces = buildRes.Interfaces
	contractInfo.Methods = buildRes.Methods
	contractInfo.IBCs = buildRes.IBCs

	if len(buildRes.Mine) != 0 {
		contractInfo.Mine = buildRes.Mine
		sdk.Require(orgID == s.sdk.Helper().GenesisHelper().OrgID(),
			types.ErrInvalidParameter, "only genesis org can deploy mining")

		allMineContract := s._getMineContract()

		if len(allMineContract) > 0 {
			sdk.Require(len(allMineContract) == 1, types.ErrInvalidParameter, "must only one mining contract")
			lastMineAddress := allMineContract[0].Address

			lastMineContract := s._contract(lastMineAddress)
			sdk.Require(lastMineContract.Name == name && lastMineContract.OrgID == orgID,
				types.ErrInvalidParameter, "cat not deploy another mining contract")
		}
	}

	s._setContract(contractInfo)

	// set new contract address to contract list
	contractVersionInfo.Name = name
	contractVersionInfo.ContractAddrList = append(contractVersionInfo.ContractAddrList, contractAddr)
	contractVersionInfo.EffectHeights = append(contractVersionInfo.EffectHeights, effectHeight)
	s._setContractVersionList(orgID, name, contractVersionInfo)

	return
}

func (s *SmartContract) checkName(name, orgID string) bool {
	if orgID == s.sdk.Helper().GenesisHelper().OrgID() {
		// genesis org can not has the same contract name with other organizations
		return s._chkContractName(name) == false
	}

	// other organizations can not has the same contract name with genesis org
	return s._chkContractVersionList(s.sdk.Helper().GenesisHelper().OrgID(), name) == false
}
