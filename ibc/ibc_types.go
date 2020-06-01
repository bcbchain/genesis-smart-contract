package IBC

import (
	"github.com/bcbchain/sdk/sdk/types"
	"strconv"
	"strings"
)

type Validator struct {
	chainID string
	pubKey  types.PubKey
}

//InfoOfValidator validator struct
type InfoOfValidator struct {
	PubKey     types.PubKey `json:"nodepubkey,omitempty"` //节点公钥
	Power      int64        `json:"power,omitempty"`      //节点记账权重
	RewardAddr string       `json:"rewardaddr,omitempty"` //节点接收奖励的地址
	Name       string       `json:"name,omitempty"`       //节点名称
	NodeAddr   string       `json:"nodeaddr,omitempty"`   //节点地址
}

type ChainInfo struct {
	SideChainName string        `json:"sideChainName"` //侧链名称
	ChainID       string        `json:"chainID"`       //侧链ID
	NodeNames     []string      `json:"NodeNames"`     //节点名称列表
	OrgName       string        `json:"orgName"`       //侧链所属组织名称
	Owner         types.Address `json:"owner"`         //侧链的所有者地址
	Status        string        `json:"status"`        //侧链状态
}

type MessageIndex struct {
	Height  int64      `json:"height"`  // 当前序号所在区块高度
	IbcHash types.Hash `json:"ibcHash"` // 跨链事务hash
}

func keyOfChainInfo(chainID string) string {
	var buf strings.Builder
	buf.WriteString("/sidechain/")
	buf.WriteString(chainID)
	buf.WriteString("/chaininfo")
	return buf.String()
}

func keyOfSequence(queueID string) string {
	var buf strings.Builder
	buf.WriteString("/ibc/seq/")
	buf.WriteString(queueID)
	return buf.String()
}

func keyOfState(ibcHash types.Hash) string {
	var buf strings.Builder
	buf.WriteString("/ibc/")
	buf.WriteString(ibcHash.String())
	buf.WriteString("/state")
	return buf.String()
}

func keyOfPackets(ibcHash types.Hash) string {
	var buf strings.Builder
	buf.WriteString("/ibc/")
	buf.WriteString(ibcHash.String())
	buf.WriteString("/packets")
	return buf.String()
}

func keyOfSideChainIDs() string {
	return "/sidechain/chainid/all"
}

func keyOfSequenceHeight(queueID string, seq uint64) string {
	seqStr := strconv.Itoa(int(seq))
	var buf strings.Builder
	buf.WriteString("/ibc/seq/")
	buf.WriteString(queueID)
	buf.WriteString("/")
	buf.WriteString(seqStr)
	return buf.String()
}

// ------------- proof -----------------
func keyOfLastQueueHash(queueID string) string {
	return "/ibc/" + queueID + "/lastQueueHash"
}

// ------------- openUrls -----------------
type setOpenURL struct {
	ChainID  string   `json:"sideChainID"`
	OpenURLs []string `json:"openURLs"`
}

func keyOfSetOpenURLs(chainId string) string {
	return "/sidechain/" + chainId + "/openurls"
}

// ------------- gasPriceRatio -----------------
type setGasPriceRatio struct {
	ChainName     string `json:"chainName"`
	ChainID       string `json:"chainID"`
	GasPriceRatio string `json:"gasPriceRatio"`
}

// ------------- removeSideChainToken -----------------
type removeSideChainToken struct {
	ChainID    string          `json:"chainID"`
	TokenAddrs []types.Address `json:"tokenAddrs"`
}
