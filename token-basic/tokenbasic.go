package tokenbasic

import (
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
)

const (
	maxGasPrice = 1000000000
)

//TokenBasic
//@:contract:token-basic
//@:version:2.3
//@:organization:orgJgaGConUyK81zibntUBjQ33PKctpk1K1G
//@:author:5e8339cb1a5cce65602fd4f57e115905348f7e83bcbe38dd77694dbe1f8903c9
type TokenBasic struct {
	sdk sdk.ISmartContract
}

//InitChain: construct function
//@:constructor
func (t *TokenBasic) InitChain() {

}

//UpdateChain UpdateChain of this TokenBasic
//@:constructor
func (t *TokenBasic) UpdateChain() {
	// fix bug peerChainBalance for yy & jiuj side chain
	t.updatePeerChainBal()
}

// Transfer is used to transfer token from sender to another specified account
// In the TokenBasic contract, it's  only used to transfer the basic token
//@:public:method:gas[500]
//@:public:interface:gas[450]
func (t *TokenBasic) Transfer(to types.Address, value bn.Number) {

	if t.sdk.Helper().BlockChainHelper().IsPeerChainAddress(to) {
		// cross chain transfer
		t.sdk.Helper().IBCHelper().Run(func() {
			t.lock(to, value)
		}).Register(t.sdk.Helper().BlockChainHelper().GetChainID(to))
	} else {
		// Do transfer
		t.sdk.Message().Sender().TransferWithNote(to, value, t.sdk.Tx().Note())
	}
}

//SetGasPrice is used to set gas price for token-basic contract
//@:public:method:gas[2000]
func (t *TokenBasic) SetGasPrice(value int64) {
	sdk.RequireMainChain()

	t.sdk.Helper().IBCHelper().Run(func() {
		t.sdk.Helper().TokenHelper().Token().SetGasPrice(value)
	}).Broadcast()
}

// SetBaseGasPrice is used to set base gas price.
// The base gas price is a minimum limit to all of token's.
// All new tokens' gas price could not be set to a value that smaller than base gas price.
//@:public:method:gas[2000]
func (t *TokenBasic) SetBaseGasPrice(value int64) {
	sdk.RequireMainChain()
	sdk.RequireOwner()

	sdk.Require(value > 0 && value <= maxGasPrice,
		types.ErrInvalidParameter, "Invalid base gas price")

	t.sdk.Helper().StateHelper().Set(std.KeyOfTokenBaseGasPrice(), &value)

	type BaseGasPrice struct {
		Value int64 `json:"value"`
	}
	t.sdk.Helper().ReceiptHelper().Emit(
		BaseGasPrice{
			Value: value,
		})
}
