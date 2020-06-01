package tokenissue

import (
	"encoding/hex"
	"fmt"
	"github.com/bcbchain/sdk/sdkimpl/object"
	"testing"

	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
	"github.com/bcbchain/sdk/utest"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (mysuit *MySuite) TestTokenIssue_NewToken(c *C) {
	utest.Init(orgID)
	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, contractMethods)
	test := NewTestObject(contractOwner)

	var testcases = []struct {
		name        string
		symbol      string
		totalSupply bn.Number
		gasprice    int64
		err         types.Error
	}{
		//测试用例： 名称为空
		{"", "mycn", bn.N(1e15), 5000, types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 名称超过长度限制
		{"mycoinmycoinmycoinmycoinmycoinmycoinmycoin", "mycn", bn.N(1e15), 5000, types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 符号为空
		{"mycoin", "", bn.N(1e15), 5000, types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 符号超出长度限制
		{"mycoin", "mycoinmycoinmycoinmycoin", bn.N(1e15), 5000, types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 燃料价格为0
		{"mycoin", "mycn", bn.N(1e12), 0, types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 燃料价格小于基础燃料价格
		{"mycoin", "mycn", bn.N(1e12), 1000, types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 燃料价格大于最大限制
		{"mycoin", "mycn", bn.N(1e12), 1e10, types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 发行量小于最小限制
		{"mycoin", "mycn", bn.N(1e8), 5000, types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 发行量为1， 小于最小限制
		{"mycoin", "mycn", bn.N(1), 5000, types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 发行量为负数
		{"mycoin", "mycn", bn.N(-1), 5000, types.Error{ErrorCode: types.ErrInvalidParameter}},

		//测试用例： 发行量为0，抢注
		{"mycoin", "mycn", bn.N(0), 5000, types.Error{ErrorCode: types.CodeOK}},
		//测试用例： 正常用例，抢注后重新发行
		{"mycoin", "mycn", bn.N(1e15), 5000, types.Error{ErrorCode: types.CodeOK}},

		//测试用例： 重复发行
		{"mycoin", "mycn", bn.N(1e15), 5000, types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 名称被占用
		{"mycoin", "mycn1", bn.N(1e15), 5000, types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 符号被占用
		{"mycoin1", "mycn", bn.N(1e15), 5000, types.Error{ErrorCode: types.ErrInvalidParameter}},
	}
	for _, t := range testcases {
		addr, err := test.run().setSender(contractOwner).NewToken(t.name, t.symbol, t.totalSupply, true, true, t.gasprice)
		utest.AssertError(err, t.err.ErrorCode)
		if t.err.ErrorCode == types.CodeOK {
			utest.Assert(addr != "")
			utest.AssertSDB(std.KeyOfToken(addr), std.Token{
				Address:          addr,
				Owner:            contractOwner.Address(),
				Name:             t.name,
				Symbol:           t.symbol,
				TotalSupply:      t.totalSupply,
				AddSupplyEnabled: true,
				BurnEnabled:      true,
				GasPrice:         t.gasprice,
			})
		} else {
			utest.Assert(addr == "")
		}
	}

	//测试用例： 已发行的代币合约执行NewToken()
	utest.NextBlock(1)
	tokencontract := test.obj.sdk.Helper().ContractHelper().ContractOfName("token-template-mycoin")
	test.obj.sdk.Message().(*object.Message).SetContract(tokencontract)
	_, err := test.run().setSender(contractOwner).NewToken("mycoin", "mycn", bn.N(1e15), true, true, 5000)
	utest.AssertError(err, types.ErrNoAuthorization)
}

func (mysuit *MySuite) TestTokenIssue_Transfer(c *C) {
	utest.Init(orgID)
	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, contractMethods)
	test := NewTestObject(contractOwner)

	addr, err := test.run().setSender(contractOwner).NewToken("mycoin", "mycn", bn.N(1e15), true, true, 5000)
	utest.AssertOK(err)
	utest.Assert(addr != "")

	pubkey, _ := hex.DecodeString("FFE0014B0B08BB79B17B996ECABEDA6BF02534B64917631BB5DE59FB411B1083")
	toaddr := test.obj.sdk.Helper().AccountHelper().AccountOfPubKey(pubkey).Address()

	var testcases = []struct {
		from  types.Address
		to    types.Address
		value bn.Number
		err   types.Error
	}{
		//测试用例： 账户没钱，自己转给自己
		{toaddr, toaddr, bn.N(1e5), types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 账户余额不足
		{toaddr, contractOwner.Address(), bn.N(1e5), types.Error{ErrorCode: types.ErrInsufficientBalance}},
		//测试用例： 正常转账
		{contractOwner.Address(), toaddr, bn.N(1e5), types.Error{ErrorCode: types.CodeOK}},
		//测试用例： 部分转账后，账户余额不足
		{contractOwner.Address(), toaddr, bn.N(1e15), types.Error{ErrorCode: types.ErrInsufficientBalance}},
		//测试用例： 部分转账后，自己转给自己
		{contractOwner.Address(), contractOwner.Address(), bn.N(1e5), types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 转账金额为负数
		{contractOwner.Address(), toaddr, bn.N(-100), types.Error{ErrorCode: types.ErrInvalidParameter}},
	}
	//test Transfer() of token-issue contract
	//测试用例： 调用代币发行合约进行转账
	utest.AssertError(test.setSender(contractOwner).Transfer(toaddr, bn.N(1e4)), types.ErrNoAuthorization)

	tokencontract := test.obj.sdk.Helper().ContractHelper().ContractOfToken(addr)
	test.obj.sdk.Message().(*object.Message).SetContract(tokencontract)

	for _, t := range testcases {
		sender := test.obj.sdk.Helper().AccountHelper().AccountOf(t.from)
		utest.AssertError(test.setSender(sender).Transfer(t.to, t.value), t.err.ErrorCode)
	}
}
func (mysuit *MySuite) TestTokenIssue_BatchTransfer(c *C) {
	utest.Init(orgID)
	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, contractMethods)
	test := NewTestObject(contractOwner)

	addr, err := test.run().setSender(contractOwner).NewToken("mycoin", "mycn", bn.N(1e15), true, true, 5000)
	utest.AssertOK(err)
	utest.Assert(addr != "")

	pubkey, _ := hex.DecodeString("FFE0014B0B08BB79B17B996ECABEDA6BF02534B64917631BB5DE59FB411B1083")
	toaddr := test.obj.sdk.Helper().AccountHelper().AccountOfPubKey(pubkey).Address()

	var testcases = []struct {
		from  types.Address
		to    types.Address
		value bn.Number
		err   types.Error
	}{
		//测试用例： 账户余额不足，自己转给自己
		{toaddr, toaddr, bn.N(1e5), types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 账户余额不足
		{toaddr, contractOwner.Address(), bn.N(1e5), types.Error{ErrorCode: types.ErrInsufficientBalance}},
		//测试用例： 正常转账
		{contractOwner.Address(), toaddr, bn.N(1e5), types.Error{ErrorCode: types.CodeOK}},
		//测试用例： 部分转账后，账户余额不足
		{contractOwner.Address(), toaddr, bn.N(1e15), types.Error{ErrorCode: types.ErrInsufficientBalance}},
		//测试用例： 部分转账后，自己转给自己
		{contractOwner.Address(), contractOwner.Address(), bn.N(1e5), types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 转账金额为负数
		{contractOwner.Address(), toaddr, bn.N(-100), types.Error{ErrorCode: types.ErrInvalidParameter}},
	}
	toList := make([]types.Address, 1)
	toList[0] = toaddr
	//test BatchTransfer() of token-issue contract
	//测试用例： 调用代币发行合约进行转账
	utest.AssertError(test.setSender(contractOwner).BatchTransfer(toList, bn.N(1e4)), types.ErrNoAuthorization)

	tokencontract := test.obj.sdk.Helper().ContractHelper().ContractOfToken(addr)
	test.obj.sdk.Message().(*object.Message).SetContract(tokencontract)

	toList = make([]types.Address, 0)
	utest.AssertError(test.setSender(contractOwner).BatchTransfer(toList, bn.N(1e5)), types.ErrInvalidParameter)

	for _, t := range testcases {
		sender := test.obj.sdk.Helper().AccountHelper().AccountOf(t.from)
		toList = make([]types.Address, 0)
		toList = append(toList, t.to)
		utest.AssertError(test.setSender(sender).BatchTransfer(toList, t.value), t.err.ErrorCode)
	}

	toList = make([]types.Address, 0)
	for i := 0; i < maxPerBatchTransfer+1; i++ {
		toList = append(toList, toaddr)
	}
	utest.AssertError(test.setSender(contractOwner).BatchTransfer(toList, bn.N(1e5)), types.ErrInvalidParameter)
}
func (mysuit *MySuite) TestTokenIssue_AddSupply(c *C) {
	utest.Init(orgID)
	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, contractMethods)
	test := NewTestObject(contractOwner)
	orgCon := test.obj.sdk.Message().Contract()

	addr, err := test.run().setSender(contractOwner).NewToken("mycoin", "mycn", bn.N(1e15), true, true, 5000)
	utest.AssertOK(err)
	utest.Assert(addr != "")

	pubkey, _ := hex.DecodeString("FFE0014B0B08BB79B17B996ECABEDA6BF02534B64917631BB5DE59FB411B1083")
	toaddr := test.obj.sdk.Helper().AccountHelper().AccountOfPubKey(pubkey).Address()

	var testcases = []struct {
		from  types.Address
		value bn.Number
		err   types.Error
	}{
		//测试用例： 非所有者执行增发
		{toaddr, bn.N(1e5), types.Error{ErrorCode: types.ErrNoAuthorization}},
		//测试用例： 所有者执行增发，增发金额为负数
		{contractOwner.Address(), bn.N(-100), types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 正常用例
		{contractOwner.Address(), bn.N(1e12), types.Error{ErrorCode: types.CodeOK}},
	}
	//test AddSupply() of token-issue contract
	//测试用例： 调用代币发行合约进行增发
	utest.AssertError(test.setSender(contractOwner).AddSupply(bn.N(1e4)), types.ErrNoAuthorization)

	tokencontract := test.obj.sdk.Helper().ContractHelper().ContractOfToken(addr)
	test.obj.sdk.Message().(*object.Message).SetContract(tokencontract)

	for _, t := range testcases {
		sender := test.obj.sdk.Helper().AccountHelper().AccountOf(t.from)
		utest.AssertError(test.setSender(sender).AddSupply(t.value), t.err.ErrorCode)
	}

	//测试用例：发行不支持燃烧的代币，执行燃烧
	test.obj.sdk.Message().(*object.Message).SetContract(orgCon)
	addr, err = test.run().setSender(contractOwner).NewToken("noaddsupply", "noadd", bn.N(1e15), false, true, 5000)
	utest.AssertOK(err)
	utest.Assert(addr != "")
	utest.NextBlock(1)
	tokencontract = test.obj.sdk.Helper().ContractHelper().ContractOfToken(addr)
	test.obj.sdk.Message().(*object.Message).SetContract(tokencontract)
	utest.AssertError(test.setSender(contractOwner).AddSupply(bn.N(1e10)), types.ErrAddSupplyNotEnabled)
}
func (mysuit *MySuite) TestTokenIssue_Burn(c *C) {
	utest.Init(orgID)
	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, contractMethods)
	test := NewTestObject(contractOwner)
	orgCon := test.obj.sdk.Message().Contract()

	addr, err := test.run().setSender(contractOwner).NewToken("mycoin", "mycn", bn.N(1e15), true, true, 5000)
	utest.AssertOK(err)
	utest.Assert(addr != "")

	pubkey, _ := hex.DecodeString("FFE0014B0B08BB79B17B996ECABEDA6BF02534B64917631BB5DE59FB411B1083")
	toaddr := test.obj.sdk.Helper().AccountHelper().AccountOfPubKey(pubkey).Address()

	var testcases = []struct {
		from  types.Address
		value bn.Number
		err   types.Error
	}{
		//测试用例： 非所有者执行燃烧
		{toaddr, bn.N(1e5), types.Error{ErrorCode: types.ErrNoAuthorization}},
		//测试用例： 所有者执行燃烧，燃烧金额为负数
		{contractOwner.Address(), bn.N(-100), types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 所有者执行燃烧，燃烧金额为总发行金额
		{contractOwner.Address(), bn.N(1e15), types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 正常燃烧
		{contractOwner.Address(), bn.N(1e12), types.Error{ErrorCode: types.CodeOK}},
		//测试用例： 部分燃烧后，再次执行全部燃烧
		{contractOwner.Address(), bn.N(1e15 - 1e12), types.Error{ErrorCode: types.ErrInvalidParameter}},
	}
	//test Burn() of token-issue contract
	//测试用例： 调用代币发行合约进行燃烧
	utest.AssertError(test.setSender(contractOwner).Burn(bn.N(1e4)), types.ErrNoAuthorization)

	tokencontract := test.obj.sdk.Helper().ContractHelper().ContractOfToken(addr)
	test.obj.sdk.Message().(*object.Message).SetContract(tokencontract)

	for _, t := range testcases {
		sender := test.obj.sdk.Helper().AccountHelper().AccountOf(t.from)
		utest.AssertError(test.setSender(sender).Burn(t.value), t.err.ErrorCode)
	}

	//测试用例：发行不支持燃烧的代币，执行燃烧
	test.obj.sdk.Message().(*object.Message).SetContract(orgCon)
	addr, err = test.run().setSender(contractOwner).NewToken("noburn", "nob", bn.N(1e15), true, false, 5000)
	utest.AssertOK(err)
	utest.Assert(addr != "")
	utest.NextBlock(1)
	tokencontract = test.obj.sdk.Helper().ContractHelper().ContractOfToken(addr)
	test.obj.sdk.Message().(*object.Message).SetContract(tokencontract)
	utest.AssertError(test.setSender(contractOwner).Burn(bn.N(1e10)), types.ErrBurnNotEnabled)
}
func (mysuit *MySuite) TestTokenIssue_SetOwner(c *C) {
	utest.Init(orgID)
	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, contractMethods)
	test := NewTestObject(contractOwner)

	addr, err := test.run().setSender(contractOwner).NewToken("mycoin", "mycn", bn.N(1e15), true, true, 5000)
	utest.AssertOK(err)
	utest.Assert(addr != "")

	pubkey, _ := hex.DecodeString("FFE0014B0B08BB79B17B996ECABEDA6BF02534B64917631BB5DE59FB411B1083")
	toaddr := test.obj.sdk.Helper().AccountHelper().AccountOfPubKey(pubkey).Address()

	var testcases = []struct {
		from     types.Address
		newOwner types.Address
		err      types.Error
	}{
		//测试用例： 非所有者调用
		{toaddr, toaddr, types.Error{ErrorCode: types.ErrNoAuthorization}},
		//测试用例： 所有者调用，newowner地址为空
		{contractOwner.Address(), "", types.Error{ErrorCode: types.ErrInvalidAddress}},
		//测试用例： 正常用例
		{contractOwner.Address(), toaddr, types.Error{ErrorCode: types.CodeOK}},
		//测试用例： 原所有者再次调用
		{contractOwner.Address(), toaddr, types.Error{ErrorCode: types.ErrNoAuthorization}},
	}
	//test SetOwner() of token-issue contract
	//测试用例： 调用代币发行合约转移所有者
	utest.AssertError(test.setSender(contractOwner).SetOwner(toaddr), types.CodeOK)

	tokencontract := test.obj.sdk.Helper().ContractHelper().ContractOfToken(addr)
	test.obj.sdk.Message().(*object.Message).SetContract(tokencontract)

	for i, t := range testcases {
		fmt.Println("case ", i)
		sender := test.obj.sdk.Helper().AccountHelper().AccountOf(t.from)
		utest.AssertError(test.setSender(sender).SetOwner(t.newOwner), t.err.ErrorCode)
	}

}
func (mysuit *MySuite) TestTokenIssue_SetGasPrice(c *C) {
	utest.Init(orgID)
	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, contractMethods)
	test := NewTestObject(contractOwner)

	addr, err := test.run().setSender(contractOwner).NewToken("mycoin", "mycn", bn.N(1e15), true, true, 5000)
	utest.AssertOK(err)
	utest.Assert(addr != "")
	utest.NextBlock(1)

	pubkey, _ := hex.DecodeString("FFE0014B0B08BB79B17B996ECABEDA6BF02534B64917631BB5DE59FB411B1083")
	toaddr := test.obj.sdk.Helper().AccountHelper().AccountOfPubKey(pubkey).Address()

	var testcases = []struct {
		from     types.Address
		gasprice int64
		err      types.Error
	}{
		//测试用例： 非所有者调用
		{toaddr, 5000, types.Error{ErrorCode: types.ErrNoAuthorization}},
		//测试用例： 所有者调用，燃料价格为负数
		{contractOwner.Address(), -10000, types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 所有者调用，燃料价格为负数
		{contractOwner.Address(), -1000, types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 所有者调用，燃料价格为0
		{contractOwner.Address(), 0, types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 所有者调用，燃料价格为1， 小于基础燃料价格
		{contractOwner.Address(), 1, types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 所有者调用，燃料价格为1000， 小于基础燃料价格
		{contractOwner.Address(), 1000, types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 正常用例
		{contractOwner.Address(), 2500, types.Error{ErrorCode: types.CodeOK}},
		//测试用例： 正常用例，值比较大
		{contractOwner.Address(), 1e8, types.Error{ErrorCode: types.CodeOK}},
		//测试用例： 所有者调用，燃料价格大于最大燃料价格限制
		{contractOwner.Address(), 1e9 + 1, types.Error{ErrorCode: types.ErrInvalidParameter}},
		//测试用例： 所有者调用，燃料价格大于最大燃料价格限制
		{contractOwner.Address(), 1e10, types.Error{ErrorCode: types.ErrInvalidParameter}},
	}
	//test SetGasPrice() of token-issue contract
	//测试用例： 调用代币发行合约设置燃料价格
	utest.AssertError(test.setSender(contractOwner).SetGasPrice(6000), types.ErrNoAuthorization)

	tokencontract := test.obj.sdk.Helper().ContractHelper().ContractOfToken(addr)
	test.obj.sdk.Message().(*object.Message).SetContract(tokencontract)

	for _, t := range testcases {
		sender := test.obj.sdk.Helper().AccountHelper().AccountOf(t.from)
		utest.AssertError(test.setSender(sender).SetGasPrice(t.gasprice), t.err.ErrorCode)
	}

}
