package netgovernance

import (
	"fmt"
	"github.com/bcbchain/sdk/common/gls"
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/types"
	"github.com/bcbchain/sdk/utest"
	"gopkg.in/check.v1"
	"math/rand"
)

//RegisterSideChain This is a method of MySuite
func (mysuit *MySuite) TestDemo_RegisterSideChain(c *check.C) {
	fmt.Println("RegisterSideChain Test!")
	utest.Init(orgID)
	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, contractInterfaces)

	gls.Mgr.SetValues(gls.Values{gls.SDKKey: utest.UTP.ISmartContract}, func() {
		test := NewTestObject(contractOwner)
		test.setSender(contractOwner).InitChain()
		mysuit.test_RegisterSideChain(contractOwner, test)
	})
}

func GetRandomName(slen int) string {
	charaterStr := "qwertyuiop[]\asdfghjkl;'zxcvbnm,./QWERTYUIOP{}|ASDFGHJKL:ZXCVBNM<>?"

	name := ""
	crs := []rune(charaterStr)
	for i := 0; i < slen; i++ {
		r := rand.Intn(len(charaterStr) - 1)
		name = name + string(crs[r:r+1])
	}

	randomName := name

	return randomName
}

func (mysuit *MySuite) test_RegisterSideChain(owner sdk.IAccount, test *TestObject) {
	//TODO
	test.setSender(owner).InitChain()

	// create accounts
	accounts := utest.NewAccounts(test.obj.sdk.Helper().GenesisHelper().Token().Name(),
		bn.N(1e9), 2)

	// register Org
	utest.RegisterOrg(accounts[0], "test")
	utest.RegisterOrg(accounts[1], "test2")

	var tests = []struct {
		account     sdk.IAccount
		sideChainID string
		orgName     string
		ownerPubKey types.PubKey
		desc        string
		code        uint32
	}{
		//权限
		{accounts[0], "BCBChain", "test", accounts[0].PubKey(), "非主链委员会调用", types.ErrNoAuthorization},
		{owner, "BCBChain", "test", accounts[0].PubKey(), "主链委员会调用", types.CodeOK},
		//侧链ID
		{owner, "BCBChain1", "test", accounts[0].PubKey(), "侧链ID有数字", types.CodeOK},
		{owner, "bcbChain", "test", accounts[0].PubKey(), "侧链ID由大小写组成", types.CodeOK},
		{owner, "BCBChain!@#$%^&*()_+=-~`][;:',.<>/?", "test", accounts[0].PubKey(), "侧链ID由特殊字符组成", types.CodeOK},
		{owner, GetRandomName(1000), "test", accounts[0].PubKey(), "侧链ID长度为1000", types.CodeOK},
		{owner, "ΑΒΓΔΕΖΗΘΙΚ∧ΜΝΞΟ∏Ρ∑ΤΥΦΧΨΩ", "test", accounts[0].PubKey(), "侧链ID为希腊字母", types.CodeOK},
		{owner, "αβγδεζηθικλμνξοπρστυφχψω", "test", accounts[0].PubKey(), "侧链ID为希腊字母", types.CodeOK},
		{owner, "АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ", "test", accounts[0].PubKey(), "侧链ID为俄文", types.CodeOK},
		{owner, "абвгдеёжзийклмнопрстуфхцчшщъыьэюя", "test", accounts[0].PubKey(), "侧链ID为俄文", types.CodeOK},
		{owner, "āáǎàōóǒòêēéěèīíǐìūúǔùǖǘǚǜü", "test", accounts[0].PubKey(), "侧链ID为拼音", types.CodeOK},
		{owner, "ぁぃぅぇぉかきくけこんさしすせそたちつってとゐなにぬねのはひふへほゑまみむめもゃゅょゎを", "test", accounts[0].PubKey(), "侧链ID为日文", types.CodeOK},
		{owner, "ァィゥヴェォカヵキクケヶコサシスセソタチツッテトヰンナニヌネノハヒフヘホヱマミムメモャュョヮヲ", "test", accounts[0].PubKey(), "侧链ID为日文", types.CodeOK},
		{owner, "侧链测试", "test", accounts[0].PubKey(), "侧链ID为中文", types.CodeOK},
		{owner, "Où sont les toilettes ?", "test", accounts[0].PubKey(), "侧链ID为法文", types.CodeOK},
		{owner, GetRandomName(1), "test", accounts[0].PubKey(), "侧链ID长度为1", types.ErrInvalidParameter},
		{owner, "", "test", accounts[0].PubKey(), "侧链ID为空", types.ErrInvalidParameter},
		{owner, "BCBChain", "test", accounts[0].PubKey(), "同组织已使用过该侧链ID", types.ErrInvalidParameter},
		{owner, "BCBChain", "test2", accounts[1].PubKey(), "不同组织已使用过该侧链ID", types.ErrInvalidParameter},

		//orgName
		{owner, "test0", "test0", accounts[1].PubKey(), "对应组织未注册", types.ErrInvalidParameter},
		{owner, "orgNameIsNil", "", accounts[1].PubKey(), "组织为空", types.ErrInvalidParameter},
		{owner, "genesisOrgIDTest", "genesis", accounts[1].PubKey(), "组织为基础组织", types.ErrInvalidParameter},

		//ownerPubKey
		{owner, "ownerPubKeyIsNil", "test", []byte(""), "PubKey为空", types.ErrInvalidParameter},
		{owner, "ownerPubKeyError", "test", []byte("test"), "PubKey错误", types.ErrInvalidParameter},
		//当前测试代码，链Owner每次重新启动后都是随机生成的一个账户，而本币的owner是一个定值，导致合约端查到的链Owner与本币Owner不一致，所以该测试能通过。
		{owner, "ownerIsGenesisOwner", "test", owner.PubKey(), "PubKey为合约Owner", types.ErrInvalidParameter},
		//todo 在侧链上调用该接口
	}

	for _, item := range tests {
		test.run(item.code, func(t *TestObject) types.Error {
			err := test.setSender(item.account).RegisterSideChain(item.sideChainID,
				item.orgName, item.ownerPubKey)
			if err.ErrorCode == types.CodeOK {
				fmt.Println(item.desc)
			}
			return err
		})
	}

}
