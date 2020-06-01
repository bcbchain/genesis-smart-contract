package tokenbasic

import (
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/types"
)

func (t *TokenBasic) updatePeerChainBal() {
	if t.sdk.Helper().BlockChainHelper().IsSideChain() {
		// 侧链数据正确，不需要修改。
		return
	}
	token := t.sdk.Message().Contract().Token()

	if t._chkSideChainID(t.sdk.Helper().BlockChainHelper().CalcSideChainID("yy")) {
		yyAddValue := bn.N(80497960000)
		t.doUpdatePeerChainBal("yy", token, yyAddValue)
	}

	if t._chkSideChainID(t.sdk.Helper().BlockChainHelper().CalcSideChainID("jiuj")) {
		jiujAddValue := bn.N(2040000)
		t.doUpdatePeerChainBal("jiuj", token, jiujAddValue)
	}
}

func (t *TokenBasic) doUpdatePeerChainBal(chainName string, token types.Address, addValue bn.Number) {
	chainID := t.sdk.Helper().BlockChainHelper().CalcSideChainID(chainName)
	yyPeerBal := t._peerChainBalance(token, chainID)
	t._setPeerChainBalance(token, chainID, yyPeerBal.Add(addValue))
}
