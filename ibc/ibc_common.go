package IBC

import (
	"errors"
	"fmt"
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/forx"
	"github.com/bcbchain/sdk/sdk/ibc"
	"github.com/bcbchain/sdk/sdk/types"
	"regexp"
	"strings"
)

func (i *Ibc) notify(toChainIDs []string) {
	origins := i.sdk.Message().Origins()

	originContract := i.sdk.Helper().ContractHelper().ContractOfAddress(origins[1])
	localChainID := i.sdk.Helper().GenesisHelper().ChainID()

	packet := ibc.Packet{
		FromChainID:  localChainID,
		OrgID:        originContract.OrgID(),
		ContractName: originContract.Name(),
		Type:         ibc.NotifyType,
		Receipts:     i.sdk.Message().InputReceipts(),
	}

	forx.Range(toChainIDs, func(index int, toChainID string) bool {

		packet.ToChainID = toChainID
		packet.QueueID = i.makeQueueID(localChainID, toChainID)

		packet.Seq = i._sequence(packet.QueueID) + 1
		i._setSequence(packet.QueueID, packet.Seq)

		packet.IbcHash = i.sdk.Helper().IBCHelper().IbcHash(toChainID)

		packet.State = ibc.State{
			Status: ibc.NoAck,
			Tag:    ibc.NotifyPending,
		}
		i._setState(packet.IbcHash, &packet.State)

		packetsOfIbcHash := i._packets(packet.IbcHash)
		packetsOfIbcHash = append(packetsOfIbcHash, packet)
		i._setPackets(packet.IbcHash, packetsOfIbcHash)

		i._setQueueIndex(packet.QueueID, packet.Seq, MessageIndex{
			Height:  i.sdk.Block().Height(),
			IbcHash: packet.IbcHash,
		})

		i.sdk.Helper().ReceiptHelper().Emit(packet)

		return true
	})
}

func (i *Ibc) checkToChainIDs(toChainIDs []string) []string {
	localChainID := i.sdk.Block().ChainID()
	newToChainIDs := make([]string, 0)
	mainChainID := i.sdk.Helper().BlockChainHelper().GetMainChainID()

	forx.Range(toChainIDs, func(index int, toChainID string) bool {
		if len(toChainID) == 0 || toChainID == localChainID {
			return forx.Continue
		}

		if i.sdk.Helper().BlockChainHelper().IsSideChain() {
			if toChainID == mainChainID &&
				!inSlice(newToChainIDs, toChainID) {

				newToChainIDs = append(newToChainIDs, toChainID)
				return forx.Continue
			}

		} else {
			chainInfo := i._chainInfo(toChainID)
			if (chainInfo.Status == "ready" || chainInfo.Status == "clear") &&
				!inSlice(newToChainIDs, toChainID) {

				newToChainIDs = append(newToChainIDs, toChainID)
				return forx.Continue
			}
		}

		return forx.Continue
	})

	return newToChainIDs
}

func inSlice(slice []string, item string) bool {
	exist := false
	forx.Range(slice, func(index int, s string) bool {
		if item == s {
			exist = true
			return forx.Break
		}
		return forx.Continue
	})
	return exist
}

func (i *Ibc) filterInvalidSideChain() []string {
	allSideChainIDs := i._sideChainIDs()

	return i.checkToChainIDs(allSideChainIDs)
}

func (i *Ibc) checkSideChain(packet *ibc.Packet) types.Error {
	err := types.Error{
		ErrorCode: types.CodeOK,
	}

	if i._chainInfo(packet.FromChainID).Status == "disabled" {
		err.ErrorCode = types.ErrInvalidParameter
		err.ErrorDesc = "From chain status can not be disabled"
		return err
	}

	if packet.ToChainID != i.sdk.Block().ChainID() && i._chainInfo(packet.ToChainID).Status != "ready" {
		err.ErrorCode = types.ErrInvalidParameter
		err.ErrorDesc = "To chain status must be ready"
		return err
	}
	return err
}

func (i *Ibc) processNotifyResult(packet *ibc.Packet, outReceipts []types.KVPair, notifyErr types.Error) {
	newPacket := i.newPacket(packet)
	if notifyErr.ErrorCode != types.CodeOK {
		newPacket.State.Tag = ibc.NotifyFailure
		newPacket.State.Log = notifyErr.ErrorDesc

	} else {
		newPacket.State.Tag = ibc.NotifySuccess
	}

	newPacket.State.Status = ibc.NoAck
	if len(outReceipts) > 0 {
		newPacket.Receipts = append(packet.Receipts, outReceipts[1:]...)
	} else {
		newPacket.Receipts = packet.Receipts
	}

	i._setState(newPacket.IbcHash, &newPacket.State)
	packetsOfIbcHash := i._packets(newPacket.IbcHash)
	packetsOfIbcHash = append(packetsOfIbcHash, *newPacket)
	i._setPackets(newPacket.IbcHash, packetsOfIbcHash)
}

func (i *Ibc) processCancelRecastResult(packet *ibc.Packet, outReceipts []types.KVPair, cancelHubErr types.Error) {
	newPacket := i.newPacket(packet)
	if cancelHubErr.ErrorCode != types.CodeOK {
		newPacket.State.Log = cancelHubErr.ErrorDesc
	}
	newPacket.State.Status = ibc.NoAck
	newPacket.State.Tag = packet.State.Tag

	if len(outReceipts) > 0 {
		newPacket.Receipts = append(packet.Receipts, outReceipts[1:]...)
	} else {
		newPacket.Receipts = packet.Receipts
	}

	newPacket.QueueID = i.makeQueueID(i.sdk.Helper().GenesisHelper().ChainID(), newPacket.FromChainID)

	newPacket.Seq = i._sequence(newPacket.QueueID) + 1
	i.savePacket(newPacket.FromChainID, packet, newPacket)

	i.sdk.Helper().ReceiptHelper().Emit(newPacket)
}

func (i *Ibc) processCancelResult(packet *ibc.Packet, outReceipts []types.KVPair, cancelErr types.Error) {
	newPacket := i.newPacket(packet)
	if cancelErr.ErrorCode != types.CodeOK {
		newPacket.State.Log = cancelErr.ErrorDesc
	}
	newPacket.State.Status = ibc.NoAck
	newPacket.State.Tag = ibc.Canceled

	if len(outReceipts) > 0 {
		newPacket.Receipts = append(packet.Receipts, outReceipts[1:]...)
	} else {
		newPacket.Receipts = packet.Receipts
	}

	packetsOfIbcHash := i._packets(newPacket.IbcHash)
	packetsOfIbcHash = append(packetsOfIbcHash, *packet)
	packetsOfIbcHash = append(packetsOfIbcHash, *newPacket)
	i._setPackets(newPacket.IbcHash, packetsOfIbcHash)
	i._setState(newPacket.IbcHash, &newPacket.State)

	final := ibc.Final{
		IBCHash: newPacket.IbcHash,
		State: ibc.State{
			Status: ibc.NoAck,
			Tag:    ibc.Canceled,
			Log:    cancelErr.ErrorDesc,
		},
	}
	i.sdk.Helper().ReceiptHelper().Emit(final)
}

func (i *Ibc) processConfirmRecastResult(packet *ibc.Packet, outReceipts []types.KVPair, confirmHubErr types.Error) {
	newPacket := i.newPacket(packet)
	if confirmHubErr.ErrorCode != types.CodeOK {
		newPacket.State.Log = confirmHubErr.ErrorDesc
	}
	newPacket.State.Status = ibc.NoAck
	newPacket.State.Tag = packet.State.Tag

	if len(outReceipts) > 0 {
		newPacket.Receipts = append(packet.Receipts, outReceipts[1:]...)
	} else {
		newPacket.Receipts = packet.Receipts
	}

	newPacket.QueueID = i.makeQueueID(i.sdk.Helper().GenesisHelper().ChainID(), packet.FromChainID)

	newPacket.Seq = i._sequence(newPacket.QueueID) + 1
	i.savePacket(newPacket.FromChainID, packet, newPacket)

	i.sdk.Helper().ReceiptHelper().Emit(newPacket)
}

func (i *Ibc) processConfirmResult(packet *ibc.Packet, outReceipts []types.KVPair, confirmErr types.Error) {
	newPacket := i.newPacket(packet)
	if confirmErr.ErrorCode != types.CodeOK {
		newPacket.State.Log = confirmErr.ErrorDesc
	}

	newPacket.State.Tag = ibc.Confirmed
	newPacket.State.Status = ibc.NoAck
	if len(outReceipts) > 0 {
		newPacket.Receipts = append(packet.Receipts, outReceipts[1:]...)
	} else {
		newPacket.Receipts = packet.Receipts
	}

	packetsOfIbcHash := i._packets(newPacket.IbcHash)
	packetsOfIbcHash = append(packetsOfIbcHash, *packet)
	packetsOfIbcHash = append(packetsOfIbcHash, *newPacket)
	i._setPackets(newPacket.IbcHash, packetsOfIbcHash)
	i._setState(newPacket.IbcHash, &newPacket.State)

	final := ibc.Final{
		IBCHash: packet.IbcHash,
		State: ibc.State{
			Status: ibc.NoAck,
			Tag:    ibc.Confirmed,
			Log:    confirmErr.ErrorDesc,
		},
	}
	i.sdk.Helper().ReceiptHelper().Emit(final)
}

func (i *Ibc) isRelay(packet *ibc.Packet) bool {
	return !i.sdk.Helper().BlockChainHelper().IsSideChain() &&
		i.sdk.Helper().GenesisHelper().ChainID() != packet.FromChainID &&
		i.sdk.Helper().GenesisHelper().ChainID() != packet.ToChainID
}

func (i *Ibc) processTryRecastResult(packet *ibc.Packet, recastOk bool, outReceipts []types.KVPair, tryHubErr types.Error) {
	newPacket := i.newPacket(packet)
	toChainID := ""
	if tryHubErr.ErrorCode != types.CodeOK {
		toChainID = packet.FromChainID
		newPacket.State.Status = ibc.NoAck
		newPacket.State.Tag = ibc.CancelPending
		newPacket.State.Log = tryHubErr.ErrorDesc

	} else if !recastOk {
		toChainID = packet.FromChainID
		newPacket.State.Status = ibc.NoAck
		newPacket.State.Tag = ibc.CancelPending

	} else {
		toChainID = packet.ToChainID
		newPacket.State.Status = ibc.NoAckWanted
		newPacket.State.Tag = ibc.RecastPending
	}

	newPacket.QueueID = i.makeQueueID(i.sdk.Helper().GenesisHelper().ChainID(), toChainID)
	newPacket.Receipts = packet.Receipts

	seq := i._sequence(newPacket.QueueID) + 1
	newPacket.Seq = seq
	i.savePacket(toChainID, packet, newPacket)

	i.sdk.Helper().ReceiptHelper().Emit(newPacket)
}

func (i *Ibc) processRecastResult(packet *ibc.Packet, recastOk bool, outReceipts []types.KVPair, recastErr types.Error) {
	newPacket := i.newPacket(packet)
	if recastErr.ErrorCode != types.CodeOK {
		newPacket.State.Tag = ibc.CancelPending
		newPacket.State.Log = recastErr.ErrorDesc

	} else if !recastOk {
		newPacket.State.Tag = ibc.CancelPending

	} else {
		newPacket.State.Tag = ibc.ConfirmPending
	}

	if len(outReceipts) > 0 {
		newPacket.Receipts = append(packet.Receipts, outReceipts[1:]...)
	} else {
		newPacket.Receipts = packet.Receipts
	}

	toChainID := ""
	if i.sdk.Helper().BlockChainHelper().IsSideChain() {
		toChainID = i.sdk.Helper().BlockChainHelper().GetMainChainID()
	} else {
		toChainID = packet.FromChainID
	}
	newPacket.State.Status = ibc.NoAck

	newPacket.QueueID = i.makeQueueID(i.sdk.Helper().GenesisHelper().ChainID(), toChainID)
	seq := i._sequence(i.makeQueueID(i.sdk.Helper().GenesisHelper().ChainID(), toChainID)) + 1
	newPacket.Seq = seq
	i.savePacket(toChainID, packet, newPacket)
	i.sdk.Helper().ReceiptHelper().Emit(*newPacket)
}

func (i *Ibc) savePacket(toChainID string, oldPacket *ibc.Packet, newPacket *ibc.Packet) {
	packetsOfIbcHash := i._packets(newPacket.IbcHash)
	if oldPacket != nil {
		packetsOfIbcHash = append(packetsOfIbcHash, *oldPacket)
	}
	if newPacket != nil {
		packetsOfIbcHash = append(packetsOfIbcHash, *newPacket)
	}

	i._setPackets(newPacket.IbcHash, packetsOfIbcHash)
	i._setState(newPacket.IbcHash, &newPacket.State)
	i._setSequence(newPacket.QueueID, newPacket.Seq)
	i._setQueueIndex(newPacket.QueueID, newPacket.Seq, MessageIndex{
		Height:  i.sdk.Block().Height(),
		IbcHash: newPacket.IbcHash,
	})
}

func (i *Ibc) isProcessedPacket(packet *ibc.Packet) bool {
	currentSeq := i._sequence(packet.QueueID)

	if packet.Seq > currentSeq {
		return false
	} else {
		packets := i._packets(packet.IbcHash)
		sdk.Require(len(packets) > 0,
			types.ErrInvalidParameter, "invalid packets seq")
		return true
	}
}

func (i *Ibc) makeQueueID(fromChainID, toChainID string) string {
	var sb strings.Builder
	sb.WriteString(fromChainID)
	sb.WriteString("->")
	sb.WriteString(toChainID)
	return sb.String()
}

func (i *Ibc) splitQueue(queueID string) (fromChainID, toChainID string, err error) {
	result := strings.Split(queueID, "->")
	if len(result) != 2 {
		return "", "", errors.New("Invalid queue: " + queueID)
	}
	return result[0], result[1], nil
}

func (i *Ibc) newPacket(oldPacket *ibc.Packet) *ibc.Packet {
	newPacket := ibc.Packet{
		FromChainID:  oldPacket.FromChainID,
		ToChainID:    oldPacket.ToChainID,
		OrgID:        oldPacket.OrgID,
		ContractName: oldPacket.ContractName,
		IbcHash:      oldPacket.IbcHash,
		Type:         oldPacket.Type,
	}
	return &newPacket
}

func (i *Ibc) hasGovernanceNotify(packets []ibc.Packet) bool {
	result := false
	forx.Range(packets, func(index int, packet ibc.Packet) bool {
		if packet.ContractName == "governance" {
			result = true
			return forx.Break
		}
		return forx.Continue
	})
	return result
}

// checkPktsProof check packets and proof data
func (i *Ibc) checkPktsProof(pktsProofs []ibc.PktsProofEx, headers []ibc.Header_2_2) []ibc.Packet {
	sdk.Require(len(pktsProofs) > 0 && len(pktsProofs) <= 10,
		types.ErrInvalidParameter, "invlaid pktsProofs length")

	validPackets := make([]ibc.Packet, 0)

	var lastSeq uint64
	forx.Range(pktsProofs, func(index int, pktsProof ibc.PktsProofEx) bool {

		if len(validPackets) != 0 {
			lastSeq = validPackets[len(validPackets)-1].Seq
		}
		newPackets := i.checkPackets(pktsProof.Packets, lastSeq)

		if len(newPackets) == 0 {
			return forx.Continue
		}

		sdk.Require(i.checkPrecommits(&headers[index], pktsProof.Precommits, i.hasGovernanceNotify(newPackets)),
			types.ErrInvalidParameter, "checkPrecommits failed")

		i.checkQueueHash(headers[index].LastQueueChains, newPackets)

		validPackets = append(validPackets, newPackets...)

		return forx.Continue
	})

	sdk.Require(len(validPackets) > 0,
		types.ErrInvalidParameter, "no valid packet")

	return validPackets
}

// checkPacket check packets data
func (i *Ibc) checkPackets(packets []ibc.Packet, lastSeq uint64) []ibc.Packet {

	newPackets := make([]ibc.Packet, 0)
	var localSeq uint64
	forx.Range(packets, func(index int, packet ibc.Packet) bool {
		i.checkQueueID(packet.QueueID)

		if localSeq == 0 {
			localSeq = i._sequence(packet.QueueID)
		}
		if packet.Seq <= localSeq {
			return forx.Continue // 已处理过的packet，跳过
		}
		if lastSeq == 0 {
			lastSeq = localSeq
		}
		sdk.Require(packet.Seq == lastSeq+1,
			types.ErrInvalidParameter, fmt.Sprintf("invalid packet: expected seq %d, obtain %d", localSeq+1, packet.Seq))

		i.checkTypeState(packet.Type, packet.State)

		i.checkFromToChainID(packet.State.Tag, packet.FromChainID, packet.ToChainID)

		sdk.Require(len(packet.IbcHash) == 32,
			types.ErrInvalidParameter, "invalid ibcHash")

		newPackets = append(newPackets, packet)
		if len(newPackets) == 1 {
			lastSeq = newPackets[0].Seq
		} else {
			lastSeq += 1
		}

		return forx.Continue
	})

	return newPackets
}

// checkTypeState check packet's type and state, it's type must be TccTxType or NotifyType;
// if type is TccTxType then state's tag must be RecastPending/ConfirmPending/CancelPending;
// if type is NotifyType then state's tag must be NotifyPending.
func (i *Ibc) checkTypeState(packetType string, state ibc.State) {
	sdk.Require(state.Status == ibc.NoAck || state.Status == ibc.NoAckWanted,
		types.ErrInvalidParameter, "invalid packet's state status")

	switch packetType {
	case ibc.TccTxType:
		sdk.Require(
			state.Tag == ibc.RecastPending ||
				state.Tag == ibc.ConfirmPending ||
				state.Tag == ibc.CancelPending,
			types.ErrInvalidParameter, "invalid packet's state tag")
	case ibc.NotifyType:
		sdk.Require(state.Tag == ibc.NotifyPending,
			types.ErrInvalidParameter, "invalid packet's state tag")
	default:
		sdk.Require(false,
			types.ErrInvalidParameter, "invalid packet's type")
	}
}

// checkQueueID check queueID, it must split by `->`, and must be match next rule
// 1, first chainID must not equal second chainID,
// 2, in side chain, first chainID must be mainChainID, in main chain, first chainID must be valid sideChainID
// 3, second chainID must be localChainID,
func (i *Ibc) checkQueueID(queueID string) {
	errMsg := "invalid queueID"
	queueIDSplit := strings.Split(queueID, "->")
	sdk.Require(len(queueIDSplit) == 2,
		types.ErrInvalidParameter, errMsg)

	localChainID := i.sdk.Block().ChainID()
	queueFromChainID := queueIDSplit[0]
	queueToChainID := queueIDSplit[1]

	sdk.Require(queueToChainID == localChainID,
		types.ErrInvalidParameter, errMsg)

	sdk.Require(queueFromChainID != queueToChainID,
		types.ErrInvalidParameter, errMsg)

	if i.sdk.Helper().BlockChainHelper().IsSideChain() {
		sdk.Require(i.isMainChainID(queueFromChainID),
			types.ErrInvalidParameter, errMsg)
	} else {
		sdk.Require(i.isValidSideChainID(queueFromChainID),
			types.ErrInvalidParameter, errMsg)
	}
}

// checkFromToChainID check fromChainID and toChainID, fromChainID cannot equal toChainID,
// and they must match different rule with different tag value
func (i *Ibc) checkFromToChainID(tag string, fromChainID, toChainID string) {
	localChainID := i.sdk.Block().ChainID()

	// fromChainID 必须不能等于 toChainID
	sdk.Require(fromChainID != toChainID,
		types.ErrInvalidParameter, "fromChainID cannot equal to toChainID")
	switch tag {
	case ibc.RecastPending:
		if i.sdk.Helper().BlockChainHelper().IsSideChain() {
			// 当前链是侧链，toChainID必须为当前链ID，同时fromChainID必须是有效的链ID
			sdk.Require(toChainID == localChainID,
				types.ErrInvalidParameter, "invalid to chainID")

			sdk.Require(i.isValidChainID(fromChainID),
				types.ErrInvalidParameter, "invalid from chainID")
		} else {
			// 当前链是主链，toChainID必须为有效链ID，同时fromChainID必须是有效的侧链ID
			sdk.Require(i.isValidChainID(toChainID),
				types.ErrInvalidParameter, "invalid to chainID")

			sdk.Require(i.isValidSideChainID(fromChainID),
				types.ErrInvalidParameter, "invalid from chainID")
		}
	case ibc.ConfirmPending, ibc.CancelPending:
		if i.sdk.Helper().BlockChainHelper().IsSideChain() {
			// 当前链是侧链，fromChainID必须为当前链ID，同时toChainID必须是有效的链ID
			sdk.Require(fromChainID == localChainID,
				types.ErrInvalidParameter, "invalid from chainID")

			sdk.Require(i.isValidChainID(toChainID),
				types.ErrInvalidParameter, "invalid to chainID")
		} else {
			// 当前链是主链，fromChainID必须为有效链ID，同时toChainID必须是有效的侧链ID
			sdk.Require(i.isValidChainID(fromChainID),
				types.ErrInvalidParameter, "invalid from chainID")

			sdk.Require(i.isValidSideChainID(toChainID),
				types.ErrInvalidParameter, "invalid to chainID")
		}
	case ibc.NotifyPending:
		sdk.Require(toChainID == localChainID,
			types.ErrInvalidParameter, "invalid to chainID")
		if i.sdk.Helper().BlockChainHelper().IsSideChain() {
			// 当前链是侧链，fromChainID必须是主链ID
			sdk.Require(i.isMainChainID(fromChainID),
				types.ErrInvalidParameter, "invalid from chainID")
		} else {
			// 当前链是主链，fromChainID必须是有效侧链ID
			sdk.Require(i.isValidSideChainID(fromChainID),
				types.ErrInvalidParameter, "invalid from chainID")
		}
	default:
		sdk.Require(false,
			types.ErrInvalidParameter, "invalid packet's state tag")
	}
}

// isValidChainID if chainID is valid chainID return true, else return false
func (i *Ibc) isValidChainID(chainID string) bool {

	return i.isMainChainID(chainID) || i.isValidSideChainID(chainID)
}

// isMainChainID if chainID is main chain id return true, else return false
func (i *Ibc) isMainChainID(chainID string) bool {
	mainChainID := i.sdk.Helper().BlockChainHelper().GetMainChainID()

	return mainChainID == chainID
}

// isValidSideChainID if chainID is valid side chain return true, else return false
func (i *Ibc) isValidSideChainID(chainID string) bool {
	mainChainID := i.sdk.Helper().BlockChainHelper().GetMainChainID()
	pattern := "^" + mainChainID + `\[[A-Za-z][a-zA-Z0-9_]{1,39}\]$`

	r, _ := regexp.Compile(pattern)
	return r.MatchString(chainID)
}
