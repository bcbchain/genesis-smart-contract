package tokenissue

import (
	"fmt"
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/forx"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
	"github.com/bcbchain/sdk/sdkimpl/object"
	"sort"
	"strings"
)

// newToken register a token
func (ti *TokenIssue) newToken(
	tokenAddress types.Address,
	owner types.Address,
	name string,
	symbol string,
	totalSupply bn.Number,
	addSupplyEnabled bool,
	burnEnabled bool,
	gasPrice int64,
	orgName string) {
	// If it owns a token, it's not the genesis contract, cannot execute this function.
	sdk.Require(ti.sdk.Message().Contract().Token() == "",
		types.ErrNoAuthorization, "The contract has already registered a token")

	sdk.Require(totalSupply.IsEqualI(0) || totalSupply.IsGEI(minTotalsupply),
		types.ErrInvalidParameter, "Invalid total supply")

	sdk.Require(gasPrice >= ti.sdk.Helper().TokenHelper().BaseGasPrice() && gasPrice <= maxGasPrice,
		types.ErrInvalidParameter, "Invalid gas price")

	if orgName == "" {
		token := ti.setNewToken(tokenAddress, owner, name, symbol, totalSupply, addSupplyEnabled, burnEnabled, gasPrice, "", "")
		ti.setNewContract(nil, token, ti.contractNameBRC20(name))
	} else {
		token := ti.setNewToken(tokenAddress, owner, name, symbol, totalSupply, addSupplyEnabled, burnEnabled, gasPrice, orgName, "BRC30")
		ti.setNewContract(nil, token, ti.contractNameBRC30(orgName, name))
	}

}

func (ti *TokenIssue) newTokenForThirdOrg(
	oldContract *std.Contract,
	tokenAddress types.Address,
	name string,
	symbol string,
	gasPrice int64,
	orgName string,
	proto string) {

	transferMethodExist := false
	transferInterExist := false
	transferWithNoteInterExist := false

	// check the contract that it must defined standard transfer method
	forx.Range(oldContract.Methods, func(index int, method std.Method) bool {
		if method.MethodID == "44d8ca60" { //transfer method
			transferMethodExist = true
			return forx.Break
		}
		return forx.Continue
	})

	// check the contract that it must defined standard transfer interface
	forx.Range(oldContract.Interfaces, func(index int, inter std.Method) bool {
		if inter.MethodID == "44d8ca60" { //transfer method
			transferInterExist = true
			return forx.Break
		}
		return forx.Continue
	})

	// check the contract that it must defined standard transfer interface
	forx.Range(oldContract.Interfaces, func(index int, inter std.Method) bool {
		if inter.MethodID == "838b8172" { //transfer method
			transferWithNoteInterExist = true
			return forx.Break
		}
		return forx.Continue
	})

	sdk.Require(transferMethodExist == true &&
		transferInterExist == true &&
		transferWithNoteInterExist == true,
		types.ErrInvalidParameter, "This contract never defined standard transfer method")

	token := ti.setNewToken(tokenAddress, oldContract.Owner, name, symbol, bn.N(0), false, false, gasPrice, orgName, proto)
	ti.setNewContractEx(oldContract, token)
}

func (ti *TokenIssue) setNewToken(
	address, owner types.Address,
	name, symbol string,
	totalSupply bn.Number,
	addSupplyEnabled, burnEnabled bool,
	gasPrice int64, orgName string,
	Proto string,
) *std.Token {

	token := std.Token{
		Address:          address,
		Owner:            owner,
		Name:             name,
		Symbol:           symbol,
		TotalSupply:      totalSupply,
		AddSupplyEnabled: addSupplyEnabled,
		BurnEnabled:      burnEnabled,
		GasPrice:         gasPrice,
		OrgName:          orgName,
		Proto:            Proto,
	}

	sdb := ti.sdk.Helper().StateHelper()
	sdb.Set(std.KeyOfToken(token.Address), &token)

	if orgName == "" {
		sdb.Set(std.KeyOfTokenWithName(token.Name), &token.Address)
		sdb.Set(std.KeyOfTokenWithSymbol(token.Symbol), &token.Address)
	} else {
		sdb.Set(std.KeyOfBRC30TokenWithName(token.OrgName, token.Name), &token.Address)
		sdb.Set(std.KeyOfBRC30TokenWithSymbol(token.OrgName, token.Symbol), &token.Address)
	}

	allTokenAddr := new([]string)
	allTokenAddr = sdb.GetEx(std.KeyOfAllToken(), allTokenAddr).(*[]string)
	*allTokenAddr = append(*allTokenAddr, token.Address)
	sdb.Set(std.KeyOfAllToken(), allTokenAddr)

	balance := std.AccountInfo{
		Address: token.Address,
		Balance: totalSupply,
	}
	sdb.Set(std.KeyOfAccountToken(token.Owner, token.Address), &balance)
	acc := ti.sdk.Helper().AccountHelper().AccountOf(token.Owner)
	acc.(*object.Account).AddAccountTokenKey(std.KeyOfAccountToken(token.Owner, token.Address))

	ti.emitNewToken(&token)
	if totalSupply.IsGreaterThanI(0) {
		ti.emitTransfer(&token)
	}
	return &token
}

func (ti *TokenIssue) setNewContractEx(oldContract *std.Contract, token *std.Token) {
	ti.setNewContract(oldContract, token, oldContract.Name)
}

func (ti *TokenIssue) setNewContract(oldContract *std.Contract, token *std.Token, contractName string) {
	sdb := ti.sdk.Helper().StateHelper()

	contract := std.Contract{
		Address:      token.Address,
		Owner:        token.Owner,
		Name:         contractName,
		EffectHeight: ti.sdk.Block().Height() + 1,
		Token:        token.Address,
	}
	if oldContract != nil {
		if oldContract.Address == token.Address {
			oldContract.Token = token.Address
			sdb.Set(std.KeyOfContract(oldContract.Address), &oldContract)
			return
		} else {
			contract.OrgID = oldContract.OrgID
			contract.Version = oldContract.Version
			contract.CodeHash = oldContract.CodeHash
			contract.Methods = oldContract.Methods
			contract.Interfaces = oldContract.Interfaces
			contract.IBCs = oldContract.IBCs
			contract.ChainVersion = oldContract.ChainVersion

			// lose old contract
			oldContract.LoseHeight = ti.sdk.Block().Height() + 1
			sdb.Set(std.KeyOfContract(oldContract.Address), &oldContract)
		}
	} else {
		contract.OrgID = ti.sdk.Message().Contract().OrgID()
		contract.Version = ti.sdk.Message().Contract().Version()
		contract.CodeHash = ti.sdk.Message().Contract().CodeHash()
		contract.Methods = ti.sdk.Message().Contract().Methods()
		contract.Interfaces = ti.sdk.Message().Contract().Interfaces()
		contract.IBCs = ti.sdk.Message().Contract().IBCs()
		contract.ChainVersion = ti.sdk.Message().Contract().ChainVersion()
	}
	contract.Account = ti.sdk.Helper().BlockChainHelper().CalcAccountFromName(contractName, contract.OrgID)
	sdb.Set(std.KeyOfContract(contract.Address), &contract)

	contractVersions := ti._contractVersions(contract.OrgID, contractName)
	contractVersions.ContractAddrList = append(contractVersions.ContractAddrList, contract.Address)
	contractVersions.EffectHeights = append(contractVersions.EffectHeights, contract.EffectHeight)
	sdb.Set(std.KeyOfContractsWithName(contract.OrgID, contract.Name), &contractVersions)

	var cons []types.Address
	key := std.KeyOfAccountContracts(contract.Owner)
	cons = *sdb.GetEx(key, &cons).(*[]types.Address)
	cons = append(cons, contract.Address)
	sdb.Set(key, &cons)

}

func (ti *TokenIssue) addSupportSideChain(tokenAddress types.Address, sideChainID string) {
	sideChainIDs := ti._supportSideChains(tokenAddress)

	index := sort.SearchStrings(sideChainIDs, sideChainID)
	if index == len(sideChainIDs) { //not found
		sideChainIDs = append(sideChainIDs, sideChainID)
	} else if sideChainIDs[index] != sideChainID { //not found
		insertChainIDs := append([]string{sideChainID}, sideChainIDs[index:]...)
		sideChainIDs = append(sideChainIDs[:index], insertChainIDs...)
	}
	ti._setSupportSideChains(tokenAddress, sideChainIDs)
}

func (ti *TokenIssue) addSideChainSupportTokens(tokenAddress types.Address, sideChainID string) {
	tokens := ti._sideChainSupportTokens(sideChainID)

	index := sort.SearchStrings(tokens, tokenAddress)
	if index == len(tokens) { //not found
		tokens = append(tokens, tokenAddress)
	} else if tokens[index] != tokenAddress { //not found
		insertTokens := append([]string{tokenAddress}, tokens[index:]...)
		tokens = append(tokens[:index], insertTokens...)
	}
	ti._setSideChainSupportTokens(sideChainID, tokens)
}

func (ti *TokenIssue) isValidLength(name, symbol string) bool {

	if ti.sdk.Helper().GenesisHelper().Token().Owner().Address() == ti.sdk.Message().Sender().Address() {
		if len(name) < minNameLenForGenesisOrg || len(name) > maxNameLen {
			return false
		}
		if len(symbol) < minSymbolLenForGenesisOrg || len(symbol) > maxSymbolLen {
			return false
		}
	} else {
		if len(name) < minNameLenForOtherOrg || len(name) > maxNameLen {
			return false
		}
		if len(symbol) < minSymbolLenForOtherOrg || len(symbol) > maxSymbolLen {
			return false
		}
	}

	return true
}

func (ti *TokenIssue) isValidNameAndSymbol(orgName, name, symbol string, totalSupply bn.Number) bool {

	if ti.isValidLength(name, symbol) == false {
		return false
	}

	var t1, t2 sdk.IToken
	if orgName != "" {
		ti.checkOrganization(orgName)
		t1 = ti.sdk.Helper().TokenHelper().TokenOfNameBRC30(orgName, name)
		t2 = ti.sdk.Helper().TokenHelper().TokenOfSymbolBRC30(orgName, symbol)
	} else {
		t1 = ti.sdk.Helper().TokenHelper().TokenOfName(name)
		t2 = ti.sdk.Helper().TokenHelper().TokenOfSymbol(symbol)
	}

	if t1 == nil && t2 == nil {
		//valid new token name and symbol
		return true
	} else if t1 != nil && t2 != nil && t1.Address() == t2.Address() &&
		t1.TotalSupply().IsEqualI(0) && totalSupply.IsGreaterThanI(0) &&
		t1.Owner().Address() == ti.sdk.Message().Sender().Address() {
		//Registered with total supply = 0, return true
		return true
	}
	return false
}

func (ti *TokenIssue) getAllTokenContract() *[]std.Contract {
	allTokenContract := make([]std.Contract, 0)

	sdb := ti.sdk.Helper().StateHelper()
	allTokenAddrs := ti._allToken()
	forx.Range(allTokenAddrs, func(i int, addr string) bool {
		contract := *sdb.McGetEx(std.KeyOfContract(addr), new(std.Contract)).(*std.Contract)
		if !strings.HasPrefix(contract.Name, "token-templet-") {
			return forx.Continue
		}

		if contract.LoseHeight == 0 {
			allTokenContract = append(allTokenContract, contract)
			return forx.Continue
		} else {
			contractVersions := ti._contractVersions(contract.OrgID, contract.Name)
			forx.Range(contractVersions.ContractAddrList, func(i int, contractAddr string) {
				contract := *sdb.McGetEx(std.KeyOfContract(contractAddr), new(std.Contract)).(*std.Contract)
				if contract.LoseHeight == 0 {
					allTokenContract = append(allTokenContract, contract)
				}
			})

		}
		return true
	})

	return &allTokenContract
}

func (ti *TokenIssue) activate(orgName, tokenName, chainName string) {

	sdk.RequireMainChain()
	sideChainID := ti.sdk.Helper().BlockChainHelper().CalcSideChainID(chainName)
	chainInfo := ti._chainInfo(sideChainID)
	sdk.Require(chainInfo.ChainID == sideChainID,
		types.ErrInvalidParameter, "invalid chainName")
	sdk.Require(ti.sdk.Message().Sender().Address() == chainInfo.Owner,
		types.ErrNoAuthorization, "")
	sdk.Require(chainInfo.Status == "ready",
		types.ErrInvalidParameter,
		fmt.Sprintf("expected chain status: ready, obtain: %s", chainInfo.Status))

	var token sdk.IToken
	if orgName != "" {
		token = ti.sdk.Helper().TokenHelper().TokenOfNameBRC30(orgName, tokenName)
	} else {
		token = ti.sdk.Helper().TokenHelper().TokenOfName(tokenName)
	}

	sdk.Require(token != nil,
		types.ErrInvalidParameter, "invalid token Name or orgName")
	sdk.Require(token.TotalSupply().IsGreaterThanI(0),
		types.ErrInvalidParameter, "invalid total supply")

	ti.addSupportSideChain(token.Address(), sideChainID)

	ti.sdk.Helper().IBCHelper().Run(func() {
		ti.emitActivate(chainName, token, token.GasPrice(), orgName)
	}).Broadcast()
}

func (ti *TokenIssue) checkOrganization(orgName string) {

	sdk.Require(len(orgName) > 0,
		types.ErrInvalidParameter, "Invalid orgName")

	orgID := ti.sdk.Helper().BlockChainHelper().CalcOrgID(orgName)

	sdk.Require(ti._chkOrganization(orgID),
		types.ErrInvalidParameter, "There is no organization with name "+orgName)
}
