package organization

import (
	"github.com/bcbchain/sdk/sdk"
)

//SetSdk This is a method of Organization
func (o *Organization) SetSdk(sdk sdk.ISmartContract) {
	o.sdk = sdk
}

//GetSdk This is a method of Organization
func (o *Organization) GetSdk() sdk.ISmartContract {
	return o.sdk
}
