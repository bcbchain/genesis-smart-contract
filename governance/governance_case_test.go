package governance

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/bcbchain/sdk/sdk/types"
	"github.com/bcbchain/sdk/utest"

	"gopkg.in/check.v1"
)

//Test is a function
//Test is a function
func Test(t *testing.T) { check.TestingT(t) }

//MySuite is a struct
type MySuite struct{}

var _ = check.Suite(&MySuite{})

const (
	pubKeyA = "A1EDF8F50848B8FA121A24E2A3A83CC5C8CBF85D6CE23A3A8413F46A717BEDA1"
	pubKeyB = "B1EDF8F50848B8FA121A24E2A3A83CC5C8CBF85D6CE23A3A8413F46A717BEDA1"
	pubKeyC = "C1EDF8F50848B8FA121A24E2A3A83CC5C8CBF85D6CE23A3A8413F46A717BEDA1"
	pubKeyD = "D1EDF8F50848B8FA121A24E2A3A83CC5C8CBF85D6CE23A3A8413F46A717BEDA1"
	pubKeyE = "E1EDF8F50848B8FA121A24E2A3A83CC5C8CBF85D6CE23A3A8413F46A717BEDA1"
)

var (
	nodeA        string
	nodeB        string
	nodeC        string
	nodeD        string
	pubKeyABytes []byte
	pubKeyBBytes []byte
	pubKeyCBytes []byte
	pubKeyDBytes []byte
	pubKeyEBytes []byte
)

//TestValidatorMgr_NewValidator is a method of MySuite
func (mysuit *MySuite) TestValidatorMgr_NewValidator(c *check.C) {
	utest.Init(orgID)
	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, nil)
	test := NewTestObject(contractOwner)
	setNodes(test)

	var testCases = []struct {
		name       string
		pubKey     types.PubKey
		rewardAddr types.Address
		power      int64
		err        types.Error
	}{
		// 1.正常用例
		{"test", pubKeyEBytes,
			test.obj.sdk.Helper().AccountHelper().AccountOfPubKey(pubKeyABytes).Address(),
			10, types.Error{ErrorCode: types.CodeOK}},
		// 2.名字重复
		{"test", []byte("testTestTestTestTestTestTestTest"),
			test.obj.sdk.Helper().AccountHelper().AccountOfPubKey(pubKeyABytes).Address(),
			10, types.Error{ErrorCode: types.ErrInvalidParameter}},
		// 3.power 超过总数的 1/3
		{"test1", []byte("ATestTestTestTestTestTestTestTes"),
			test.obj.sdk.Helper().AccountHelper().AccountOfPubKey([]byte("ATestNewValidatorTestNewValidato")).Address(),
			50, types.Error{ErrorCode: types.ErrInvalidParameter}},
		// 4.pubKey 重复
		{"test2", pubKeyEBytes,
			test.obj.sdk.Helper().AccountHelper().AccountOfPubKey([]byte("ATestNewValidatorTestNewValidato")).Address(),
			10, types.Error{ErrorCode: types.ErrInvalidParameter}},
		// 5.pubKey 为空
		{"test3", []byte(""),
			test.obj.sdk.Helper().AccountHelper().AccountOfPubKey([]byte("ATestNewValidatorTestNewValidato")).Address(),
			10, types.Error{ErrorCode: types.ErrInvalidParameter}},
		// 6.名字，公钥，奖励地址均为空
		{"", []byte(""),
			"", 10, types.Error{ErrorCode: types.ErrInvalidParameter}},
		// 7.奖励地址为空
		{"test4", []byte("testNewValidatorTestNewValidatoC"),
			"", 10, types.Error{ErrorCode: types.ErrInvalidAddress}},
		// 8.奖励地址无效
		{"test5", []byte("testNewValidatorTestNewValidatoC"),
			"a", 10, types.Error{ErrorCode: types.ErrInvalidAddress}},
		// 9.power 为0，正常用例
		{"test5", []byte("testNewValidatorTestNewValidatoC"),
			test.obj.sdk.Helper().AccountHelper().AccountOfPubKey([]byte("ATestNewValidatorTestNewValidato")).Address(),
			0, types.Error{ErrorCode: types.ErrInvalidParameter}},
	}

	for i, c := range testCases {
		err := test.run().setSender(contractOwner).NewValidator(c.name, c.pubKey, c.rewardAddr, c.power)
		fmt.Println("Index:", i+1, "ErrorDesc: "+err.ErrorDesc)
		utest.AssertError(err, c.err.ErrorCode)

		if c.err.ErrorCode == types.CodeOK {
			utest.AssertSDB(test.obj.sdk.Message().Contract().KeyPrefix()+"/validator/"+test.obj.sdk.Helper().BlockChainHelper().CalcAccountFromPubKey(c.pubKey),
				InfoOfValidator{
					PubKey:     c.pubKey,
					Name:       c.name,
					NodeAddr:   test.obj.sdk.Helper().BlockChainHelper().CalcAccountFromPubKey(c.pubKey),
					RewardAddr: c.rewardAddr,
					Power:      c.power,
				})
			found := false
			res := test.obj.sdk.Helper().StateHelper().GetEx("/validators/all/0", &[][]byte{}).(*[][]byte)
			for _, v := range *res {
				if hex.EncodeToString(v) == hex.EncodeToString(c.pubKey) {
					found = true
					break
				}
			}
			utest.Assert(found)
		}
	}
}

//TestValidatorMgr_SetPower is a method of MySuite
func (mysuit *MySuite) TestValidatorMgr_SetPower(c *check.C) {
	utest.Init(orgID)
	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, nil)
	test := NewTestObject(contractOwner)
	setNodes(test)

	var testCases = []struct {
		pubKey types.PubKey
		power  int64
		err    types.Error
	}{
		// 1.正常用例
		{pubKeyBBytes, 16, types.Error{ErrorCode: types.CodeOK}},
		// 2.power 超过总数的 1/3
		{[]byte("ETestNewValidatorTestNewValidato"), 200, types.Error{ErrorCode: types.ErrInvalidParameter}},
		// 3.pubKey 为空
		{[]byte(""), 10, types.Error{ErrorCode: types.ErrInvalidParameter}},
		// 4.最大 power 超过总数的 1/3
		{pubKeyBBytes, 2, types.Error{ErrorCode: types.ErrInvalidParameter}},
		// 5.无效的 pubKey
		{[]byte("ETestNewValidatorTestNew"), 11, types.Error{ErrorCode: types.ErrInvalidParameter}},
	}

	for i, c := range testCases {
		err := test.run().setSender(contractOwner).SetPower(c.pubKey, c.power)
		fmt.Println("Index:", i+1, "ErrorDesc: "+err.ErrorDesc)
		utest.AssertError(err, c.err.ErrorCode)

		if err.ErrorCode == types.CodeOK {
			r := *test.obj.sdk.Helper().StateHelper().Get("/validator/"+nodeB, &InfoOfValidator{}).(*InfoOfValidator)
			utest.Assert(r.Power == c.power)
		}
	}
}

//TestValidatorMgr_SetRewardAddr is a method of MySuite
func (mysuit *MySuite) TestValidatorMgr_SetRewardAddr(c *check.C) {
	utest.Init(orgID)
	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, nil)
	test := NewTestObject(contractOwner)
	setNodes(test)

	var testCases = []struct {
		pubKey     types.PubKey
		rewardAddr types.Address
		err        types.Error
	}{
		// 1.正常用例
		{pubKeyBBytes, nodeC, types.Error{ErrorCode: types.CodeOK}},
		// 2.无效的 pubKey
		{[]byte("ETestNewValidatorTestNewValidat"), nodeB, types.Error{ErrorCode: types.ErrInvalidParameter}},
		// 3.无效的奖励地址
		{pubKeyCBytes, nodeB + "test", types.Error{ErrorCode: types.ErrInvalidAddress}},
		// 4.奖励地址为空
		{pubKeyCBytes, "", types.Error{ErrorCode: types.ErrInvalidAddress}},
		// 5.pubKey 为空
		{[]byte(""), nodeD, types.Error{ErrorCode: types.ErrInvalidParameter}},
	}

	for i, c := range testCases {
		err := test.run().setSender(contractOwner).SetRewardAddr(c.pubKey, c.rewardAddr)
		fmt.Println("Index:", i+1, "ErrorDesc: "+err.ErrorDesc)
		utest.AssertError(err, c.err.ErrorCode)

		if err.ErrorCode == types.CodeOK {
			r := *test.obj.sdk.Helper().StateHelper().Get("/validator/"+test.obj.sdk.Helper().AccountHelper().AccountOfPubKey(c.pubKey).Address(), &InfoOfValidator{}).(*InfoOfValidator)
			utest.Assert(r.RewardAddr == c.rewardAddr)
		}
	}
}

//TestValidatorMgr_SetRewardStrategy is a method of MySuite
func (mysuit *MySuite) TestValidatorMgr_SetRewardStrategy(c *check.C) {
	utest.Init(orgID)
	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, nil)
	test := NewTestObject(contractOwner)
	currentBlockHeight := test.obj.sdk.Block().Height()
	oldStrategy := []RewardStrategy{
		{Strategy: []Reward{
			{Name: "validators", RewardPercent: "20.00", Address: "a1"},
			{Name: "validators", RewardPercent: "30.00", Address: "b"},
			{Name: "validators", RewardPercent: "22.00", Address: "c"},
			{Name: "validators", RewardPercent: "28.00", Address: "d"}}, EffectHeight: currentBlockHeight},
		{Strategy: []Reward{
			{Name: "validators", RewardPercent: "20.00", Address: "a2"},
			{Name: "validators", RewardPercent: "30.00", Address: "b"},
			{Name: "validators", RewardPercent: "22.00", Address: "c"},
			{Name: "validators", RewardPercent: "28.00", Address: "d"}}, EffectHeight: currentBlockHeight + 4},
		{Strategy: []Reward{
			{Name: "validators", RewardPercent: "20.00", Address: "a3"},
			{Name: "validators", RewardPercent: "30.00", Address: "b"},
			{Name: "validators", RewardPercent: "22.00", Address: "c"},
			{Name: "validators", RewardPercent: "28.00", Address: "d"}}, EffectHeight: currentBlockHeight - 4},
	}

	test.obj.sdk.Helper().StateHelper().Set("/rewardstrategys", oldStrategy)
	setNodes(test)

	strategyStr := `{"rewardStrategy":[{"name":"validators","rewardPercent":"50.50","address":"` + nodeD + `"},` +
		`{"name":"noValidators","rewardPercent":"49.50","address":"` + nodeC + `"}],"effectHeight":100}`
	var testCases = []struct {
		strategy     string
		effectHeight int64
		err          types.Error
	}{
		// 1.正常用例
		{strategyStr, test.obj.sdk.Block().Height() + 8, types.Error{ErrorCode: types.CodeOK}},
		// 2.生效高度小于存在的生效高度
		{strategyStr, test.obj.sdk.Block().Height() + 2, types.Error{ErrorCode: types.ErrInvalidParameter}},
		// 3.生效高度为 0
		{strategyStr, 0, types.Error{ErrorCode: types.ErrInvalidParameter}},
		// 4.奖励策略空
		{"", test.obj.sdk.Block().Height() + 9, types.Error{ErrorCode: types.ErrInvalidParameter}},
		// 5.无效的奖励策略
		{"testInvalidStrategy", test.obj.sdk.Block().Height() + 9, types.Error{ErrorCode: types.ErrInvalidParameter}},
	}

	for i, c := range testCases {
		err := test.run().setSender(contractOwner).SetRewardStrategy(c.strategy)
		fmt.Println("Index:", i+1, "ErrorDesc: "+err.ErrorDesc)
		utest.AssertError(err, c.err.ErrorCode)

		if err.ErrorCode == types.CodeOK {
			newStrategy := RewardStrategy{}
			json.Unmarshal([]byte(strategyStr), &newStrategy)

			a := *test.obj.sdk.Helper().StateHelper().GetEx("/rewardstrategys", new([]RewardStrategy)).(*[]RewardStrategy)
			fmt.Println("a:", a)
			r := false
			for _, v := range a {
				if v.EffectHeight == 100 {
					if v.Strategy[0].Name == "validators" {
						r = true
					}
				}
			}
			utest.Assert(r)
		}
	}
}

func setNodes(test *TestObject) {
	pubKeyABytes, _ = hex.DecodeString(pubKeyA)
	pubKeyBBytes, _ = hex.DecodeString(pubKeyB)
	pubKeyCBytes, _ = hex.DecodeString(pubKeyC)
	pubKeyDBytes, _ = hex.DecodeString(pubKeyD)
	pubKeyEBytes, _ = hex.DecodeString(pubKeyE)
	test.obj.sdk.Helper().StateHelper().Set("/validators/all/0", [][]byte{pubKeyABytes, pubKeyBBytes, pubKeyCBytes, pubKeyDBytes})

	nodeA = test.obj.sdk.Helper().AccountHelper().AccountOfPubKey(pubKeyABytes).Address()
	test.obj.sdk.Helper().StateHelper().Set("/validator/"+nodeA, InfoOfValidator{Name: "test-node-A",
		NodeAddr: nodeA, Power: 10})

	nodeB = test.obj.sdk.Helper().AccountHelper().AccountOfPubKey(pubKeyBBytes).Address()
	test.obj.sdk.Helper().StateHelper().Set("/validator/"+nodeB, InfoOfValidator{Name: "test-node-B",
		NodeAddr: nodeB, Power: 15})

	nodeC = test.obj.sdk.Helper().AccountHelper().AccountOfPubKey(pubKeyCBytes).Address()
	test.obj.sdk.Helper().StateHelper().Set("/validator/"+nodeC, InfoOfValidator{Name: "test-node-C",
		NodeAddr: nodeC, Power: 15})

	nodeD = test.obj.sdk.Helper().AccountHelper().AccountOfPubKey(pubKeyDBytes).Address()
	test.obj.sdk.Helper().StateHelper().Set("/validator/"+nodeD, InfoOfValidator{Name: "test-node-D",
		NodeAddr: nodeD, Power: 10})
}
