package tokenbasic

import (
	"fmt"
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/types"
)

const (
	Recast   = "recast"
	Destroy  = "destroy"
	Transfer = "transfer"
	Lock     = "lock"
	Unlock   = "unlock"
)

type IBCTransfer struct {
	Height    int64         `josn:"height"`    // 发起跨链转账高度
	From      types.Address `json:"from"`      // 发起方地址
	To        types.Address `json:"to"`        // 接收方地址
	Value     bn.Number     `json:"value"`     // 跨链转账金额(单位为cong)
	Lock      bn.Number     `json:"lock"`      // 锁定金额
	Destroy   bn.Number     `json:"destroy"`   // 销毁金额
	PreCoin   bn.Number     `json:"preCoin"`   // 预重铸金额
	FinalCoin bn.Number     `json:"finalCoin"` // 重铸金额
}

func keyOfChainInfo(chainID string) string {
	return fmt.Sprintf("/sidechain/%s/chaininfo", chainID)
}

func keyOfPeerChainBal(tokenAddr types.Address, chainID string) string {
	return fmt.Sprintf("/token/%s/%s/balance", tokenAddr, chainID)
}

type AssetChange struct {
	Version     string        `json:"version"`     // 版本，通过该字段实现合约兼容性处理
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
