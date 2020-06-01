package organization

import (
	"fmt"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
	"github.com/bcbchain/sdk/utest"
	"testing"

	"gopkg.in/check.v1"
)

//Test This is a function
func Test(t *testing.T) { check.TestingT(t) }

//MySuite This is a struct
type MySuite struct{}

var _ = check.Suite(&MySuite{})

//TestOrganizationMgr_RegisterOrganization This is a method of MySuite
func (mysuit *MySuite) TestOrganizationMgr_RegisterOrganization(c *check.C) {
	utest.Init(orgID)
	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, nil)
	test := NewTestObject(contractOwner)

	var longName string
	for i := 0; i < 257; i++ {
		longName = longName + "a"
	}

	testCases := []struct {
		name string
		err  types.Error
	}{
		// 正常用例
		{"testOrg", types.Error{ErrorCode: types.CodeOK}},
		// 名字已经存在
		{"testOrg", types.Error{ErrorCode: types.ErrInvalidParameter}},
		// name 为空
		{"", types.Error{ErrorCode: types.ErrInvalidParameter}},
		// name 长度超过 256 字符
		{longName, types.Error{ErrorCode: types.ErrInvalidParameter}},
		// name 必须 utf-8 编码
		{string([]byte{0xff, 0xfe, 0xfd}), types.Error{ErrorCode: types.ErrInvalidParameter}},
	}

	for _, testCase := range testCases {
		res, err := test.run().RegisterOrganization(testCase.name)
		utest.AssertError(err, testCase.err.ErrorCode)
		if testCase.err.ErrorCode == types.CodeOK {
			utest.Assert(len(res) > 0)
			newOrgID := test.obj.sdk.Helper().BlockChainHelper().CalcOrgID(testCase.name)
			fmt.Println(newOrgID)
			utest.AssertSDB(test.obj.sdk.Message().Contract().KeyPrefix()+"/organization/"+newOrgID, std.Organization{
				OrgID:            newOrgID,
				Name:             testCase.name,
				OrgOwner:         test.obj.sdk.Message().Sender().Address(),
				ContractAddrList: []types.Address{},
				OrgCodeHash:      []byte{},
				Signers:          []types.PubKey{},
			})
		}
	}
}

func (mysuit *MySuite) TestOrganizationMgr_SetSigners(c *check.C) {
	utest.Init(orgID)
	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, nil)
	test := NewTestObject(contractOwner)

	var longName string
	for i := 0; i < 257; i++ {
		longName = longName + "a"
	}

	orgID, _ := test.run().setSender(contractOwner).RegisterOrganization("testOrg")

	testCases := []struct {
		orgID string
		pk    []types.PubKey
		err   types.Error
	}{
		{"", nil, types.Error{ErrorCode: types.ErrInvalidParameter}},
		{"test", nil, types.Error{ErrorCode: types.ErrInvalidParameter}},
		{orgID, nil, types.Error{ErrorCode: types.ErrInvalidParameter}},
	}

	for _, c := range testCases {
		err := test.run().setSender(contractOwner).SetSigners(c.orgID, c.pk)
		utest.AssertError(err, c.err.ErrorCode)
	}
}
