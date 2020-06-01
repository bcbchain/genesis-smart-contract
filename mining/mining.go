package mining

import (
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/forx"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
)

//Mining This is struct of contract
//@:contract:mining
//@:version:2.0
//@:organization:orgJgaGConUyK81zibntUBjQ33PKctpk1K1G
//@:author:5e8339cb1a5cce65602fd4f57e115905348f7e83bcbe38dd77694dbe1f8903c9
type Mining struct {
	sdk sdk.ISmartContract
}

//InitChain Constructor of this Mining
//@:constructor
func (m *Mining) InitChain() {

	// update contract mines list
	m.updateContractMines()

	// init mining start height
	startHeight := m.sdk.Block().Height()
	m._setMiningStartHeight(startHeight)

	// init mining reward amount
	rewardAmount := int64(150000000)
	m._setMiningRewardAmount(rewardAmount)
}

//UpdateChain Constructor of this Mining
//@:constructor
func (m *Mining) UpdateChain() {
	//This method is automatically selected when the block height reaches the contract effective block height.

	m.updateContractMines()
}

//mine  from Mining
//@:public:mine
func (m *Mining) Mine() int64 {

	//奖励接收地址
	proposerAddress := m.sdk.Block().ProposerAddress()
	proposerRewardAddress := m.sdk.Block().RewardAddress()
	currentHeight := m.sdk.Block().Height()

	//计算奖励金额
	rewardAmount := m.calcRewardAmount()

	//奖励转账
	token := m.sdk.Helper().GenesisHelper().Token().Address()
	contract := m.sdk.Helper().ContractHelper().ContractOfName("mining")
	contractAcct := contract.Account()
	if contractAcct.BalanceOfToken(token).IsGEI(rewardAmount) {
		contractAcct.TransferByToken(token, proposerRewardAddress, bn.N(rewardAmount))
		m.emitMine(proposerAddress, proposerRewardAddress, currentHeight, rewardAmount)
	} else {
		rewardAmount = 0
	}

	return rewardAmount
}

func (m *Mining) calcRewardAmount() (rewardAmount int64) {
	currentHeight := m.sdk.Block().Height()
	blockNum := currentHeight - m._miningStartHeight()
	rewardAmount = m._miningRewardAmount()

	if blockNum > 0 && blockNum%66000000 == 0 {
		rewardAmount = rewardAmount / 2
		if rewardAmount == 0 {
			rewardAmount = 1
		}
		m._setMiningRewardAmount(rewardAmount)
	}
	return
}

func (m *Mining) emitMine(proposer, rewardAddr types.Address, height int64, rewardValue int64) {
	// Name of Receipt: mining::mine
	type mine struct {
		Proposer    types.Address `json:"proposer"`    // 提案人地址
		RewardAddr  types.Address `json:"rewardAddr"`  // 接收奖励地址
		Height      int64         `json:"height"`      // 挖矿区块高度
		RewardValue int64         `json:"rewardValue"` // 奖励金额(单位：cong)
	}

	m.sdk.Helper().ReceiptHelper().Emit(mine{
		Proposer:    proposer,
		Height:      height,
		RewardAddr:  rewardAddr,
		RewardValue: rewardValue})
}

func (m *Mining) updateContractMines() {
	// init /contract/mines value
	conVerList := m._contractVersionList()
	contractMines := m._contractMines()

	newContractMines := make([]std.MineContract, 0)
	forx.Range(contractMines, func(index int, contractMine std.MineContract) {
		bExist := false
		forx.Range(conVerList.ContractAddrList, func(index int, addr types.Address) bool {
			if contractMine.Address == addr {
				bExist = true
				return forx.Break
			}

			return forx.Continue
		})
		if bExist == false {
			newContractMines = append(newContractMines, contractMine)
		}
	})

	// add now contract
	contract := m.sdk.Helper().ContractHelper().ContractOfName("mining")
	newContractMines = append(newContractMines, std.MineContract{
		MineHeight: m.sdk.Block().Height(),
		Address:    contract.Address()})
	m._setContractMines(newContractMines)
}
