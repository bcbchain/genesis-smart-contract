package blacklist

import (
	"github.com/bcbchain/sdk/sdk/types"
	"github.com/bcbchain/sdk/utest"
	"gopkg.in/check.v1"
	"testing"
)

//Test This is a function
func Test(t *testing.T) { check.TestingT(t) }

//MySuite This is a struct
type MySuite struct{}

var _ = check.Suite(&MySuite{})

//TestBlacklist_AddAddress This is a method of MySuite
func (mysuit *MySuite) TestBlacklist_AddAddress(c *check.C) {
	utest.Init(orgID)
	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, contractInterfaces)
	test := NewTestObject(contractOwner)
	test.setSender(contractOwner).InitChain()

	testCases := []struct {
		addressList []types.Address
		err         types.Error
		desc        string
	}{
		{[]types.Address{}, types.Error{ErrorCode: types.ErrInvalidParameter, ErrorDesc: "Cannot be empty blacklist."}, "设置的地址列表为空"},
		{[]types.Address{"a"}, types.Error{ErrorCode: types.ErrInvalidAddress, ErrorDesc: "Address chainID is error! "}, "无效的地址"},
		{[]types.Address{"localKvG4ayU644JD7BHhEVmP5sof2Lekopj5K", "localJgaGConUyK81zibntUBjQ33PKctpk1K1G"}, types.Error{ErrorCode: types.CodeOK, ErrorDesc: ""}, "正常用例"},
		{[]types.Address{"localKvG4ayU644JD7BHhEVmP5sof2Lekopj5K"}, types.Error{ErrorCode: types.CodeOK, ErrorDesc: ""}, "正常用例"},
		{[]types.Address{contractOwner.Address()}, types.Error{ErrorCode: types.ErrInvalidParameter, ErrorDesc: "Cannot contain owner address in blacklist"}, "地址不能为合约拥有者地址"},
	}

	for _, v := range testCases {
		err := test.run().setSender(contractOwner).AddAddress(v.addressList)
		utest.AssertError(err, v.err.ErrorCode)
		if err.ErrorCode != types.CodeOK {
			utest.AssertErrorMsg(err, v.err.ErrorDesc)
		} else {
			for _, a := range v.addressList {
				utest.Assert(test.obj.sdk.Helper().StateHelper().Check("/blacklist/" + a))
				utest.AssertEquals(*test.obj.sdk.Helper().StateHelper().GetEx("/blacklist/"+a, new(string)).(*string), "true")
			}
		}
	}
}

//TestBlacklist_DelAddress This is a method of MySuite
func (mysuit *MySuite) TestBlacklist_DelAddress(c *check.C) {
	utest.Init(orgID)
	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, contractInterfaces)
	test := NewTestObject(contractOwner)
	test.setSender(contractOwner).InitChain()
	testCases := []struct {
		addressList []types.Address
		err         types.Error
		desc        string
	}{
		{[]types.Address{}, types.Error{ErrorCode: types.ErrInvalidParameter, ErrorDesc: "Cannot be empty blacklist."}, "设置的地址列表为空"},
		{[]types.Address{"a"}, types.Error{ErrorCode: types.ErrInvalidAddress, ErrorDesc: "Address chainID is error! "}, "无效的地址"},
		{[]types.Address{"localKvG4ayU644JD7BHhEVmP5sof2Lekopj5K", "localJgaGConUyK81zibntUBjQ33PKctpk1K1G"}, types.Error{ErrorCode: types.CodeOK, ErrorDesc: ""}, "正常用例"},
		{[]types.Address{"localKvG4ayU644JD7BHhEVmP5sof2Lekopj5K"}, types.Error{ErrorCode: types.CodeOK, ErrorDesc: ""}, "正常用例"},
		{[]types.Address{contractOwner.Address()}, types.Error{ErrorCode: types.ErrInvalidParameter, ErrorDesc: "Cannot contain owner address in blacklist"}, "地址不能为合约拥有者地址"},
	}

	for _, v := range testCases {
		err := test.run().setSender(contractOwner).DelAddress(v.addressList)
		utest.AssertError(err, v.err.ErrorCode)
		if err.ErrorCode != types.CodeOK {
			utest.AssertErrorMsg(err, v.err.ErrorDesc)
		} else {
			for _, a := range v.addressList {
				utest.Assert(test.obj.sdk.Helper().StateHelper().Check("/blacklist/" + a))
				utest.AssertEquals(*test.obj.sdk.Helper().StateHelper().GetEx("/blacklist/"+a, new(string)).(*string), "false")
			}
		}
	}
}

//TestBlacklist_SetOwner This is a method of MySuite
func (mysuit *MySuite) TestBlacklist_SetOwner(c *check.C) {
	utest.Init(orgID)
	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, contractInterfaces)
	test := NewTestObject(contractOwner)
	test.setSender(contractOwner).InitChain()
}
