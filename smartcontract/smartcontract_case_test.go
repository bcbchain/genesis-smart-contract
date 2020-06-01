package smartcontract

import (
	"fmt"
	"testing"

	"github.com/bcbchain/sdk/sdk/jsoniter"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
	"github.com/bcbchain/sdk/utest"

	"gopkg.in/check.v1"
)

//Test This is a function
func Test(t *testing.T) { check.TestingT(t) }

//MySuite This is a struct
type MySuite struct{}

var _ = check.Suite(&MySuite{})

//TestSmcManager_DeployContract This is a method of MySuite
// nolint gocyclo
func (mysuit *MySuite) TestSmcManager_DeployContract(c *check.C) {
	utest.Init(orgID)
	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, contractInterfaces)
	test := NewTestObject(contractOwner)
	orgID := test.obj.sdk.Helper().BlockChainHelper().CalcOrgID("testOrg")
	newOrganization := std.Organization{
		OrgID:            orgID,
		Name:             "testOrg",
		OrgOwner:         test.obj.sdk.Message().Sender().Address(),
		ContractAddrList: []types.Address{},
		OrgCodeHash:      []byte{},
	}
	test.obj.sdk.Helper().StateHelper().Set("/organization/"+orgID, &newOrganization)
	test.obj.sdk.Helper().StateHelper().Set("/organization/"+test.obj.sdk.Message().Contract().OrgID(), &std.Organization{
		OrgID:    test.obj.sdk.Message().Contract().OrgID(),
		OrgOwner: test.obj.sdk.Message().Sender().Address(),
	})

	// 部署新的合约和基础组织部署或者升级合约
	testCases := []struct {
		name         string
		version      string
		orgID        string
		codeHash     types.Hash
		codeData     []byte
		codeDevSig   string
		codeOrgSig   string
		effectHeight int64
		err          types.Error
	}{
		{"testName", "v1.0", orgID, []byte("test"), []byte("hello"), "test", "test", test.obj.sdk.Block().Height() + 2, types.Error{ErrorCode: types.ErrNoAuthorization}},
		{"", "v1.0", orgID, []byte("test"), []byte("hello"), "test", "test", test.obj.sdk.Block().Height() + 3, types.Error{ErrorCode: types.ErrInvalidParameter}},
		{"testName", "", orgID, []byte("test"), []byte("hello"), "test", "test", test.obj.sdk.Block().Height() + 3, types.Error{ErrorCode: types.ErrNoAuthorization}},
		{"testNameA", "v1.0", "", []byte("test"), []byte("hello"), "test", "test", test.obj.sdk.Block().Height() + 3, types.Error{ErrorCode: types.ErrInvalidParameter}},
		{"testNameB", "v1.0", orgID, []byte{}, []byte("hello"), "test", "test", test.obj.sdk.Block().Height() + 3, types.Error{ErrorCode: types.ErrNoAuthorization}},
		{"testNameC", "v1.0", orgID, []byte("test"), []byte{}, "test", "test", test.obj.sdk.Block().Height() + 3, types.Error{ErrorCode: types.ErrNoAuthorization}},
		{"testNameD", "v1.0", orgID, []byte("test"), []byte("hello"), "", "test", test.obj.sdk.Block().Height() + 3, types.Error{ErrorCode: types.ErrNoAuthorization}},
		{"testNameE", "v1.0", orgID, []byte("test"), []byte("hello"), "test", "", test.obj.sdk.Block().Height() + 3, types.Error{ErrorCode: types.ErrNoAuthorization}},
		{"testNameF", "v1.0", orgID, []byte("test"), []byte("hello"), "test", "test", 0, types.Error{ErrorCode: types.ErrNoAuthorization}},
		{"testNameG", "v1.0", orgID, []byte("test"), []byte("hello"), "test", "test", test.obj.sdk.Block().Height() - 2, types.Error{ErrorCode: types.ErrNoAuthorization}},
		{"testNameH", "v1.0", orgID + "test", []byte("test"), []byte("hello"), "test", "test", test.obj.sdk.Block().Height() + 50, types.Error{ErrorCode: types.ErrInvalidParameter}},
		{"testNameI", "v1.0", test.obj.sdk.Message().Contract().OrgID(), []byte("test"), []byte("hello"), "test", "test", test.obj.sdk.Block().Height() + 100, types.Error{ErrorCode: types.ErrNoAuthorization}},
		{"testName", "v1.0", test.obj.sdk.Message().Contract().OrgID(), []byte("test"), []byte("hello"), "test", "test", test.obj.sdk.Block().Height() + 200, types.Error{ErrorCode: types.ErrNoAuthorization}},
		{"testNameI", "v1.1", test.obj.sdk.Message().Contract().OrgID(), []byte("testA"), []byte("helloA"), "test", "test", test.obj.sdk.Block().Height() + 200, types.Error{ErrorCode: types.ErrNoAuthorization}},
	}

	for i, v := range testCases {
		conAddr, err := test.run().setSender(contractOwner).DeployContract(v.name, v.version, v.orgID, v.codeHash, v.codeData, v.codeDevSig, v.codeOrgSig, v.effectHeight)
		fmt.Println(i)
		utest.AssertError(err, v.err.ErrorCode)
		if err.ErrorCode == types.CodeOK {
			utest.Assert(conAddr ==
				test.obj.sdk.Helper().BlockChainHelper().CalcContractAddress(v.name, v.version, test.obj.sdk.Message().Contract().Owner().Address()))

			var preFix string
			if v.orgID != test.obj.sdk.Message().Contract().OrgID() {
				preFix = "/" + orgID + "/" + v.name
			} else {
				preFix = ""
			}
			// 检查合约信息
			utest.AssertSDB(test.obj.sdk.Message().Contract().KeyPrefix()+"/contract/"+conAddr, std.Contract{
				Address:      conAddr,
				Account:      test.obj.sdk.Helper().BlockChainHelper().CalcAccountFromName(v.name, orgID),
				Owner:        test.obj.sdk.Message().Contract().Owner().Address(),
				Name:         v.name,
				Version:      v.version,
				CodeHash:     v.codeHash,
				EffectHeight: v.effectHeight,
				LoseHeight:   0,
				KeyPrefix:    preFix,
				Methods:      nil, // 与 build 返回一致
				Interfaces:   nil,
				Token:        "",
				OrgID:        v.orgID,
			})

			// 检查合约元数据
			codeDevSigBytes, _ := jsoniter.Marshal(v.codeDevSig)
			codeOrgSigBytes, _ := jsoniter.Marshal(v.codeOrgSig)
			utest.AssertSDB(test.obj.sdk.Message().Contract().KeyPrefix()+"/contract/code/"+conAddr, std.ContractMeta{
				Name:         v.name,
				ContractAddr: conAddr,
				OrgID:        v.orgID,
				EffectHeight: v.effectHeight,
				LoseHeight:   0,
				CodeData:     v.codeData,
				CodeHash:     v.codeHash,
				CodeDevSig:   codeDevSigBytes,
				CodeOrgSig:   codeOrgSigBytes,
			})

			// 检查组织信息
			orgInfo := test.obj.sdk.Helper().StateHelper().Get("/organization/"+v.orgID, new(std.Organization)).(*std.Organization)
			found := false
			for _, addr := range orgInfo.ContractAddrList {
				if addr == conAddr {
					found = true
				}
			}
			utest.Assert(found)

			// 检查合约列表
			conList := test.obj.sdk.Helper().StateHelper().Get("/organization/"+v.orgID, new(std.ContractVersionList)).(*std.ContractVersionList)
			foundCon := false
			for _, addr := range conList.ContractAddrList {
				if addr == conAddr {
					foundCon = true
				}
			}
			utest.Assert(foundCon)

			// 检查上一个合约的失效高度
			conVersionInfo := *test.obj.sdk.Helper().StateHelper().Get("/contract/"+v.orgID+"/"+v.name, new(std.ContractVersionList)).(*std.ContractVersionList)
			if len(conVersionInfo.ContractAddrList) > 1 {
				lastCon := test.obj.sdk.Helper().StateHelper().Get(
					"/contract/"+conVersionInfo.ContractAddrList[len(conVersionInfo.ContractAddrList)-2], new(std.Contract)).(*std.Contract)
				utest.Assert(lastCon.LoseHeight == v.effectHeight)
			}
		}
	}

	// 升级合约
	oldContract := &std.Contract{
		Name:      "oldName",
		Address:   "oldContractAddr",
		Version:   "v1.1.1",
		OrgID:     orgID,
		Token:     "oldToken",
		Account:   "testAccount",
		Owner:     "testOwner",
		KeyPrefix: "/oldName",
	}
	test.obj.sdk.Helper().StateHelper().Set("/contract/oldContractAddr", oldContract)

	test.obj.sdk.Helper().StateHelper().Set("/contract/"+orgID+"/"+"oldName", &std.ContractVersionList{
		Name:             "oldName",
		ContractAddrList: []types.Address{"oldContractAddr"},
		EffectHeights:    []int64{5},
	})

	// 升级合约
	testCasesForUpgrade := []struct {
		name         string
		version      string
		orgID        string
		codeHash     types.Hash
		codeData     []byte
		codeDevSig   string
		codeOrgSig   string
		effectHeight int64
		err          types.Error
	}{
		{"oldName", "v1.1.2", orgID, []byte("testA"), []byte("helloA"), "test", "test", test.obj.sdk.Block().Height() + 20, types.Error{ErrorCode: types.ErrNoAuthorization}},
		{"oldName", "v1.1.2", orgID, []byte("testA"), []byte("helloB"), "test", "test", test.obj.sdk.Block().Height() + 21, types.Error{ErrorCode: types.ErrNoAuthorization}},
		{"oldName", "v1.2.4.1", orgID, []byte("testA"), []byte("helloB"), "test", "test", test.obj.sdk.Block().Height() + 22, types.Error{ErrorCode: types.ErrNoAuthorization}},
		{"oldName", "v1.2.4", orgID, []byte("testA"), []byte("helloB"), "test", "test", test.obj.sdk.Block().Height(), types.Error{ErrorCode: types.ErrNoAuthorization}},
		{"oldName", "v1.2.2", orgID + "test", []byte("testA"), []byte("helloB"), "test", "test", test.obj.sdk.Block().Height() + 26, types.Error{ErrorCode: types.ErrInvalidParameter}},
		{"oldName", "v1.0.2", orgID, []byte("testA"), []byte("helloB"), "test", "test", test.obj.sdk.Block().Height() + 190, types.Error{ErrorCode: types.ErrNoAuthorization}},
	}

	for _, v := range testCasesForUpgrade {
		conAddr, err := test.run().setSender(contractOwner).DeployContract(v.name, v.version, v.orgID, v.codeHash, v.codeData, v.codeDevSig, v.codeOrgSig, v.effectHeight)
		utest.AssertError(err, v.err.ErrorCode)
		if err.ErrorCode == types.CodeOK {
			utest.Assert(conAddr ==
				test.obj.sdk.Helper().BlockChainHelper().CalcContractAddress(v.name, v.version, test.obj.sdk.Message().Contract().Owner().Address()))
			utest.AssertSDB(test.obj.sdk.Message().Contract().KeyPrefix()+"/contract/"+conAddr, std.Contract{
				Address:      conAddr,
				Account:      oldContract.Account,
				Owner:        oldContract.Owner,
				Name:         v.name,
				Version:      v.version,
				CodeHash:     v.codeHash,
				EffectHeight: v.effectHeight,
				LoseHeight:   0,
				KeyPrefix:    "/" + v.name,
				Methods:      nil, // 与 build 返回一致
				Interfaces:   nil,
				Token:        oldContract.Token,
				OrgID:        v.orgID,
			})

			// 检查合约元数据
			codeDevSigBytes, _ := jsoniter.Marshal(v.codeDevSig)
			codeOrgSigBytes, _ := jsoniter.Marshal(v.codeOrgSig)
			utest.AssertSDB(test.obj.sdk.Message().Contract().KeyPrefix()+"/contract/code/"+conAddr, std.ContractMeta{
				Name:         v.name,
				ContractAddr: conAddr,
				OrgID:        v.orgID,
				EffectHeight: v.effectHeight,
				LoseHeight:   0,
				CodeData:     v.codeData,
				CodeHash:     v.codeHash,
				CodeDevSig:   codeDevSigBytes,
				CodeOrgSig:   codeOrgSigBytes,
			})

			// 检查组织信息
			orgInfo := test.obj.sdk.Helper().StateHelper().Get("/organization/"+v.orgID, new(std.Organization)).(*std.Organization)
			found := false
			for _, conAddr := range orgInfo.ContractAddrList {
				if conAddr == conAddr {
					found = true
				}
			}
			utest.Assert(found)

			// 检查合约列表
			conList := test.obj.sdk.Helper().StateHelper().Get("/organization/"+v.orgID, new(std.ContractVersionList)).(*std.ContractVersionList)
			foundCon := false
			for _, addr := range conList.ContractAddrList {
				if addr == conAddr {
					foundCon = true
				}
			}
			utest.Assert(foundCon)

			// 检查上一个合约的失效高度
			conVersionInfo := *test.obj.sdk.Helper().StateHelper().Get("/contract/"+v.orgID+"/"+v.name, new(std.ContractVersionList)).(*std.ContractVersionList)
			if len(conVersionInfo.ContractAddrList) > 1 {
				lastCon := test.obj.sdk.Helper().StateHelper().Get(
					"/contract/"+conVersionInfo.ContractAddrList[len(conVersionInfo.ContractAddrList)-2], new(std.Contract)).(*std.Contract)
				utest.Assert(lastCon.LoseHeight == v.effectHeight)
			}
		}
	}
}

//TestSmcManager_ForbidInternalContract This is a method of MySuite
func (mysuit *MySuite) TestSmcManager_ForbidInternalContract(c *check.C) {
	utest.Init(orgID)
	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, contractInterfaces)
	test := NewTestObject(contractOwner)

	contract := std.Contract{
		Address:      test.obj.sdk.Helper().BlockChainHelper().CalcContractAddress("a", "1.0", orgID),
		CodeHash:     []byte{},
		EffectHeight: 0,
		LoseHeight:   0,
		Methods:      nil,
		Interfaces:   nil,
	}
	test.obj.sdk.Helper().StateHelper().Set("/contract/"+contract.Address, &contract)

	contractHasForbid := std.Contract{
		Address:      "testForbid",
		CodeHash:     []byte{},
		EffectHeight: 0,
		LoseHeight:   5,
		Methods:      nil,
		Interfaces:   nil,
	}
	test.obj.sdk.Helper().StateHelper().Set("/contract/"+contractHasForbid.Address, &contractHasForbid)

	testCases := []struct {
		err          types.Error
		contractAddr types.Address
		effectHeight int64
	}{
		{types.Error{ErrorCode: types.ErrNoAuthorization}, contract.Address, test.obj.sdk.Block().Height() + 3},
		{types.Error{ErrorCode: types.ErrInvalidAddress}, "", test.obj.sdk.Block().Height() + 4},
		{types.Error{ErrorCode: types.ErrInvalidAddress}, contract.Address + "test", test.obj.sdk.Block().Height() + 5},
		{types.Error{ErrorCode: types.ErrInvalidAddress}, contractHasForbid.Address, test.obj.sdk.Block().Height() + 5},
		{types.Error{ErrorCode: types.ErrInvalidParameter}, contractHasForbid.Address, test.obj.sdk.Block().Height() - 1},
	}

	for i, v := range testCases {
		err := test.run().setSender(contractOwner).ForbidInternalContract(v.contractAddr, v.effectHeight)
		fmt.Println(i)
		utest.AssertError(err, v.err.ErrorCode)
		if err.ErrorCode == types.CodeOK {
			con := *test.obj.sdk.Helper().StateHelper().Get("/contract/"+contract.Address, new(std.Contract)).(*std.Contract)
			utest.Assert(con.LoseHeight == v.effectHeight)
		}
	}
}
