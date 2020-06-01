package IBC

import (
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/ibc"
	"github.com/bcbchain/sdk/sdk/types"
)

//Ibc This is struct of contract
//@:contract:ibc
//@:version:2.3
//@:organization:orgJgaGConUyK81zibntUBjQ33PKctpk1K1G
//@:author:5e8339cb1a5cce65602fd4f57e115905348f7e83bcbe38dd77694dbe1f8903c9
type Ibc struct {
	sdk sdk.ISmartContract
}

//InitChain Constructor of this Ibc
//@:constructor
func (i *Ibc) InitChain() {

}

//@:public:ibc:gas[1500]
func (i *Ibc) Register(toChainID string) {
	sdk.Require(len(toChainID) > 0,
		types.ErrInvalidParameter, "Invalid toChainID")

	localChainID := i.sdk.Helper().GenesisHelper().ChainID()
	sdk.Require(toChainID != localChainID,
		types.ErrInvalidParameter, "toChainID cannot be local chainID")

	queueID := ""
	if i.sdk.Helper().BlockChainHelper().IsSideChain() {
		// side chain
		mainChainID := i.sdk.Helper().BlockChainHelper().GetMainChainID()
		if toChainID != mainChainID {
			sdk.Require(i.isValidSideChainID(toChainID),
				types.ErrInvalidParameter, "Invalid toChainID")
		}

		queueID = i.makeQueueID(localChainID, mainChainID)

	} else {
		// main chain
		chainInfo := i._chainInfo(toChainID)
		sdk.Require(chainInfo.Status == "ready",
			types.ErrInvalidParameter, "chain status must be ready")

		queueID = i.makeQueueID(localChainID, toChainID)
	}

	origins := i.sdk.Message().Origins()
	//sdk.Require(len(origins) == 2, types.ErrInvalidParameter, "invalid origins")
	invokeContract := i.sdk.Helper().ContractHelper().ContractOfAddress(origins[len(origins)-1])

	packet := ibc.Packet{
		FromChainID:  localChainID,
		ToChainID:    toChainID,
		QueueID:      queueID,
		Seq:          i._sequence(queueID) + 1,
		OrgID:        invokeContract.OrgID(),
		ContractName: invokeContract.Name(),
		IbcHash:      i.sdk.Helper().IBCHelper().IbcHash(toChainID),
		Type:         ibc.TccTxType,
		State: ibc.State{
			Status: ibc.NoAckWanted,
			Tag:    ibc.RecastPending,
		},
		Receipts: i.sdk.Message().InputReceipts(),
	}

	i.savePacket(toChainID, nil, &packet)

	i.sdk.Helper().ReceiptHelper().Emit(packet)
}

//@:public:ibc:gas[1000]
func (i *Ibc) Notify(toChainIDs []string) {
	origins := i.sdk.Message().Origins()
	sdk.Require(len(origins) == 2,
		types.ErrInvalidParameter, "invalid origins")

	newChainIDs := i.checkToChainIDs(toChainIDs)
	sdk.Require(len(newChainIDs) > 0,
		types.ErrInvalidParameter, "toChainIDs can not be empty")

	i.notify(newChainIDs)
}

//@:public:ibc:gas[50000]
func (i *Ibc) Broadcast() {
	origins := i.sdk.Message().Origins()
	sdk.Require(len(origins) == 2,
		types.ErrInvalidParameter, "invalid origins")

	toChainIDs := i.filterInvalidSideChain()
	if len(toChainIDs) == 0 {
		return
	}

	i.notify(toChainIDs)
}

//@:public:method:gas[0]
func (i *Ibc) Input(pktsProofs []ibc.PktsProofEx, headers []ibc.Header_2_2) {
	sdk.Require(len(pktsProofs) == len(headers),
		types.ErrInvalidParameter, "")

	validPackets := i.checkPktsProof(pktsProofs, headers)

	i.input(validPackets)
}
