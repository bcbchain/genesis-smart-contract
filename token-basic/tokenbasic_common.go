package tokenbasic

func (t *TokenBasic) isCallByIBC() bool {
	origins := t.sdk.Message().Origins()
	if len(origins) < 1 {
		return false
	}

	lastContract := t.sdk.Helper().ContractHelper().ContractOfAddress(origins[len(origins)-1])
	if lastContract.Name() == "ibc" && lastContract.OrgID() == t.sdk.Helper().GenesisHelper().OrgID() {
		return true
	}
	return false
}
