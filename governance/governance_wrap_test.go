package governance

import (
	"fmt"
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/types"
	"github.com/bcbchain/sdk/sdkimpl/object"
	"github.com/bcbchain/sdk/sdkimpl/sdkhelper"
	"github.com/bcbchain/sdk/utest"
)

var (
	contractName       = "validator" //contract name
	contractMethods    = []string{"NewValidator(string,types.PubKey,types.Address,int64)", "SetPower(types.PubKey,int64)", "SetRewardAddr(types.PubKey,types.Address)", "SetRewardStrategy(string)"}
	contractInterfaces = []string{}
	orgID              = "orgJgaGConUyK81zibntUBjQ33PKctpk1K1G"
)

//TestObject This is a struct for test
type TestObject struct {
	obj *Governance
}

//FuncRecover recover panic by Assert
func FuncRecover(err *types.Error) {
	if rerr := recover(); rerr != nil {
		if _, ok := rerr.(types.Error); ok {
			err.ErrorCode = rerr.(types.Error).ErrorCode
			err.ErrorDesc = rerr.(types.Error).ErrorDesc
			fmt.Println(err)
		} else {
			panic(rerr)
		}
	}
}

//NewTestObject This is a function
func NewTestObject(sender sdk.IAccount) *TestObject {
	return &TestObject{&Governance{sdk: utest.UTP.ISmartContract}}
}

//transfer This is a method of TestObject
func (t *TestObject) transfer(balance bn.Number) *TestObject {
	contract := t.obj.sdk.Message().Contract()
	utest.Transfer(t.obj.sdk.Message().Sender(), t.obj.sdk.Helper().GenesisHelper().Token().Name(), contract.Account().Address(), balance)
	t.obj.sdk = sdkhelper.OriginNewMessage(t.obj.sdk, contract, t.obj.sdk.Message().MethodID(), t.obj.sdk.Message().(*object.Message).OutputReceipts())
	return t
}

//setSender This is a method of TestObject
func (t *TestObject) setSender(sender sdk.IAccount) *TestObject {
	t.obj.sdk = utest.SetSender(sender.Address())
	return t
}

//run This is a method of TestObject
func (t *TestObject) run() *TestObject {
	t.obj.sdk = utest.ResetMsg()
	return t
}

//InitChain This is a method of TestObject
func (t *TestObject) InitChain() {
	utest.NextBlock(1)
	t.obj.InitChain()
	utest.Commit()
	return
}

//NewValidator This is a method of TestObject
func (t *TestObject) NewValidator(
	name string,
	pubKey types.PubKey,
	rewardAddr types.Address,
	power int64) (err types.Error) {
	err.ErrorCode = types.CodeOK
	defer FuncRecover(&err)
	utest.NextBlock(1)
	t.obj.NewValidator(name, pubKey, rewardAddr, power)
	utest.Commit()
	return
}

//SetPower This is a method of TestObject
func (t *TestObject) SetPower(pubKey types.PubKey, power int64) (err types.Error) {
	err.ErrorCode = types.CodeOK
	defer FuncRecover(&err)
	utest.NextBlock(1)
	t.obj.SetPower(pubKey, power)
	utest.Commit()
	return
}

//SetRewardAddr This is a method of TestObject
func (t *TestObject) SetRewardAddr(pubKey types.PubKey, rewardAddr types.Address) (err types.Error) {
	err.ErrorCode = types.CodeOK
	defer FuncRecover(&err)
	utest.NextBlock(1)
	t.obj.SetRewardAddr(pubKey, rewardAddr)
	utest.Commit()
	return
}

//SetRewardStrategy This is a method of TestObject
func (t *TestObject) SetRewardStrategy(strategy string) (err types.Error) {
	err.ErrorCode = types.CodeOK
	defer FuncRecover(&err)
	utest.NextBlock(1)
	t.obj.SetRewardStrategy(strategy)
	utest.Commit()
	return
}
