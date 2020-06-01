package IBC

import (
	"bytes"
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/forx"
	"github.com/bcbchain/sdk/sdk/ibc"
	"github.com/bcbchain/sdk/sdk/types"
)

func (i *Ibc) checkPrecommits(h *ibc.Header_2_2, precommits []ibc.Precommit, hasGovernance bool) bool {
	chainValidators := i._chainValidators(h.ChainID)
	if len(chainValidators) == 0 {
		return false
	}

	var totalPower, verifiedPower int64
	forx.Range(chainValidators, func(key string, validator InfoOfValidator) bool {
		totalPower += validator.Power

		return forx.Continue
	})

	calcHash := i.sdk.Helper().IBCHelper().CalcBlockHash(h)

	forx.Range(precommits, func(index int, precommit ibc.Precommit) bool {
		validator, ok := chainValidators[precommit.ValidatorAddress]
		if !ok {
			return forx.Continue
		}

		if bytes.Compare(calcHash, precommit.BlockID.Hash) != 0 {
			return forx.Continue
		}

		if i.sdk.Helper().IBCHelper().VerifyPrecommit(validator.PubKey, precommit, h.ChainID, h.Height) {
			verifiedPower += validator.Power
		}

		return forx.Continue
	})

	if hasGovernance {
		return verifiedPower >= totalPower*1/2
	}

	return verifiedPower > totalPower*2/3
}

func (i *Ibc) checkQueueHash(queueChain *ibc.QueueChain, packets []ibc.Packet) {
	queueID := packets[0].QueueID
	lastQueueHash := i._lastQueueHash(queueID)

	var queueBlock ibc.QueueBlock
	forx.Range(queueChain.QueueBlocks, func(index int, queueItem ibc.QueueBlock) bool {
		if queueItem.QueueID == queueID {
			queueBlock = queueItem
			return forx.Break
		}

		return forx.Continue
	})

	// check last queue hash
	sdk.Require(lastQueueHash == nil || bytes.Compare(lastQueueHash, queueBlock.LastQueueHash) == 0,
		types.ErrInvalidParameter, "invalid lastQueueHash")

	// check queue hash
	queueHash := i.calcQueueHash(queueID, queueChain, packets)

	sdk.Require(bytes.Compare(queueHash, queueBlock.QueueHash) == 0,
		types.ErrInvalidParameter, "invalid queueHash")

	i._setLastQueueHash(queueID, queueHash)
}

func (i *Ibc) calcQueueHash(queueID string, queueChain *ibc.QueueChain, packets []ibc.Packet) types.Hash {
	lastQueueHash := i._lastQueueHash(queueID)

	// calc queue hash
	var queueHash types.Hash
	lastIndex := len(packets) - 1
	forx.Range(packets, func(index int, packet ibc.Packet) bool {
		itemHash := i.sdk.Helper().IBCHelper().CalcQueueHash(packet, lastQueueHash)
		if index != lastIndex {
			lastQueueHash = itemHash
		} else {
			queueHash = itemHash
		}

		return forx.Continue
	})

	return queueHash
}
