package smartcontract

import (
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/types"
	"github.com/bcbchain/sdk/utest"
)

var (
	contractName       = "smartcontract" //contract name
	contractMethods    = []string{"RegisterOrganization()TYPE", "DeployContract(string,string,string,types.Hash,[]byte,string,string,int64)TYPE", "ForbidInternalContract(types.Address,int64)"}
	contractInterfaces = []string{}
	orgID              = "orgJgaGConUyK81zibntUBjQ33PKctpk1K1G"
)

//TestObject This is a struct for test
type TestObject struct {
	obj *SmartContract
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
	return &TestObject{&SmartContract{sdk: utest.UTP.ISmartContract}}
}

//transfer This is a method of TestObject
func (t *TestObject) transfer(balance bn.Number) *TestObject {
	t.obj.sdk.Message().Sender().Transfer(t.obj.sdk.Message().Contract().Account().Address(), balance)
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

//DeployContract This is a method of TestObject
func (t *TestObject) DeployContract(
	name string,
	version string,
	orgID string,
	codeHash types.Hash,
	codeData []byte,
	codeDevSig string,
	codeOrgSig string,
	effectHeight int64,
) (result0 types.Address, err types.Error) {
	err.ErrorCode = types.CodeOK
	defer FuncRecover(&err)
	utest.NextBlock(1)
	result0 = t.obj.DeployContract(name, version, orgID, codeHash, codeData, codeDevSig, codeOrgSig, effectHeight, "")
	utest.Commit()
	return
}

//ForbidInternalContract This is a method of TestObject
func (t *TestObject) ForbidInternalContract(contractAddr types.Address, effectHeight int64) (err types.Error) {
	err.ErrorCode = types.CodeOK
	defer FuncRecover(&err)
	utest.NextBlock(1)
	t.obj.ForbidContract(contractAddr, effectHeight)
	utest.Commit()
	return
}
