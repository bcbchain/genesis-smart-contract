package governance

import (
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/jsoniter"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
)

///////////////////////////  notify core function  //////////////////////////////
func (g *Governance) newValidatorForNotify(inputReceipt std.Receipt) {
	var receipt newValidator
	err := jsoniter.Unmarshal(inputReceipt.Bytes, &receipt)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	chainID := g.sdk.Helper().BlockChainHelper().GetChainID(receipt.NodeAddr)
	nodes := g._chainValidators(chainID)
	nodes[receipt.NodeAddr] = InfoOfValidator{
		PubKey:     receipt.NodePubKey,
		Power:      receipt.Power,
		RewardAddr: receipt.RewardAddr,
		Name:       receipt.Name,
		NodeAddr:   receipt.NodeAddr,
	}
	g._setChainValidators(chainID, nodes)

}

func (g *Governance) setPowerForNotify(inputReceipt std.Receipt) {
	var receipt setPower
	err := jsoniter.Unmarshal(inputReceipt.Bytes, &receipt)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	chainID := g.sdk.Helper().BlockChainHelper().GetChainID(receipt.NodeAddr)
	nodes := g._chainValidators(chainID)
	if receipt.Power == 0 {
		delete(nodes, receipt.NodeAddr)
	} else {
		validator := nodes[receipt.NodeAddr]
		validator.Power = receipt.Power
		nodes[receipt.NodeAddr] = validator
	}

	g._setChainValidators(chainID, nodes)

}
