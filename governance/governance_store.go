package governance

// all validators pubKeys
func (g *Governance) _validators() []string {
	return *g.sdk.Helper().StateHelper().GetEx("/validators/all/0", &[]string{}).(*[]string)
}

func (g *Governance) _setValidators(value []string) {
	g.sdk.Helper().StateHelper().Set("/validators/all/0", &value)
}

func (g *Governance) _chkValidators() bool {
	return g.sdk.Helper().StateHelper().Check("/validators/all/0")
}

// validator
func (g *Governance) _validator(nodeAddr string) InfoOfValidator {
	return *g.sdk.Helper().StateHelper().GetEx("/validator/"+nodeAddr, &InfoOfValidator{}).(*InfoOfValidator)
}

func (g *Governance) _setValidator(validator InfoOfValidator) {
	g.sdk.Helper().StateHelper().Set("/validator/"+validator.NodeAddr, &validator)
}

func (g *Governance) _delValidator(nodeAddr string) {
	g.sdk.Helper().StateHelper().McDelete("/validator/" + nodeAddr)
}

func (g *Governance) _chkValidator(validatorAddr string) bool {
	return g.sdk.Helper().StateHelper().Check("/validator/" + validatorAddr)
}

// 奖励策略
func (g *Governance) _rewardStrategies() (rewardStrategys []RewardStrategy) {
	return *g.sdk.Helper().StateHelper().GetEx("/rewardstrategys", &rewardStrategys).(*[]RewardStrategy)
}

func (g *Governance) _setRewardStrategies(rewardStrategys []RewardStrategy) {
	g.sdk.Helper().StateHelper().Set("/rewardstrategys", &rewardStrategys)
}

func (g *Governance) _chkRewardStrategies() bool {
	return g.sdk.Helper().StateHelper().Check("/rewardstrategys")
}

func (g *Governance) _chainValidators(chainID string) (chainValidators map[string]InfoOfValidator) {
	return *g.sdk.Helper().StateHelper().GetEx("/ibc/"+chainID, &chainValidators).(*map[string]InfoOfValidator)
}

// ChainValidators
func (g *Governance) _setChainValidators(chainID string, cvp map[string]InfoOfValidator) {
	g.sdk.Helper().StateHelper().McSet("/ibc/"+chainID, &cvp)
}

// SetBlockInterval
func (g *Governance) _setConfig(tdmConfig map[string]interface{}) {
	g.sdk.Helper().StateHelper().Set("/config", &tdmConfig)
}

func (g *Governance) _BVMStatus() bool {
	var enable bool
	return *g.sdk.Helper().StateHelper().GetEx(g.keyOfBVMStatus(), &enable).(*bool)
}

func (g *Governance) _setBVMStatus(enable bool) {
	g.sdk.Helper().StateHelper().Set(g.keyOfBVMStatus(), &enable)
}

func (g *Governance) _chkBVMStatus() bool {
	return g.sdk.Helper().StateHelper().Check(g.keyOfBVMStatus())
}

func (g *Governance) keyOfBVMStatus() string {
	return "/bvm/status"
}
