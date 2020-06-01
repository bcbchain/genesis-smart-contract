package organization

import (
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/forx"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
	"unicode/utf8"
)

//Organization This is struct of contract
//@:contract:organization
//@:version:2.2
//@:organization:orgJgaGConUyK81zibntUBjQ33PKctpk1K1G
//@:author:5e8339cb1a5cce65602fd4f57e115905348f7e83bcbe38dd77694dbe1f8903c9
type Organization struct {
	sdk sdk.ISmartContract
}

//InitChain Constructor of this Organization
//@:constructor
func (o *Organization) InitChain() {

}

//UpdateChain Constructor of this Organization
//@:constructor
func (o *Organization) UpdateChain() {
	if !o._chkAllOrganization() {
		var allOrg []string
		orgList := []string{
			"orgJgaGConUyK81zibntUBjQ33PKctpk1K1G", //genesis
			"orgHiMBuqDog1EJwAjoWFBmta2Rt7uDpAzi",  //yy
			"orgCZkw5xz9DYa3h5pJ2CzZSuGHRCj2ot5xq", //jiujiu
			"org3H4fcdAKNi6MpZNbUT6oCktHYNrpDtB7p", //jiuj
			"org6epUdAFZ93p5RPcw3hAqLwJ6Nr7GZQkz"}  //bcbjr
		forx.Range(orgList, func(i int, v string) {
			if o._chkOrganization(v) {
				allOrg = append(allOrg, v)
			}

		})
		o._setAllOrganization(allOrg)
	}
}

const (
	//PubKeyLen is public key length
	PubKeyLen = 32
	//MaxNameLen is the max length of organization name
	MaxNameLen = 256
)

//RegisterOrganization register a new organization and return org ID
//@:public:method:gas[500000]
func (o *Organization) RegisterOrganization(name string) string {

	sdk.RequireMainChain()

	sdk.Require(name != "" && len(name) <= MaxNameLen,
		types.ErrInvalidParameter, "Invalid name.")

	sdk.Require(utf8.Valid([]byte(name)),
		types.ErrInvalidParameter, "Invalid name, the name must be utf-8 encoding")

	orgID := o.sdk.Helper().BlockChainHelper().CalcOrgID(name)
	sdk.Require(o._chkOrganization(orgID) == false,
		types.ErrInvalidParameter, "Organization name already exists.")

	newOrganization := std.Organization{
		OrgID:            orgID,
		Name:             name,
		OrgOwner:         o.sdk.Message().Sender().Address(),
		ContractAddrList: []types.Address{},
		OrgCodeHash:      []byte{},
		Signers:          []types.PubKey{},
	}
	o._setOrganization(newOrganization)

	allOrg := o._allOrganization()
	allOrg = append(allOrg, orgID)
	o._setAllOrganization(allOrg)

	o.emitNewOrganization(
		orgID,
		name,
		o.sdk.Message().Sender().Address(),
	)
	return orgID
}

// SetSigners set organization signers
//@:public:method:gas[500000]
func (o *Organization) SetSigners(orgID string, pubKeys []types.PubKey) {

	sdk.Require(orgID != "",
		types.ErrInvalidParameter, "Invalid organization ID.")
	sdk.Require(o._chkOrganization(orgID) == true,
		types.ErrInvalidParameter, "Invalid organization ID.")
	sdk.Require(len(pubKeys) > 0,
		types.ErrInvalidParameter, "PubKey list can not be empty.")

	forx.Range(pubKeys, func(i int, pubKey types.PubKey) bool {
		sdk.Require(len(pubKey) == PubKeyLen,
			types.ErrInvalidParameter, "Invalid pubkey.")
		return true
	})

	org := o._organization(orgID)
	sdk.Require(o.sdk.Message().Sender().Address() == org.OrgOwner, types.ErrNoAuthorization, "No authorization")

	org.Signers = pubKeys
	o._setOrganization(org)

	o.emitSetSigners(
		orgID,
		pubKeys,
	)
}
