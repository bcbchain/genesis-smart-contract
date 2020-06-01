package governance

import "github.com/bcbchain/sdk/sdk/types"

//InfoOfValidator validator struct
type InfoOfValidator struct {
	PubKey     types.PubKey `json:"nodepubkey,omitempty"` //节点公钥
	Power      int64        `json:"power,omitempty"`      //节点记账权重
	RewardAddr string       `json:"rewardaddr,omitempty"` //节点接收奖励的地址
	Name       string       `json:"name,omitempty"`       //节点名称
	NodeAddr   string       `json:"nodeaddr,omitempty"`   //节点地址
}

//RewardStrategy reward strategy
type RewardStrategy struct {
	Strategy     []Reward `json:"rewardStrategy,omitempty"` //奖励策略
	EffectHeight int64    `json:"effectHeight,omitempty"`   //生效高度
}

//Reward reward name percent and address
type Reward struct {
	Name          string `json:"name"`          // 被奖励者名称
	RewardPercent string `json:"rewardPercent"` // 奖励比例
	Address       string `json:"address"`       // 被奖励者地址
}

type newValidator struct {
	Name       string        `json:"name,omitempty"`       //节点名称
	NodePubKey types.PubKey  `json:"nodePubKey,omitempty"` //节点公钥
	NodeAddr   types.Address `json:"nodeAddr,omitempty"`   //节点地址
	RewardAddr types.Address `json:"rewardAddr,omitempty"` //节点接收奖励的地址
	Power      int64         `json:"power,omitempty"`      //节点记账权重
}

type setPower struct {
	Name       string        `json:"name,omitempty"`       //节点名称
	NodePubKey types.PubKey  `json:"nodePubKey,omitempty"` //节点公钥
	NodeAddr   types.Address `json:"nodeAddr,omitempty"`   //节点地址
	RewardAddr types.Address `json:"rewardAddr,omitempty"` //节点接收奖励的地址
	Power      int64         `json:"power,omitempty"`      //节点记账权重
}

type setRewardAddr struct {
	Name       string        `json:"name,omitempty"`       //节点名称
	NodePubKey types.PubKey  `json:"nodePubKey,omitempty"` //节点公钥
	NodeAddr   types.Address `json:"nodeAddr,omitempty"`   //节点地址
	RewardAddr types.Address `json:"rewardAddr,omitempty"` //节点接收奖励的地址
	Power      int64         `json:"power,omitempty"`      //节点记账权重
}

type setRewardStrategy struct {
	Strategy     []Reward `json:"rewardStrategy,omitempty"` //奖励策略
	EffectHeight int64    `json:"effectHeight,omitempty"`   //生效高度
}

type SetConfig struct {
	CreateEmptyBlocks         bool `json:"create_empty_blocks"`          // 是否出空块
	ForceIntervalBlockSwitch  bool `json:"force_interval_block_switch"`  // 出块间隔开关
	CreateEmptyBlocksInterval int  `json:"create_empty_blocks_interval"` // 出块时间间隔(millisecond)
}

type BVMStatus struct {
	Enable bool `json:"enable"`
}
