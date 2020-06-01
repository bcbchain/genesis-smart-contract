package tokenissue

import (
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/types"
	"github.com/bcbchain/sdk/utest"
)

var (
	contractName    = "token-issue" //contract name
	contractMethods = []string{"NewToken(string,string,bn.Number,bool,bool,int64)types.Address(types.Address,types.Error)", "Transfer(types.Address,bn.Number)types.Error", "BatchTransfer([]types.Address,bn.Number)types.Error", "AddSupply(bn.Number)types.Error", "Burn(bn.Number)types.Error", "SetOwner(types.Address)types.Error", "SetGasPrice(int64)types.Error"}
	orgID           = "orgAJrbk6Wdf7TCbunrXXS5kKvbWVszhC1T"
)

type TestObject struct {
	obj *TokenIssue
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

func NewTestObject(sender sdk.IAccount) *TestObject {
	return &TestObject{&TokenIssue{sdk: utest.UTP.ISmartContract}}
}
func (t *TestObject) transfer(balance bn.Number) *TestObject {
	t.obj.sdk.Message().Sender().Transfer(t.obj.sdk.Message().Contract().Account().Address(), balance)
	return t
}
func (t *TestObject) setSender(sender sdk.IAccount) *TestObject {
	t.obj.sdk = utest.SetSender(sender.Address())
	return t
}
func (t *TestObject) run() *TestObject {
	t.obj.sdk = utest.ResetMsg()
	return t
}
func (t *TestObject) NewToken(name string,
	symbol string,
	totalSupply bn.Number,
	addSupplyEnabled bool,
	burnEnabled bool,
	gasprice int64) (addr types.Address, err types.Error) {
	err.ErrorCode = types.CodeOK
	defer FuncRecover(&err)

	utest.NextBlock(1)
	addr = t.obj.NewToken(name, symbol, totalSupply, addSupplyEnabled, burnEnabled, gasprice)
	utest.Commit()
	return
}

func (t *TestObject) Transfer(to types.Address, value bn.Number) (err types.Error) {
	err.ErrorCode = types.CodeOK
	defer FuncRecover(&err)
	utest.NextBlock(1)
	t.obj.Transfer(to, value)
	utest.Commit()
	return
}

func (t *TestObject) BatchTransfer(toList []types.Address, value bn.Number) (err types.Error) {
	err.ErrorCode = types.CodeOK
	defer FuncRecover(&err)
	utest.NextBlock(1)
	t.obj.BatchTransfer(toList, value)
	utest.Commit()
	return
}

func (t *TestObject) AddSupply(value bn.Number) (err types.Error) {
	err.ErrorCode = types.CodeOK
	defer FuncRecover(&err)
	utest.NextBlock(1)
	t.obj.AddSupply(value)

	utest.Commit()
	return
}

func (t *TestObject) Burn(value bn.Number) (err types.Error) {
	err.ErrorCode = types.CodeOK
	defer FuncRecover(&err)
	utest.NextBlock(1)
	t.obj.Burn(value)
	utest.Commit()
	return
}

func (t *TestObject) SetOwner(newOnwer types.Address) (err types.Error) {
	err.ErrorCode = types.CodeOK
	defer FuncRecover(&err)
	utest.NextBlock(1)
	t.obj.SetOwner(newOnwer)
	utest.Commit()
	return
}

func (t *TestObject) SetGasPrice(value int64) (err types.Error) {
	err.ErrorCode = types.CodeOK
	defer FuncRecover(&err)
	utest.NextBlock(1)
	t.obj.SetGasPrice(value)
	utest.Commit()
	return
}
