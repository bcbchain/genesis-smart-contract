package tokenissue

import (
	"github.com/bcbchain/sdk/sdk"
)

//SetSdk set sdk
func (ti *TokenIssue) SetSdk(sdk sdk.ISmartContract) {
	ti.sdk = sdk
}

//GetSdk get  sdk
func (ti *TokenIssue) GetSdk() sdk.ISmartContract {
	return ti.sdk
}
