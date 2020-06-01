package mining

import (
	"github.com/bcbchain/sdk/sdk"
)

//SetSdk This is a method of Mining
func (m *Mining) SetSdk(sdk sdk.ISmartContract) {
	m.sdk = sdk
}

//GetSdk This is a method of Mining
func (m *Mining) GetSdk() sdk.ISmartContract {
	return m.sdk
}
