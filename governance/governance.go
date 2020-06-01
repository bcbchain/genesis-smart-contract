package governance

import (
	"github.com/bcbchain/sdk/sdk/forx"
	"strconv"
	"strings"

	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/jsoniter"
	"github.com/bcbchain/sdk/sdk/types"
)

//Governance manage validators
//@:contract:governance
//@:version:2.5
//@:organization:orgJgaGConUyK81zibntUBjQ33PKctpk1K1G
//@:author:5e8339cb1a5cce65602fd4f57e115905348f7e83bcbe38dd77694dbe1f8903c9
type Governance struct {
	sdk sdk.ISmartContract
}

//InitChain Constructor of the contract
//@:constructor
func (g *Governance) InitChain() {

}

// UpdateChain - Update contract
//@:constructor
func (g *Governance) UpdateChain() {
	nodeAddrs := g._validators()
	mainChainID := g.sdk.Helper().BlockChainHelper().GetMainChainID()
	nodes := g._chainValidators(mainChainID)

	// Nodes need to delete
	dipper := "bcb5NkjpneYYRNGhALTCVkTPkrNvZVLPxLKS"
	neptune := "bcbAvVnBf7pmMMQ8zJuFUmqhXoTy4M49DUcJ"
	pluto := "bcbP2gPSpJkm4RXevJdAXr2cdoXbiZdRg9t7"

	if !g.sdk.Helper().BlockChainHelper().IsSideChain() {
		forx.Range(nodeAddrs, func(i int, nodeAddr string) {
			if nodeAddr == dipper || nodeAddr == neptune || nodeAddr == pluto {
				// update validator power = 0
				v := g._validator(nodeAddr)
				v.Power = 0
				g._setValidator(v)

				// delete
				delete(nodes, nodeAddr)
			}
		})

	} else {
		forx.Range(nodes, func(nodeAddr string, validator InfoOfValidator) {
			if nodeAddr == dipper ||
				nodeAddr == neptune ||
				nodeAddr == pluto {
				delete(nodes, nodeAddr)
			}
		})
	}

	// delete chainValidator
	g._setChainValidators(mainChainID, nodes)
}

const (
	//PubKeyLen is public key length
	PubKeyLen = 32
	//MaxNameLen is the max length of validator name
	MaxNameLen = 40
)

//NewValidator add a new validator
//@:public:method:gas[50000]
func (g *Governance) NewValidator(
	name string,
	pubKey types.PubKey,
	rewardAddr types.Address,
	power int64) {

	sdk.RequireOwner()
	sdk.Require(len(name) != 0 && len(name) <= MaxNameLen,
		types.ErrInvalidParameter, "Invalid name.")
	sdk.Require(len(pubKey) == PubKeyLen,
		types.ErrInvalidParameter, "Invalid pubKey.")
	sdk.RequireAddress(rewardAddr)
	sdk.Require(power > 0,
		types.ErrInvalidParameter, "Invalid power.")

	nodeAddr := g.sdk.Helper().BlockChainHelper().CalcAccountFromPubKey(pubKey)
	if g._chkValidator(nodeAddr) == true {
		sdk.Require(g._validator(nodeAddr).Power == 0,
			types.ErrInvalidParameter, "Validator is already exist.")
	}

	allValidators, _ := g.getAllValidators()
	if len(allValidators) >= 3 {
		g.checkPower(nodeAddr, power, allValidators)
	} else {
		sdk.Require(power == allValidators[0].Power,
			types.ErrInvalidParameter, "Power should be equal with other validator when the number of the validator less than three")
	}

	forx.Range(allValidators, func(i int, oldValidator InfoOfValidator) bool {
		sdk.Require(oldValidator.Name != name,
			types.ErrInvalidParameter, "Validator's name is already exist.")
		return true
	})

	newValidator := InfoOfValidator{
		Name:       name,
		PubKey:     pubKey.Bytes(),
		NodeAddr:   nodeAddr,
		RewardAddr: rewardAddr,
		Power:      power,
	}
	g._setValidator(newValidator)

	// 保存所有验证者节点信息
	allValidatorAddrs := g._validators()
	allValidatorAddrs = append(allValidatorAddrs, nodeAddr)
	g._setValidators(allValidatorAddrs)

	// 保存链的验证者节点公钥
	chainID := g.sdk.Block().ChainID()
	nodes := g._chainValidators(chainID)
	nodes[nodeAddr] = newValidator
	g._setChainValidators(chainID, nodes)

	// 如果是主链就调用广播接口，如果是侧链就调用通知接口
	if g.sdk.Helper().BlockChainHelper().IsSideChain() {
		//get mainChainID
		mainChainID := g.sdk.Helper().BlockChainHelper().GetMainChainID()

		// notify
		toChainIDs := make([]string, 0)
		toChainIDs = append(toChainIDs, mainChainID)
		g.sdk.Helper().IBCHelper().Run(func() {
			g.emitNewValidator(
				newValidator.Name,
				newValidator.PubKey,
				newValidator.NodeAddr,
				newValidator.RewardAddr,
				newValidator.Power,
			)
		}).Notify(toChainIDs)

	} else {
		// broadcast
		g.sdk.Helper().IBCHelper().Run(func() {
			g.emitNewValidator(
				newValidator.Name,
				newValidator.PubKey,
				newValidator.NodeAddr,
				newValidator.RewardAddr,
				newValidator.Power,
			)
		}).Broadcast()
	}

}

//SetPower set power for a validator
//@:public:method:gas[20000]
func (g *Governance) SetPower(pubKey types.PubKey, power int64) {

	sdk.RequireOwner()
	sdk.Require(len(pubKey) == PubKeyLen,
		types.ErrInvalidParameter, "Invalid pubKey.")
	sdk.Require(power >= 0,
		types.ErrInvalidParameter, "Invalid power.")

	nodeAddr := g.sdk.Helper().BlockChainHelper().CalcAccountFromPubKey(pubKey)
	validator := g._validator(nodeAddr)
	sdk.Require(g._chkValidator(nodeAddr) == true,
		types.ErrInvalidParameter, "Validator is not exist.")
	sdk.Require(validator.Power > 0,
		types.ErrInvalidParameter, "Validator is not exist.")

	allValidators, totalPower := g.getAllValidators()
	sdk.Require(len(allValidators) >= 4,
		types.ErrInvalidParameter, "The number of validator should be more than three")

	g.checkPower(nodeAddr, power, allValidators)

	chainID := g.sdk.Block().ChainID()
	nodes := g._chainValidators(chainID)
	if power == 0 {
		sdk.Require(validator.Power < totalPower/3,
			types.ErrInvalidParameter, "The power of deleted validator should be less than 1/3 total power")

		// delete chainValidators
		delete(nodes, nodeAddr)
	} else {
		nodeAddr := g.sdk.Helper().BlockChainHelper().CalcAccountFromPubKey(pubKey)

		node := nodes[nodeAddr]
		node.Power = power
		nodes[nodeAddr] = node
	}

	g._setChainValidators(chainID, nodes)

	validator.Power = power
	g._setValidator(validator)

	if g.sdk.Helper().BlockChainHelper().IsSideChain() {
		//get mainChainID
		mainChainID := g.sdk.Helper().BlockChainHelper().GetMainChainID()

		// notify
		toChainIDs := make([]string, 0)
		toChainIDs = append(toChainIDs, mainChainID)
		g.sdk.Helper().IBCHelper().Run(func() {
			g.emitSetPower(
				validator.Name,
				validator.PubKey,
				validator.NodeAddr,
				validator.RewardAddr,
				validator.Power,
			)
		}).Notify(toChainIDs)
	} else {
		// broadcast
		g.sdk.Helper().IBCHelper().Run(func() {
			g.emitSetPower(
				validator.Name,
				validator.PubKey,
				validator.NodeAddr,
				validator.RewardAddr,
				validator.Power,
			)
		}).Broadcast()
	}
}

//SetRewardAddr set a reward address by pubKey
//@:public:method:gas[20000]
func (g *Governance) SetRewardAddr(pubKey types.PubKey, rewardAddr types.Address) {

	sdk.RequireOwner()
	sdk.Require(len(pubKey) == PubKeyLen,
		types.ErrInvalidParameter, "Invalid pubKey.")
	sdk.RequireAddress(rewardAddr)

	nodeAddr := g.sdk.Helper().BlockChainHelper().CalcAccountFromPubKey(pubKey)
	validator := g._validator(nodeAddr)
	sdk.Require(g._chkValidator(nodeAddr) == true,
		types.ErrInvalidParameter, "Validator is not exist.")
	sdk.Require(validator.Power > 0,
		types.ErrInvalidParameter, "Validator is not exist.")

	validator.RewardAddr = rewardAddr
	g._setValidator(validator)

	g.emitSetRewardAddr(
		validator.Name,
		validator.PubKey,
		validator.NodeAddr,
		validator.RewardAddr,
		validator.Power,
	)
}

//SetRewardStrategy update reward strategy
//@:public:method:gas[50000]
func (g *Governance) SetRewardStrategy(strategy string) {

	sdk.RequireOwner()

	rwdStrategy := g.checkRewardStrategy(strategy)
	g.updateRewardStrategy(rwdStrategy)
}

// SetBlockFrequency set generate block Interval
//@:public:method:gas[20000]
func (g *Governance) SetConfig(configJson string) {

	var cfgMap = make(map[string]interface{}, 3)
	cfg := &SetConfig{}

	sdk.RequireOwner()
	sdk.Require(len(configJson) != 0,
		types.ErrInvalidParameter, "config must not null")

	err := jsoniter.Unmarshal([]byte(configJson), cfg)
	sdk.Require(err == nil, types.ErrInvalidParameter, "Invalid params")

	interval := cfg.CreateEmptyBlocksInterval

	// 对时间作合法性检查
	sdk.Require(interval >= 0 && interval < 50000,
		types.ErrInvalidParameter, "Invalid interval time")

	cfgMap["force_interval_block_switch"] = cfg.ForceIntervalBlockSwitch
	cfgMap["create_empty_blocks_interval"] = cfg.CreateEmptyBlocksInterval
	cfgMap["create_empty_blocks"] = cfg.CreateEmptyBlocks

	g.emitSetConfig(
		cfg.CreateEmptyBlocks,
		cfg.ForceIntervalBlockSwitch,
		cfg.CreateEmptyBlocksInterval,
	)
	g._setConfig(cfgMap)
}

//SetBVMStatus set BVM on or off
//@:public:method:gas[20000]
func (g *Governance) SetBVMStatus(enable bool) {

	sdk.RequireOwner()

	if enable {
		sdk.Require(g._chkBVMStatus(),
			types.ErrInvalidParameter, "The first time can only disable BVM")

		sdk.Require(!g._BVMStatus(),
			types.ErrInvalidParameter, "BVM cannot be enabled repeatedly")

		g._setBVMStatus(enable)

	} else {
		if !g._chkBVMStatus() {
			g._setBVMStatus(enable)

		} else {
			sdk.Require(g._BVMStatus(),
				types.ErrInvalidParameter, "BVM cannot be disable repeatedly")

			g._setBVMStatus(enable)
		}
	}

	g.emitSetBVMStatus(enable)
}

func (g *Governance) checkPower(nodeAddr types.Address, power int64, allValidators []InfoOfValidator) {
	totalPower := bn.N(power)
	maxPower := bn.N(power)

	forx.Range(allValidators, func(i int, validator InfoOfValidator) bool {
		if validator.NodeAddr != nodeAddr {
			totalPower = totalPower.AddI(validator.Power)
			if bn.N(validator.Power).IsGreaterThan(maxPower) {
				maxPower = bn.N(validator.Power)
			}
		}
		return true
	})

	// If the maxPower is equal to or over 1/3 totalPower, panic
	sdk.Require(maxPower.IsLessThan(totalPower.DivI(3)),
		types.ErrInvalidParameter, "Invalid power, max power is greater than or equal to 1/3 total power.")
}

func (g *Governance) checkRewardStrategy(strategy string) (rwdStrategy *RewardStrategy) {
	rwdStrategy = &RewardStrategy{}
	err := jsoniter.Unmarshal([]byte(strategy), rwdStrategy)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	g.checkRewardStrategyList(rwdStrategy.Strategy)
	g.checkEffectHeight(rwdStrategy.EffectHeight)

	return
}

func (g *Governance) checkRewardStrategyList(rwdStrategyList []Reward) {

	var percent int
	var haveNameOfValidators bool
	forx.Range(rwdStrategyList, func(i int, st Reward) bool {
		// check name length
		sdk.Require(len(st.Name) > 0 && len(st.Name) <= MaxNameLen,
			types.ErrInvalidParameter, "Invalid name in strategy")

		nodePerStr := st.RewardPercent
		// check percent format
		if strings.Contains(st.RewardPercent, ".") {
			index := strings.IndexByte(st.RewardPercent, '.')
			sub := []byte(st.RewardPercent)[index+1:]
			sdk.Require(len(sub) == 2,
				types.ErrInvalidParameter, "Invalid reward percent")
			nodePerStr = strings.Replace(st.RewardPercent, ".", "", -1)
			if len(nodePerStr) == 5 {
				sdk.Require(nodePerStr == "10000",
					types.ErrInvalidParameter, "Invalid reward percent")
			} else {
				sdk.Require(len(nodePerStr) == 4,
					types.ErrInvalidParameter, "Invalid reward percent")
			}
		} else {
			nodePerStr += "00"
		}

		nodePer, err := strconv.Atoi(nodePerStr)
		sdk.RequireNotError(err, types.ErrInvalidParameter)

		percent = percent + nodePer
		sdk.Require(nodePer > 0 && percent <= 10000,
			types.ErrInvalidParameter, "Invalid reward percent")

		// Check Address, check name with "validators", for validators, we don't care about its address
		if st.Name == "validators" {
			haveNameOfValidators = true
		} else {
			sdk.RequireAddress(st.Address)
		}
		return true
	})

	sdk.Require(percent == 10000,
		types.ErrInvalidParameter, "Invalid reward percent")
	sdk.Require(haveNameOfValidators,
		types.ErrInvalidParameter, "Lose name of validators")
}

func (g *Governance) checkEffectHeight(effectHeight int64) {

	sdk.Require(effectHeight > g.sdk.Block().Height(),
		types.ErrInvalidParameter, "Invalid effect height: it should greater than current block height.")

	rewardStrategies := g._rewardStrategies()

	forx.Range(rewardStrategies, func(i int, item RewardStrategy) bool {
		sdk.Require(effectHeight > item.EffectHeight,
			types.ErrInvalidParameter, "Invalid effect height: it should greater than others.")
		return true
	})
}

func (g *Governance) updateRewardStrategy(rwdStrategy *RewardStrategy) {

	rewardStrategies := g._rewardStrategies()
	rewardStrategies = append(rewardStrategies, *rwdStrategy)

	var first int
	forx.Range(rewardStrategies, func(i int, rewardStrategy RewardStrategy) bool {
		if rewardStrategy.EffectHeight <= g.sdk.Block().Height() {
			first = i
		}
		return true
	})
	rewardStrategies = rewardStrategies[first:]
	g._setRewardStrategies(rewardStrategies)

	g.emitSetRewardStrategy(
		rwdStrategy.Strategy,
		rwdStrategy.EffectHeight,
	)
}

// get all validators
func (g *Governance) getAllValidators() (allValidators []InfoOfValidator, totalPower int64) {

	sdk.Require(g._chkValidators(),
		types.ErrInvalidParameter, "Get all validators err, no pubKeys in db.")

	allAddrs := g._validators()
	forx.Range(allAddrs, func(i int, addr string) bool {
		validator := g._validator(addr)
		if validator.Power > 0 {
			allValidators = append(allValidators, validator)
			totalPower = totalPower + validator.Power
		}

		return true
	})
	return
}
