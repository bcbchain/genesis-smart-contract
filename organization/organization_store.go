package organization

import (
	"github.com/bcbchain/sdk/sdk/std"
)

func (o *Organization) _organization(orgID string) std.Organization {
	return *o.sdk.Helper().StateHelper().GetEx("/organization/"+orgID, new(std.Organization)).(*std.Organization)
}

func (o *Organization) _setOrganization(org std.Organization) {
	o.sdk.Helper().StateHelper().Set("/organization/"+org.OrgID, &org)
}

func (o *Organization) _chkOrganization(orgID string) bool {
	return o.sdk.Helper().StateHelper().Check("/organization/" + orgID)
}

func (o *Organization) _allOrganization() []string {
	return *o.sdk.Helper().StateHelper().GetEx("/organization/all/0", new([]string)).(*[]string)
}

func (o *Organization) _setAllOrganization(allOrg []string) {
	o.sdk.Helper().StateHelper().Set("/organization/all/0", &allOrg)
}

func (o *Organization) _chkAllOrganization() bool {
	return o.sdk.Helper().StateHelper().Check("/organization/all/0")
}
