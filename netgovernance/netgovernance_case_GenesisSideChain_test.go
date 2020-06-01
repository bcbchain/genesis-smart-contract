package netgovernance

import (
	"fmt"
	"github.com/bcbchain/sdk/common/gls"
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/types"
	"github.com/bcbchain/sdk/utest"
	"gopkg.in/check.v1"
)

const DC = "Diamond Coin"

//GenesisSideChain This is a method of MySuite
func (mysuit *MySuite) TestDemo_GenesisSideChain(c *check.C) {
	fmt.Println("GenesisSideChain Test!")
	utest.Init(orgID)
	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, contractInterfaces)

	gls.Mgr.SetValues(gls.Values{gls.SDKKey: utest.UTP.ISmartContract}, func() {
		test := NewTestObject(contractOwner)
		test.setSender(contractOwner).InitChain()
		mysuit.test_GenesisSideChain(contractOwner, test)
	})
	fmt.Println("")
}

func (mysuit *MySuite) test_GenesisSideChain(owner sdk.IAccount, test *TestObject) {
	//TODO
	//
	test.setSender(owner).InitChain()
	genesisToken := test.obj.sdk.Helper().GenesisHelper().Token().Name()
	accounts := utest.NewAccounts(genesisToken, bn.N(1e11), 3)

	//transfer token
	utest.Transfer(nil, owner.Address(), genesisToken, bn.N(1e11))
	utest.Transfer(nil, accounts[0].Address(), DC, bn.N(1e15))

	//注册组织A,注册侧链A
	orgA := "orgA"
	utest.RegisterOrg(accounts[0], orgA)
	chainA := "chainA"
	test.run(types.CodeOK, func(t *TestObject) types.Error {
		return t.setSender(owner).RegisterSideChain(chainA, orgA, accounts[0].PubKey())
	})

	//组织A,注册侧链B
	chainB := "chainB"
	test.run(types.CodeOK, func(t *TestObject) types.Error {
		return t.setSender(owner).RegisterSideChain(chainB, orgA, accounts[0].PubKey())
	})

	//组织A,注册侧链C
	chainC := "chainC"
	test.run(types.CodeOK, func(t *TestObject) types.Error {
		return t.setSender(owner).RegisterSideChain(chainC, orgA, accounts[0].PubKey())
	})

	var tests = []struct {
		account     sdk.IAccount
		tokenName   string
		sideChainID string
		nodeName    string
		nodePubKey  types.PubKey
		initReward  bn.Number
		rewardAddr  types.Address
		openURL     string
		desc        string
		code        uint32
	}{
		//场景
		{accounts[1], test.obj.sdk.Helper().GenesisHelper().Token().Name(), chainA, "nodeA", owner.PubKey(), bn.N(1e9), accounts[1].Address(), "11", "非侧链委员会调用", types.ErrNoAuthorization},
		{owner, test.obj.sdk.Helper().GenesisHelper().Token().Name(), chainA, "nodeA", owner.PubKey(), bn.N(1e9), accounts[1].Address(), "11", "主链委员会调用", types.ErrNoAuthorization},
		{accounts[0], DC, chainA, "nodeA", owner.PubKey(), bn.N(1e9), accounts[1].Address(), "11", "转账币种不为 BCB", types.ErrInvalidParameter},

		//sideChainID
		{accounts[0], test.obj.sdk.Helper().GenesisHelper().Token().Name(), "", "nodeA", owner.PubKey(), bn.N(1e9), accounts[1].Address(), "11", "sideChainID 为空", types.ErrInvalidParameter},
		{accounts[0], test.obj.sdk.Helper().GenesisHelper().Token().Name(), "NotExit", "nodeA", owner.PubKey(), bn.N(1e9), accounts[1].Address(), "11", "sideChainID 未注册", types.ErrInvalidParameter},
		{accounts[0], test.obj.sdk.Helper().GenesisHelper().Token().Name(), chainA, "nodeA", owner.PubKey(), bn.N(1e9), accounts[1].Address(), "11", "正常流程", types.CodeOK},
		{accounts[0], test.obj.sdk.Helper().GenesisHelper().Token().Name(), chainA, "nodeA", accounts[1].PubKey(), bn.N(1e9), accounts[1].Address(), "11", "sideChainID 已被创世", types.ErrInvalidParameter},

		//nodeName
		{accounts[0], test.obj.sdk.Helper().GenesisHelper().Token().Name(), chainB, "", accounts[1].PubKey(), bn.N(1e9), accounts[1].Address(), "11", "nodeName 为空", types.ErrInvalidParameter},
		{accounts[0], test.obj.sdk.Helper().GenesisHelper().Token().Name(), chainB, GetRandomName(41), accounts[1].PubKey(), bn.N(1e9), accounts[1].Address(), "11", "nodeName 为空", types.ErrInvalidParameter},
		{accounts[0], test.obj.sdk.Helper().GenesisHelper().Token().Name(), chainB, "nodeA", accounts[1].PubKey(), bn.N(1e9), accounts[1].Address(), "11", "nodeName 已被使用--成功", types.CodeOK},

		//nodePubKey
		{accounts[0], test.obj.sdk.Helper().GenesisHelper().Token().Name(), chainB, "chainB", accounts[1].PubKey()[1:], bn.N(1e9), accounts[1].Address(), "11", "nodePubKey 不满足32字节", types.ErrInvalidParameter},
		//todo 为主链的验证者节点
		{accounts[0], test.obj.sdk.Helper().GenesisHelper().Token().Name(), chainC, "chainC", owner.PubKey(), bn.N(1e9), accounts[1].Address(), "11", "nodePubKey 已被使用", types.ErrInvalidParameter},

		//initReward
		{accounts[0], test.obj.sdk.Helper().GenesisHelper().Token().Name(), chainC, "chainC", accounts[2].PubKey(), bn.N(1e11), accounts[1].Address(), "11", "initReward > 转入代币量", types.ErrInvalidParameter},
		{accounts[0], test.obj.sdk.Helper().GenesisHelper().Token().Name(), chainC, "chainC", accounts[2].PubKey(), bn.N(-1), accounts[1].Address(), "11", "initReward > 转入代币量", types.ErrInvalidParameter},

		//rewardAddr
		//{accounts[0],test.obj.sdk.Helper().GenesisHelper().Token().Name(),chainC,"chainC",accounts[2].PubKey(),bn.N(1),accounts[1].Address()[3:],"11","rewardAddr 格式错误",types.ErrInvalidParameter},
		//{accounts[0],test.obj.sdk.Helper().GenesisHelper().Token().Name(),chainC,"chainC",accounts[2].PubKey(),bn.N(1),"","11","rewardAddr 为空",types.ErrInvalidParameter},
		//todo 为主链上的地址
		//todo 为其他侧链上的地址

		//openURL
		{accounts[0], test.obj.sdk.Helper().GenesisHelper().Token().Name(), chainC, "chainC", accounts[2].PubKey(), bn.N(-1), accounts[2].Address(), "", "openURL 为空", types.ErrInvalidParameter},
	}

	for _, item := range tests {

		test.run(item.code, func(t *TestObject) types.Error {
			t.setSender(item.account)
			utest.Assert(t.transfer(item.tokenName, bn.N(1e9)) != nil)

			err := t.GenesisSideChain(item.sideChainID,
				item.nodeName, item.nodePubKey, item.initReward, item.rewardAddr, item.openURL)
			if err.ErrorCode == types.CodeOK {
				fmt.Println(item.desc)
			}

			return err
		})

	}

}
