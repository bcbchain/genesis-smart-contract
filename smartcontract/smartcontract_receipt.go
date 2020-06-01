package smartcontract

import "github.com/bcbchain/sdk/sdk/types"

func (s *SmartContract) emitAuthorize(deployer types.Address, orgID string) {
	type authorize struct {
		OrgID    string        `json:"orgID"`
		Deployer types.Address `json:"deployer"`
	}

	s.sdk.Helper().ReceiptHelper().Emit(
		authorize{
			OrgID:    orgID,
			Deployer: deployer,
		})
}

func (s *SmartContract) emitDeployContract(
	contractAddr types.Address,
	name string,
	version string,
	orgID string,
	codeHash types.Hash,
	codeData []byte,
	codeDevSig string,
	codeOrgSig string,
	effectHeight int64,
	owner types.Address) {

	type deployContract struct {
		ContractAddr types.Address `json:"contractAddr"`
		Name         string        `json:"name"`
		Version      string        `json:"version"`
		OrgID        string        `json:"orgId"`
		CodeHash     types.Hash    `json:"codeHash"`
		CodeData     []byte        `json:"codeData"`
		CodeDevSig   string        `json:"codeDevSig"`
		CodeOrgSig   string        `json:"codeOrgSig"`
		EffectHeight int64         `json:"effectHeight"`
		Owner        types.Address `json:"owner"`
	}

	s.sdk.Helper().ReceiptHelper().Emit(
		deployContract{
			ContractAddr: contractAddr,
			Name:         name,
			Version:      version,
			OrgID:        orgID,
			CodeHash:     codeHash,
			CodeData:     codeData,
			CodeDevSig:   codeDevSig,
			CodeOrgSig:   codeOrgSig,
			EffectHeight: effectHeight,
			Owner:        owner,
		})
}

func (s *SmartContract) emitForbidContract(
	contractAddr types.Address,
	loseHeight int64) {

	type forbidContract struct {
		ContractAddr types.Address `json:"contractAddr"`
		LoseHeight   int64         `json:"loseHeight"`
	}

	s.sdk.Helper().ReceiptHelper().Emit(
		forbidContract{
			ContractAddr: contractAddr,
			LoseHeight:   loseHeight,
		})
}
