package smartcontract

import (
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
)

// contract
func (s *SmartContract) _contract(addr string) std.Contract {
	return *s.sdk.Helper().StateHelper().GetEx("/contract/"+addr, new(std.Contract)).(*std.Contract)
}

func (s *SmartContract) _setContract(contract std.Contract) {
	s.sdk.Helper().StateHelper().Set("/contract/"+contract.Address, &contract)
}

func (s *SmartContract) _chkContract(contractAddr string) bool {
	return s.sdk.Helper().StateHelper().Check("/contract/" + contractAddr)
}

// contractName
func (s *SmartContract) _contractName(name string) bool {
	t := false
	return *s.sdk.Helper().StateHelper().GetEx("/contract/thirdparty/name/"+name, &t).(*bool)
}

func (s *SmartContract) _setContractName(name string) {
	t := true
	s.sdk.Helper().StateHelper().Set("/contract/thirdparty/name/"+name, &t)
}

func (s *SmartContract) _chkContractName(name string) bool {
	return s.sdk.Helper().StateHelper().Check("/contract/thirdparty/name/" + name)
}

// organization
func (s *SmartContract) _organization(orgID string) std.Organization {
	return *s.sdk.Helper().StateHelper().GetEx("/organization/"+orgID, new(std.Organization)).(*std.Organization)
}

func (s *SmartContract) _setOrganization(org std.Organization) {
	s.sdk.Helper().StateHelper().Set("/organization/"+org.OrgID, &org)
}

func (s *SmartContract) _chkOrganization(orgID string) bool {
	return s.sdk.Helper().StateHelper().Check("/organization/" + orgID)
}

// orgAuthDeployContract
func (s *SmartContract) _orgAuthDeployContract(orgID string) types.Address {
	return *s.sdk.Helper().StateHelper().GetEx("/organization/"+orgID+"/auth", new(types.Address)).(*types.Address)
}

func (s *SmartContract) _setOrgAuthDeployContract(orgID string, addr types.Address) {
	s.sdk.Helper().StateHelper().Set("/organization/"+orgID+"/auth", &addr)
}

func (s *SmartContract) _chkOrgAuthDeployContract(orgID string) bool {
	return s.sdk.Helper().StateHelper().Check("/organization/" + orgID + "/auth")
}

//contractVersionList
func (s *SmartContract) _contractVersionList(orgID, name string) std.ContractVersionList {
	return *s.sdk.Helper().StateHelper().GetEx("/contract/"+orgID+"/"+name, new(std.ContractVersionList)).(*std.ContractVersionList)
}

func (s *SmartContract) _setContractVersionList(orgID, name string, contractVersionList std.ContractVersionList) {
	s.sdk.Helper().StateHelper().Set("/contract/"+orgID+"/"+name, &contractVersionList)
}

func (s *SmartContract) _chkContractVersionList(orgID, name string) bool {
	return s.sdk.Helper().StateHelper().Check("/contract/" + orgID + "/" + name)
}

// contractMeta
func (s *SmartContract) _contractMeta(contractAddr string) std.ContractMeta {
	return *s.sdk.Helper().StateHelper().GetEx("/contract/code/"+contractAddr, new(std.ContractMeta)).(*std.ContractMeta)
}

func (s *SmartContract) _setContractMeta(contractAddr types.Address, contractMeta std.ContractMeta) {
	s.sdk.Helper().StateHelper().Set("/contract/code/"+contractAddr, &contractMeta)
}

func (s *SmartContract) _chkContractMeta(contractAddr string) bool {
	return s.sdk.Helper().StateHelper().Check("/contract/code/" + contractAddr)
}

func (s *SmartContract) _effectHeightContractAddrs(height string) (contractWithHeight []std.ContractWithEffectHeight) {
	return *s.sdk.Helper().StateHelper().GetEx("/"+height, &contractWithHeight).(*[]std.ContractWithEffectHeight)
}

func (s *SmartContract) _setEffectHeightContractAddrs(height string, contractWithHeight []std.ContractWithEffectHeight) {
	s.sdk.Helper().StateHelper().Set("/"+height, &contractWithHeight)
}

func (s *SmartContract) _setMineContract(mineContract []std.MineContract) {
	s.sdk.Helper().StateHelper().Set(std.KeyOfMineContracts(), &mineContract)
}

func (s *SmartContract) _getMineContract() (mineContract []std.MineContract) {
	return *s.sdk.Helper().StateHelper().GetEx(std.KeyOfMineContracts(), &mineContract).(*[]std.MineContract)
}

func (s *SmartContract) _setAllContractAddr(allContractAddr []types.Address) {
	s.sdk.Helper().StateHelper().Set(std.KeyOfAllContracts(), &allContractAddr)
}
