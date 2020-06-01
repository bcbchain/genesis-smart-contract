package organization

import (
	"github.com/bcbchain/sdk/sdk/types"
)

func (o *Organization) emitNewOrganization(
	orgID string,
	name string,
	orgOwner types.Address) {

	type newOrganization struct {
		OrgID    string        `json:"orgID"`    // 组织机构ID
		Name     string        `json:"name"`     // 组织名字
		OrgOwner types.Address `json:"orgOwner"` // 组织拥有者地址
	}

	o.sdk.Helper().ReceiptHelper().Emit(
		newOrganization{
			OrgID:    orgID,
			Name:     name,
			OrgOwner: orgOwner,
		},
	)
}

func (o *Organization) emitSetSigners(
	orgID string,
	signers []types.PubKey) {

	type setSigners struct {
		OrgID   string         `json:"orgID"`   // 组织机构ID
		Signers []types.PubKey `json:"signers"` // 签名公钥列表
	}

	o.sdk.Helper().ReceiptHelper().Emit(
		setSigners{
			OrgID:   orgID,
			Signers: signers,
		},
	)
}
