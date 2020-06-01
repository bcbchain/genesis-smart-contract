package IBC

import (
	"github.com/bcbchain/sdk/sdk/forx"
	"github.com/bcbchain/sdk/sdk/ibc"
	"github.com/bcbchain/sdk/sdk/jsoniter"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
	"strings"
)

// 先判断 QueueID 中 toChainID 一定是我，然后再验签，验签之后的数据只允许在 ，log 中保存，交易是成功的。
func (i *Ibc) input(packets []ibc.Packet) {
	var (
		queueID  = packets[0].QueueID
		sequence uint64
	)

	forx.Range(packets, func(index int, packet ibc.Packet) bool {

		i.processPacket(packet)

		i._setQueueIndex(queueID, packet.Seq, MessageIndex{
			Height:  i.sdk.Block().Height(),
			IbcHash: packet.IbcHash,
		})
		sequence = packet.Seq
		return true
	})

	// Saves the sequence number of the source chain to the current chain.
	i._setSequence(queueID, sequence)
}

func (i *Ibc) processPacket(packet ibc.Packet) {

	switch packet.Type {
	case ibc.TccTxType:
		i.processTccTxPacket(packet)

	case ibc.NotifyType:
		i.processNotifyPacket(packet)
	}
}

func (i *Ibc) processTccTxPacket(packet ibc.Packet) {

	if i.isRelay(&packet) {
		switch packet.State.Tag {
		case ibc.RecastPending:
			if err := i.checkSideChain(&packet); err.ErrorCode != types.CodeOK {
				i.processTryRecastResult(&packet, false, nil, err)

			} else {
				ok, outReceipts, err := i.sdk.Helper().IBCStubHelper().TryRecast(packet.IbcHash, packet.OrgID, packet.ContractName, packet.Receipts)
				i.processTryRecastResult(&packet, ok, outReceipts, err)
			}

		case ibc.ConfirmPending:
			outReceipts, err := i.sdk.Helper().IBCStubHelper().ConfirmRecast(packet.IbcHash, packet.OrgID, packet.ContractName, packet.Receipts)
			i.processConfirmRecastResult(&packet, outReceipts, err)

		case ibc.CancelPending:
			outReceipts, err := i.sdk.Helper().IBCStubHelper().CancelRecast(packet.IbcHash, packet.OrgID, packet.ContractName, packet.Receipts)
			i.processCancelRecastResult(&packet, outReceipts, err)
		}

	} else {
		switch packet.State.Tag {
		case ibc.RecastPending:
			if i.sdk.Helper().BlockChainHelper().IsSideChain() {
				ok, outReceipts, err := i.sdk.Helper().IBCStubHelper().Recast(packet.IbcHash, packet.OrgID, packet.ContractName, packet.Receipts)
				i.processRecastResult(&packet, ok, outReceipts, err)

			} else {
				if err := i.checkSideChain(&packet); err.ErrorCode != types.CodeOK {
					i.processRecastResult(&packet, false, nil, err)

				} else {
					ok, outReceipts, err := i.sdk.Helper().IBCStubHelper().Recast(packet.IbcHash, packet.OrgID, packet.ContractName, packet.Receipts)
					i.processRecastResult(&packet, ok, outReceipts, err)
				}
			}

		case ibc.ConfirmPending:
			outReceipts, err := i.sdk.Helper().IBCStubHelper().Confirm(packet.IbcHash, packet.OrgID, packet.ContractName, packet.Receipts)
			i.processConfirmResult(&packet, outReceipts, err)

		case ibc.CancelPending:
			outReceipts, err := i.sdk.Helper().IBCStubHelper().Cancel(packet.IbcHash, packet.OrgID, packet.ContractName, packet.Receipts)
			i.processCancelResult(&packet, outReceipts, err)
		}
	}
}

func (i *Ibc) processNotifyPacket(packet ibc.Packet) {
	if packet.ContractName == "netgovernance" {
		forx.Range(packet.Receipts, func(index int, kvPair types.KVPair) bool {
			var receipt std.Receipt
			_ = jsoniter.Unmarshal(kvPair.Value, &receipt)

			if strings.HasSuffix(string(kvPair.Key), "/netgovernance.setOpenURL") {
				newOpenURLs := new(setOpenURL)
				_ = jsoniter.Unmarshal(receipt.Bytes, newOpenURLs)

				i._setOpenURLs(newOpenURLs.ChainID, newOpenURLs.OpenURLs)

				i.sdk.Helper().ReceiptHelper().Emit(newOpenURLs)

			} else if strings.HasSuffix(string(kvPair.Key), "/netgovernance.setGasPriceRatio") {
				newGasPriceRatio := new(setGasPriceRatio)
				_ = jsoniter.Unmarshal(receipt.Bytes, newGasPriceRatio)

				i._setGasPriceRatio(newGasPriceRatio.GasPriceRatio)

				i.sdk.Helper().ReceiptHelper().Emit(newGasPriceRatio)

			} else if strings.HasSuffix(string(kvPair.Key), "/netgovernance.removeSideChainToken") {
				removeToken := new(removeSideChainToken)
				_ = jsoniter.Unmarshal(receipt.Bytes, removeToken)

				forx.Range(removeToken.TokenAddrs, func(in int, token string) bool {
					token = i.sdk.Helper().BlockChainHelper().RecalcAddressEx(token,
						i.sdk.Helper().BlockChainHelper().GetLocalChainName())
					scIDs := i._supportSideChains(token)
					index := -1
					forx.Range(scIDs, func(i int, scId string) bool {
						if scId == removeToken.ChainID {
							index = i
							return forx.Break
						}
						return forx.Continue
					})

					if index != -1 {
						resultSCIDs := make([]string, 0, len(scIDs)-1)
						resultSCIDs = append(scIDs[:index], scIDs[index+1:]...)
						i._setSupportSideChains(token, resultSCIDs)
					}
					return forx.Continue
				})

				i.sdk.Helper().ReceiptHelper().Emit(removeToken)
			}

			return forx.Continue
		})

	} else {
		outReceipts, err := i.sdk.Helper().IBCStubHelper().Notify(packet.IbcHash, packet.OrgID, packet.ContractName, packet.Receipts)
		i.processNotifyResult(&packet, outReceipts, err)
	}
}
