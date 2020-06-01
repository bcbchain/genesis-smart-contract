package tokenbasic

import (
	"github.com/bcbchain/sdk/sdk"
)

//SetSdk - set sdk
func (t *TokenBasic) SetSdk(sdk sdk.ISmartContract) {
	t.sdk = sdk
}

//GetSdk - get sdk
func (t *TokenBasic) GetSdk() sdk.ISmartContract {
	return t.sdk
}
