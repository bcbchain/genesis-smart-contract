package governance

import (
	"github.com/bcbchain/sdk/sdk"
)

//SetSdk This is a method of Validator
func (g *Governance) SetSdk(sdk sdk.ISmartContract) {
	g.sdk = sdk
}

//GetSdk This is a method of Validator
func (g *Governance) GetSdk() sdk.ISmartContract {
	return g.sdk
}
