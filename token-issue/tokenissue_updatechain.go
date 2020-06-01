package tokenissue

import (
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/forx"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
)

func (ti *TokenIssue) upgradeTokenContract() {
	// upgrade token's contract
	contracts := ti.getAllTokenContract()
	tokenIssue := ti.sdk.Message().Contract()
	stateHelper := ti.sdk.Helper().StateHelper()

	forx.Range(*contracts, func(i int, c std.Contract) bool {
		// update last token contract
		c.LoseHeight = ti.sdk.Block().Height()
		stateHelper.McSet(std.KeyOfContract(c.Address), &c)

		// calc token contract's address
		addr := ti.sdk.Helper().BlockChainHelper().CalcContractAddress(
			c.Name,
			tokenIssue.Version(),
			tokenIssue.OrgID())
		newContract := std.Contract{
			Address:      addr,
			Account:      c.Account,
			Owner:        c.Owner,
			Name:         c.Name,
			Version:      tokenIssue.Version(),
			CodeHash:     tokenIssue.CodeHash(),
			EffectHeight: tokenIssue.EffectHeight(),
			LoseHeight:   0,
			KeyPrefix:    c.KeyPrefix,
			Methods:      tokenIssue.Methods(),
			Interfaces:   tokenIssue.Interfaces(),
			IBCs:         tokenIssue.IBCs(),
			Token:        c.Token,
			OrgID:        c.OrgID,
			ChainVersion: c.ChainVersion,
		}
		stateHelper.McSet(std.KeyOfContract(addr), &newContract)

		// add new contract to contract version information with addresses and effectHeights
		//var cvl std.ContractVersionList
		key := std.KeyOfContractsWithName(c.OrgID, c.Name)
		cvl := stateHelper.McGetEx(key, new(std.ContractVersionList)).(*std.ContractVersionList)
		cvl.ContractAddrList = append(cvl.ContractAddrList, addr)
		cvl.EffectHeights = append(cvl.EffectHeights, newContract.EffectHeight)
		stateHelper.McSet(key, &cvl)

		// add new contract address to account's contract addresses
		key = std.KeyOfAccountContracts(newContract.Owner)
		cons := *stateHelper.McGetEx(key, new([]types.Address)).(*[]types.Address)
		cons = append(cons, addr)
		stateHelper.McSet(key, &cons)
		return true
	})
}

func (ti *TokenIssue) updatePeerBal() {
	if ti.sdk.Helper().BlockChainHelper().IsSideChain() {
		// 侧链数据正确，不需要修改。
		return
	}
	tokenDC := ti.sdk.Helper().TokenHelper().TokenOfSymbol("dc")
	if tokenDC == nil {
		return
	}

	tokenDCAddr := tokenDC.Address()
	yyChainID := ti.sdk.Helper().BlockChainHelper().CalcSideChainID("yy")
	if ti._chkSideChainID(yyChainID) {
		yyAddValue := bn.N(-3900000)
		ti.doUpdatePeerBal("yy", tokenDCAddr, yyAddValue)
	}

	jiujChainID := ti.sdk.Helper().BlockChainHelper().CalcSideChainID("jiuj")
	if ti._chkSideChainID(jiujChainID) {
		jiujAddValue := bn.N(3900000)
		ti.doUpdatePeerBal("jiuj", tokenDCAddr, jiujAddValue)
	}
}

func (ti *TokenIssue) doUpdatePeerBal(chainName string, token types.Address, addValue bn.Number) {
	chainID := ti.sdk.Helper().BlockChainHelper().CalcSideChainID(chainName)
	yyPeerBal := ti._peerChainBalance(token, chainID)
	ti._setPeerChainBalance(token, chainID, yyPeerBal.Add(addValue))
}
