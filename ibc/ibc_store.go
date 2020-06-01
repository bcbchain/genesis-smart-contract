package IBC

import (
	"github.com/bcbchain/sdk/sdk/ibc"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
)

func (i *Ibc) _sequence(queueID string) uint64 {
	return i.sdk.Helper().StateHelper().GetUint64(keyOfSequence(queueID))
}

func (i *Ibc) _setSequence(queueID string, value uint64) {
	i.sdk.Helper().StateHelper().SetUint64(keyOfSequence(queueID), value)
}

func (i *Ibc) _packets(ibcHash types.Hash) []ibc.Packet {
	return *i.sdk.Helper().StateHelper().GetEx(keyOfPackets(ibcHash), new([]ibc.Packet)).(*[]ibc.Packet)
}

func (i *Ibc) _setPackets(ibcHash types.Hash, packets []ibc.Packet) {
	i.sdk.Helper().StateHelper().Set(keyOfPackets(ibcHash), &packets)
}

func (i *Ibc) _chainInfo(chainID string) ChainInfo {
	return *i.sdk.Helper().StateHelper().GetEx(keyOfChainInfo(chainID), new(ChainInfo)).(*ChainInfo)
}

func (i *Ibc) _sideChainIDs() []string {
	return *i.sdk.Helper().StateHelper().GetEx(keyOfSideChainIDs(), new([]string)).(*[]string)
}

func (i *Ibc) _setState(ibcHash types.Hash, state *ibc.State) {
	i.sdk.Helper().StateHelper().Set(keyOfState(ibcHash), state)
}

func (i *Ibc) _setQueueIndex(queueID string, seq uint64, msgIndex MessageIndex) {
	i.sdk.Helper().StateHelper().Set(keyOfSequenceHeight(queueID, seq), &msgIndex)
}

// --------------- proof ------------------------
func (i *Ibc) _lastQueueHash(queueID string) types.Hash {
	return *i.sdk.Helper().StateHelper().GetEx(keyOfLastQueueHash(queueID), new(types.Hash)).(*types.Hash)
}

func (i *Ibc) _setLastQueueHash(queueID string, lastQueueHash types.Hash) {
	i.sdk.Helper().StateHelper().Set(keyOfLastQueueHash(queueID), lastQueueHash)
}

func (i *Ibc) _chainValidators(chainID string) (chainValidators map[string]InfoOfValidator) {
	return *i.sdk.Helper().StateHelper().GetEx("/ibc/"+chainID, &chainValidators).(*map[string]InfoOfValidator)
}

// --------------- openUrls ------------------------
func (i *Ibc) _setOpenURLs(chainID string, urls []string) {
	i.sdk.Helper().StateHelper().Set(keyOfSetOpenURLs(chainID), &urls)
}

// --------------- gasPriceRatio ------------------------
func (i *Ibc) _setGasPriceRatio(gasPriceRatio string) {
	i.sdk.Helper().StateHelper().Set(std.KeyOfGasPriceRatio(), gasPriceRatio)
}

func (i *Ibc) _supportSideChains(tokenAddr types.Address) []string {
	return *i.sdk.Helper().StateHelper().GetEx(std.KeyOfSupportSideChains(tokenAddr), new([]string)).(*[]string)
}

func (i *Ibc) _setSupportSideChains(tokenAddr types.Address, sideChainIDs []string) {
	i.sdk.Helper().StateHelper().Set(std.KeyOfSupportSideChains(tokenAddr), &sideChainIDs)
}
