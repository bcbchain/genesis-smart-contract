package governance

import (
	"github.com/bcbchain/sdk/sdk/types"
)

func (g *Governance) emitNewValidator(
	name string,
	nodePubKey types.PubKey,
	nodeAddr types.Address,
	rewardAddr types.Address,
	power int64) {

	receipt := newValidator{
		Name:       name,
		NodePubKey: nodePubKey,
		NodeAddr:   nodeAddr,
		RewardAddr: rewardAddr,
		Power:      power,
	}

	g.sdk.Helper().ReceiptHelper().Emit(receipt)
}

func (g *Governance) emitSetPower(
	name string,
	nodePubKey types.PubKey,
	nodeAddr types.Address,
	rewardAddr types.Address,
	power int64) {

	receipt := setPower{
		Name:       name,
		NodePubKey: nodePubKey,
		NodeAddr:   nodeAddr,
		RewardAddr: rewardAddr,
		Power:      power,
	}

	g.sdk.Helper().ReceiptHelper().Emit(receipt)
}

func (g *Governance) emitSetRewardAddr(
	name string,
	nodePubKey types.PubKey,
	nodeAddr types.Address,
	rewardAddr types.Address,
	power int64) {

	receipt := setRewardAddr{
		Name:       name,
		NodePubKey: nodePubKey,
		NodeAddr:   nodeAddr,
		RewardAddr: rewardAddr,
		Power:      power,
	}

	g.sdk.Helper().ReceiptHelper().Emit(receipt)
}

func (g *Governance) emitSetRewardStrategy(rewards []Reward, effectHeight int64) {

	receipt := setRewardStrategy{
		Strategy:     rewards,
		EffectHeight: effectHeight,
	}

	g.sdk.Helper().ReceiptHelper().Emit(receipt)
}

func (g *Governance) emitSetConfig(createEmptyBlock bool, enable bool, interval int) {

	receipt := SetConfig{
		CreateEmptyBlocks:         createEmptyBlock,
		ForceIntervalBlockSwitch:  enable,
		CreateEmptyBlocksInterval: interval,
	}
	g.sdk.Helper().ReceiptHelper().Emit(receipt)
}

func (g *Governance) emitSetBVMStatus(enable bool) {

	receipt := BVMStatus{
		Enable: enable,
	}

	g.sdk.Helper().ReceiptHelper().Emit(receipt)
}
