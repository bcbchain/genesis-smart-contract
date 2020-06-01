package mining

import (
	"fmt"
	"github.com/bcbchain/sdk/common/gls"
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/types"
	"github.com/bcbchain/sdk/utest"
	"gopkg.in/check.v1"
	"math"
	"testing"
)

const nextHeight = 100000

//Test This is a function
func Test(t *testing.T) { check.TestingT(t) }

//MySuite This is a struct
type MySuite struct{}

var _ = check.Suite(&MySuite{})

//TestMining_Mine is a method of MySuite
func (mysuit *MySuite) TestMining_Mine(c *check.C) {
	utest.Init(orgID)

	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, contractInterfaces)
	gls.Mgr.SetValues(gls.Values{gls.SDKKey: utest.UTP.ISmartContract}, func() {
		test := NewTestObject(contractOwner)
		//test.setSender(contractOwner).InitChain()

		//给合约账户转钱
		contract := utest.UTP.Message().Contract()
		genesisOwner := utest.UTP.Helper().GenesisHelper().Token().Owner()
		utest.Assert(test.run().setSender(utest.UTP.Helper().AccountHelper().AccountOf(genesisOwner.Address())) != nil)

		utest.Transfer(nil, contract.Account().Address(), bn.N(1.98*1e16))

		fmt.Println("genesisBalance", utest.UTP.Helper().GenesisHelper().Token().Owner().BalanceOfToken(utest.GetGenesisToken().Address))
		fmt.Println("contractBalance", test.obj.sdk.Helper().AccountHelper().AccountOf(contract.Account().Address()).BalanceOfName("TSC"))
		//获取下奖励地址金额
		fmt.Println("=== Run UnitTestcase: Mine() mine")

		testCases := []struct {
			flag bool //初始化InitChain标志位
			err  types.Error
			desc string
		}{
			{true, types.Error{ErrorCode: types.CodeOK, ErrorDesc: ""}, "测试用例"},
		}

		for _, v := range testCases {
			if v.flag == true {
				test.setSender(contractOwner).InitChain()
			}

			indexHeight := test.obj.sdk.Block().Height()
			for {
				indexHeight = indexHeight + nextHeight
				rewardAmount := calcRewardBal(indexHeight, test.run().setSender(contractOwner).obj._miningStartHeight())

				beforMineBal := test.obj.sdk.Helper().AccountHelper().AccountOf(test.obj.sdk.Block().RewardAddress()).BalanceOfName("TSC")
				a := test.run().setSender(contractOwner)
				err := a.Mine()
				utest.AssertError(err, v.err.ErrorCode)
				if err.ErrorCode == types.CodeOK {
					afterMineBal := test.obj.sdk.Helper().AccountHelper().AccountOf(test.obj.sdk.Block().RewardAddress()).BalanceOfName("TSC")
					rewardBal := afterMineBal.Sub(beforMineBal)

					//fmt.Println("rewardBal",rewardBal)
					if !(rewardAmount == 0) {
						utest.Assert(rewardBal.IsEqualI(rewardAmount))
					} else {
						utest.Assert(rewardBal.IsEqualI(int64(1)))
						break
					}
				} else {
					utest.AssertErrorMsg(err, v.err.ErrorDesc)
					break
				}
			}
		}
		fmt.Println("--- Pass UnitTestcase: Mine() mine")
	})
}

func calcRewardBal(cHeight, sHeight int64) (rewardAmount int64) {
	blockNum := cHeight - sHeight
	if blockNum == 0 {
		//给奖励地址150000000cong
		rewardAmount = int64(150000000)
	} else {
		rewardAmount = int64(150000000 / int64(math.Pow(2, float64(blockNum/66000000))))
	}
	return
}
