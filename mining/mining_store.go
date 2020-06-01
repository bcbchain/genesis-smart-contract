package mining

import (
	"github.com/bcbchain/sdk/sdk/std"
)

func keyOfStartHeight() string  { return "/mining/start/height" }
func keyOfRewardAmount() string { return "/mining/reward/amount" }

//_miningStore This is a get mining start height method of Mining
func (m *Mining) _miningStartHeight() int64 {
	return *m.sdk.Helper().StateHelper().McGetEx(keyOfStartHeight(), new(int64)).(*int64)
}

//_setSampleStore This is a set mining start height method of Mining
func (m *Mining) _setMiningStartHeight(v int64) {
	m.sdk.Helper().StateHelper().McSet(keyOfStartHeight(), &v)
}

//_chkMiningStore This is a check mining start height method of Mining
func (m *Mining) _chkMiningStartHeight() bool {
	return m.sdk.Helper().StateHelper().McCheck(keyOfStartHeight())
}

//_delMiningStore This is a delete mining start height method of Mining
func (m *Mining) _delMiningStartHeight() {
	m.sdk.Helper().StateHelper().McDelete(keyOfStartHeight())
}

//_setMiningRewardAmount This is a set mining reward amount method of Mining
func (m *Mining) _setMiningRewardAmount(v int64) {
	m.sdk.Helper().StateHelper().McSet(keyOfRewardAmount(), &v)
}

//_MiningRewardAmount This is a get mining reward amount method of Mining
func (m *Mining) _miningRewardAmount() int64 {
	return *m.sdk.Helper().StateHelper().McGetEx(keyOfRewardAmount(), new(int64)).(*int64)
}

//_chkMiningRewardAmount This is a check mining reward amount method of Mining
func (m *Mining) _chkMiningRewardAmount() bool {
	return m.sdk.Helper().StateHelper().McCheck(keyOfRewardAmount())
}

//_delMiningRewardAmount This is a delete mining reward amount method of Mining
func (m *Mining) _delMiningRewardAmount() {
	m.sdk.Helper().StateHelper().McDelete(keyOfRewardAmount())
}

//_contractMines This is a get contract mines method
func (m *Mining) _contractMines() []std.MineContract {
	return *m.sdk.Helper().StateHelper().GetEx(std.KeyOfMineContracts(), new([]std.MineContract)).(*[]std.MineContract)
}

//_setContractMines This is a get contract mines method
func (m *Mining) _setContractMines(v []std.MineContract) {
	m.sdk.Helper().StateHelper().Set(std.KeyOfMineContracts(), v)
}

//_contractVersionList This is a get contract version list method
func (m *Mining) _contractVersionList() *std.ContractVersionList {

	contract := m.sdk.Helper().ContractHelper().ContractOfName("mining")
	key := std.KeyOfContractsWithName(contract.OrgID(), contract.Name())

	return m.sdk.Helper().StateHelper().GetEx(key, new(std.ContractVersionList)).(*std.ContractVersionList)
}
