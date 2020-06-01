package netgovernance

import (
	"github.com/bcbchain/sdk/sdk"
)

//SetSdk This is a method of ScGovernance
func (ng *NetGovernance) SetSdk(sdk sdk.ISmartContract) {
	ng.sdk = sdk
}

//GetSdk This is a method of ScGovernance
func (ng *NetGovernance) GetSdk() sdk.ISmartContract {
	return ng.sdk
}
