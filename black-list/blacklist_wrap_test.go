package blacklist

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
	contractName       = "black-list" //contract name
	contractMethods    = []string{"AddAddress([]types.Address)", "DelAddress([]types.Address)", "SetOwner(types.Address)"}
	contractInterfaces = []string{}
	orgID              = "orgJgaGConUyK81zibntUBjQ33PKctpk1K1G"
)

//TestObject This is a struct for test
type TestObject struct {
	obj *Blacklist
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
	return &TestObject{&Blacklist{sdk: utest.UTP.ISmartContract}}
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

//AddAddress This is a method of TestObject
func (t *TestObject) AddAddress(blacklist []types.Address) (err types.Error) {
	err.ErrorCode = types.CodeOK
	defer FuncRecover(&err)
	utest.NextBlock(1)
	t.obj.AddAddress(blacklist)
	utest.Commit()
	return
}

//DelAddress This is a method of TestObject
func (t *TestObject) DelAddress(blacklist []types.Address) (err types.Error) {
	err.ErrorCode = types.CodeOK
	defer FuncRecover(&err)
	utest.NextBlock(1)
	t.obj.DelAddress(blacklist)
	utest.Commit()
	return
}

//SetOwner This is a method of TestObject
func (t *TestObject) SetOwner(newOwnerAddr types.Address) (err types.Error) {
	err.ErrorCode = types.CodeOK
	defer FuncRecover(&err)
	utest.NextBlock(1)
	t.obj.SetOwner(newOwnerAddr)
	utest.Commit()
	return
}
