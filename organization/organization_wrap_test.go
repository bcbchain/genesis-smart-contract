package organization

import (
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/types"
	"github.com/bcbchain/sdk/sdkimpl/object"
	"github.com/bcbchain/sdk/sdkimpl/sdkhelper"
	"github.com/bcbchain/sdk/utest"
)

var (
	contractName    = "organizationmgr" //contract name
	contractMethods = []string{"RegisterOrganization(string)TYPE"}
	orgID           = "orgNUjCm1i8RcoW2kVTbDw4vKW6jzfMxewJHjkhuiduhjuikjuyhnnjkuhujk111"
)

//TestObject This is a struct for test
type TestObject struct {
	obj *Organization
}

//FuncRecover recover panic by Assert
func FuncRecover(err *types.Error) {
	if rerr := recover(); rerr != nil {
		if _, ok := rerr.(types.Error); ok {
			err.ErrorCode = rerr.(types.Error).ErrorCode
			err.ErrorDesc = rerr.(types.Error).ErrorDesc
		} else {
			panic(rerr)
		}
	}
}

//NewTestObject This is a function
func NewTestObject(sender sdk.IAccount) *TestObject {
	return &TestObject{&Organization{sdk: utest.UTP.ISmartContract}}
}

//transfer This is a method of TestObject
func (t *TestObject) transfer(balance bn.Number) *TestObject {
	t.obj.sdk.Message().Sender().TransferByName(t.obj.sdk.Helper().GenesisHelper().Token().Name(), t.obj.sdk.Message().Contract().Account().Address(), balance)
	t.obj.sdk = sdkhelper.OriginNewMessage(t.obj.sdk, t.obj.sdk.Message().Contract(), t.obj.sdk.Message().MethodID(), t.obj.sdk.Message().(*object.Message).OutputReceipts())
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

//RegisterOrganization This is a method of TestObject
func (t *TestObject) RegisterOrganization(name string) (result0 string, err types.Error) {
	err.ErrorCode = types.CodeOK
	defer FuncRecover(&err)
	utest.NextBlock(1)
	result0 = t.obj.RegisterOrganization(name)
	utest.Commit()
	return
}

//SetSigners This is a method of TestObject
func (t *TestObject) SetSigners(orgID string, pubKeys []types.PubKey) (err types.Error) {
	err.ErrorCode = types.CodeOK
	defer FuncRecover(&err)
	utest.NextBlock(1)
	t.obj.SetSigners(orgID, pubKeys)
	utest.Commit()
	return
}
