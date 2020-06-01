package blacklist

import (
	"github.com/bcbchain/sdk/sdk"
)

//SetSdk This is a method of Blacklist
func (b *Blacklist) SetSdk(sdk sdk.ISmartContract) {
	b.sdk = sdk
}

//GetSdk This is a method of Blacklist
func (b *Blacklist) GetSdk() sdk.ISmartContract {
	return b.sdk
}
