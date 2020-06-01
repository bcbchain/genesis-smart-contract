package smartcontract

import (
	"github.com/bcbchain/sdk/sdk"
)

//SetSdk set sdk
func (s *SmartContract) SetSdk(sdk sdk.ISmartContract) {
	s.sdk = sdk
}

//GetSdk get  sdk
func (s *SmartContract) GetSdk() sdk.ISmartContract {
	return s.sdk
}
