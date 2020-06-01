package tokenbasic

import (
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
)

func (t *TokenBasic) _chkSideChainID(chainID string) bool {
	return t.sdk.Helper().StateHelper().Check(keyOfChainInfo(chainID))
}

func (t *TokenBasic) _setAccountBalance(account, token types.Address, balance bn.Number) {
	key := std.KeyOfAccountToken(account, token)
	info := std.AccountInfo{
		Address: t.sdk.Message().Contract().Token(),
		Balance: balance,
	}
	t.sdk.Helper().StateHelper().Set(key, &info)
}

func (t *TokenBasic) _accountBalance(account, token types.Address) bn.Number {
	key := std.KeyOfAccountToken(account, token)
	return t.sdk.Helper().StateHelper().GetEx(key, new(std.AccountInfo)).(*std.AccountInfo).Balance
}

func (t *TokenBasic) _peerChainBalance(tokenAddr types.Address, chainID string) bn.Number {
	b := bn.N(0)
	return *t.sdk.Helper().StateHelper().GetEx(keyOfPeerChainBal(tokenAddr, chainID), &b).(*bn.Number)
}

func (t *TokenBasic) _setPeerChainBalance(tokenAddr types.Address, chainID string, balance bn.Number) {
	key := keyOfPeerChainBal(tokenAddr, chainID)
	t.sdk.Helper().StateHelper().Set(key, &balance)
}
