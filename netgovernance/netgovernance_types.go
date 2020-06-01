package netgovernance

import (
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
	"strconv"
	"strings"
)

type AddressVersion int32 // 侧链地址版本

const (
	AddressVersion1 = iota // 版本：主链ID + [侧链名称] + 地址编码
	AddressVersion2        // 版本：主链ID + 地址编码 + "0" + base58.encode(侧链名称)
)

const (
	PubKeyLen  = 32
	MaxNameLen = 40
	Init       = "init"
	Ready      = "ready"
	Clear      = "clear"
	Disabled   = "disabled"
)

type GenesisInfo struct {
	ChainID      string      `json:"chain_id"`
	ChainVersion string      `json:"chain_version"`
	GenesisTime  string      `json:"genesis_time"`
	AppHash      string      `json:"app_hash"`
	AppState     AppState    `json:"app_state"`
	Validators   []Validator `json:"validators"`
}

type OrgBind struct {
	OrgName string        `json:"orgName"`
	Owner   types.Address `json:"owner"`
}

type MainChainInfo struct {
	OpenUrls   []string             `json:"openUrls"`
	Validators map[string]Validator `json:"validators"`
}

type AppState struct {
	Organization   string        `json:"organization"`
	GasPriceRatio  string        `json:"gas_price_ratio"`
	Token          std.Token     `json:"token"`
	RewardStrategy []Reward      `json:"rewardStrategy"`
	Contracts      []Contract    `json:"contracts"`
	OrgBind        OrgBind       `json:"orgBind"`
	MainChain      MainChainInfo `json:"mainChain"`
}

// Reward reward info
type Reward struct {
	Name          string `json:"name"`          // 被奖励者名称
	RewardPercent string `json:"rewardPercent"` // 奖励比例
	Address       string `json:"address"`       // 被奖励者地址
}

// Contract contract info
type Contract struct {
	Name       string    `json:"name,omitempty"`
	Version    string    `json:"version,omitempty"`
	Owner      string    `json:"owner,omitempty"`
	Code       string    `json:"code"`
	CodeHash   string    `json:"codeHash,omitempty"`
	CodeDevSig Signature `json:"codeDevSig,omitempty"`
	CodeOrgSig Signature `json:"codeOrgSig,omitempty"`
}

// Signature sig for contract code
type Signature struct {
	PubKey    string `json:"pubkey"`
	Signature string `json:"signature"`
}

type ChainInfo struct {
	SideChainName string         `json:"sideChainName"`         //侧链名称
	ChainID       string         `json:"chainID"`               //侧链ID
	NodeNames     []string       `json:"NodeNames"`             //节点名称列表
	OrgName       string         `json:"orgName"`               //侧链所属组织名称
	Owner         types.Address  `json:"owner"`                 //侧链的所有者地址
	Height        int64          `json:"height"`                //侧链创世时在主链上的高度
	Status        string         `json:"status"`                //侧链状态
	GasPriceRatio string         `json:"gasPriceRatio"`         //燃料价格调整比例
	AddrVersion   AddressVersion `json:"addrVersion,omitempty"` //地址版本
}

func keyOfChainInfo(chainID string) string {
	return "/sidechain/" + chainID + "/chaininfo"
}

//InfoOfValidator validator struct
type Validator struct {
	PubKey     types.PubKey `json:"nodepubkey,omitempty"`  //节点公钥
	Power      int64        `json:"power,omitempty"`       //节点记账权重
	RewardAddr string       `json:"reward_addr,omitempty"` //节点接收奖励的地址
	Name       string       `json:"name,omitempty"`        //节点名称
	NodeAddr   string       `json:"nodeaddr,omitempty"`    //节点地址
}

type ContractData struct {
	Name     string         `json:"name"`
	Version  string         `json:"version"`
	CodeByte types.HexBytes `json:"codeByte"`
}

func keyOfOpenURLs(chainId string) string {
	return "/sidechain/" + chainId + "/openurls"
}

func keyOfOrganization(orgID string) string {
	return "/organization/" + orgID
}

func keyOfBCBValidator(nodeAddr types.Address) string {
	return "/validator/" + nodeAddr
}

func keyOfContractCode(contractAddr types.Address) string {
	return "/contract/code/" + contractAddr
}

func keyOfSideChainIDs() string {
	return "/sidechain/chainid/all"
}

func keyOfWorldAppState() string {
	return "/world/appstate"
}

func keyOfSequence(queueID string) string {
	var buf strings.Builder
	buf.WriteString("/ibc/seq/")
	buf.WriteString(queueID)
	return buf.String()
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

func keyOfLastQueueHash(queueID string) string {
	return "/ibc/" + queueID + "/lastQueueHash"
}

func keyOfPeerChainBal(tokenAddr types.Address, chainID string) string {
	return "/token/" + tokenAddr + "/" + chainID + "/balance"
}

func keyOfTokenAll() string {
	return "/token/all/0"
}

func keyOfSideChainSupportTokens(sideChainID string) string {
	return "/sidechain/supporttokens/" + sideChainID
}

func keyOfSupportSCList(tokenAddr types.Address) string {
	return "/token/supportsidechains/" + tokenAddr
}
