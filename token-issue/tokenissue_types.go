package tokenissue

import (
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/types"
	"strings"
)

const (
	Recast   = "recast"
	Destroy  = "destroy"
	Transfer = "transfer"
	Lock     = "lock"
	Unlock   = "unlock"
)

type ChainInfo struct {
	SideChainName string        `json:"sideChainName"` //侧链名称
	ChainID       string        `json:"chainID"`       //侧链ID
	NodeNames     []string      `json:"NodeNames"`     //节点名称列表
	OrgName       string        `json:"orgName"`       //侧链所属组织名称
	Owner         types.Address `json:"owner"`         //侧链的所有者地址
	Status        string        `json:"status"`        //侧链状态
}

func keyOfChainInfo(chainID string) string {
	var sb strings.Builder
	sb.WriteString("/sidechain/")
	sb.WriteString(chainID)
	sb.WriteString("/chaininfo")
	return sb.String()
}

func keyOfPeerChainBal(tokenAddr types.Address, chainID string) string {
	var sb strings.Builder
	sb.WriteString("/token/")
	sb.WriteString(tokenAddr)
	sb.WriteString("/")
	sb.WriteString(chainID)
	sb.WriteString("/balance")
	return sb.String()
}

func keyOfSupportSCList(tokenAddr types.Address) string {
	var sb strings.Builder
	sb.WriteString("/token/supportsidechains/")
	sb.WriteString(tokenAddr)
	return sb.String()
}

func keyOfSideChainSupportTokens(sideChainID string) string {
	return "/sidechain/supporttokens/" + sideChainID
}

func keyOfOrganization(orgID string) string {
	return "/organization/" + orgID
}

type Activate struct {
	ChainName    string        `json:"chainName"`    //侧链名称
	Address      types.Address `json:"address"`      //代币地址
	Owner        types.Address `json:"owner"`        //代币owner
	OrgID        string        `json:"orgID"`        //组织ID
	ContractName string        `json:"contractName"` //合约名称
	Name         string        `json:"name"`         //代币名称
	Symbol       string        `json:"symbol"`       //代币符号
	GasPrice     int64         `json:"gasPrice"`     //燃料价格
	OrgName      string        `json:"orgName"`      //组织名称
}

type AssetChange struct {
	Version     string        `json:"version"`     // 合约版本
	Type        string        `json:"type"`        // 资产变更类型：recast、destroy、transfer、lock、unlock
	Token       types.Address `json:"token"`       // 代币地址
	From        types.Address `json:"from"`        // 跨链代币转移资金来源地址
	To          types.Address `json:"to"`          // 跨链代币转移接收者地址
	Value       bn.Number     `json:"value"`       // 代币转移数量
	IBCHash     types.Hash    `json:"ibcHash"`     // 跨链交易 hash
	ChangeItems []ChangeItem  `json:"changeItems"` // 资金修改数据，只保存当前链修改的信息
}

type ChangeItem struct {
	ChainID          string        // 链 ID
	Address          types.Address // IBC 合约地址
	PeerChainID      string        // 侧链 ID(如果当前是在侧链，此字段值与 ChainID 相同)
	PeerChainBalance bn.Number     // 代币余额
}
