package blacklist

import "github.com/bcbchain/sdk/sdk/types"

func (b *Blacklist) _blacklistAddr(addr types.Address) string {
	return *b.sdk.Helper().StateHelper().GetEx("/blacklist/"+addr, new(string)).(*string)
}

func (b *Blacklist) _setBlacklistAddr(addr types.Address, status string) {
	b.sdk.Helper().StateHelper().Set("/blacklist/"+addr, &status)
}

func (b *Blacklist) _chkBlacklistAddr(addr types.Address) bool {
	return b.sdk.Helper().StateHelper().Check("/blacklist/" + addr)
}
