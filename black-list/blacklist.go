package blacklist

import (
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/forx"
	"github.com/bcbchain/sdk/sdk/types"
)

//Blacklist This is struct of contract
//@:contract:black-list
//@:version:2.0
//@:organization:orgJgaGConUyK81zibntUBjQ33PKctpk1K1G
//@:author:5e8339cb1a5cce65602fd4f57e115905348f7e83bcbe38dd77694dbe1f8903c9
type Blacklist struct {
	sdk sdk.ISmartContract
}

//InitChain Constructor of this Blacklist
//@:constructor
func (b *Blacklist) InitChain() {
}

//AddAddress add addresses to blacklist
//@:public:method:gas[500]
func (b *Blacklist) AddAddress(blacklist []types.Address) {
	sdk.RequireOwner()
	sdk.Require(len(blacklist) > 0,
		types.ErrInvalidParameter, "Cannot be empty blacklist.")

	forx.Range(blacklist, func(i int, addr types.Address) bool {
		sdk.RequireAddress(addr)
		sdk.Require(addr != b.sdk.Message().Contract().Owner().Address(),
			types.ErrInvalidParameter, "Cannot contain owner address in blacklist")

		b._setBlacklistAddr(addr, "true")
		return true
	})

	b.emitAddAddress(blacklist)
}

//DelAddress delete addresses from blacklist
//@:public:method:gas[500]
func (b *Blacklist) DelAddress(blacklist []types.Address) {
	sdk.RequireOwner()
	sdk.Require(len(blacklist) > 0,
		types.ErrInvalidParameter, "Cannot be empty blacklist.")

	forx.Range(blacklist, func(i int, addr types.Address) bool {
		sdk.RequireAddress(addr)
		sdk.Require(addr != b.sdk.Message().Contract().Owner().Address(),
			types.ErrInvalidParameter, "Cannot contain owner address in blacklist")

		b._setBlacklistAddr(addr, "false")
		return true
	})

	b.emitDelAddress(blacklist)
}

//SetOwner set new owner for current contract
//@:public:method:gas[500]
func (b *Blacklist) SetOwner(newOwnerAddr types.Address) {
	b.sdk.Message().Contract().SetOwner(newOwnerAddr)
}

func (b *Blacklist) emitAddAddress(blacklist []types.Address) {
	type addAddress struct {
		Blacklist []types.Address `json:"blacklist"`
	}

	b.sdk.Helper().ReceiptHelper().Emit(
		addAddress{
			Blacklist: blacklist,
		},
	)
}

func (b *Blacklist) emitDelAddress(blacklist []types.Address) {
	type delAddress struct {
		Blacklist []types.Address `json:"blacklist"`
	}

	b.sdk.Helper().ReceiptHelper().Emit(
		delAddress{
			Blacklist: blacklist,
		},
	)
}
