package governance

import (
	"fmt"
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/jsoniter"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
	"strings"
)

//@:public:ibc:gas[0]
func (g *Governance) Notify(ibcHash types.Hash) {
	sdk.Require(g.checkCalledByIBC(),
		types.ErrInvalidParameter, "Notify only called by ibc")

	inputReceiptsLen := len(g.sdk.Message().InputReceipts())
	sdk.Require(inputReceiptsLen >= 1,
		types.ErrInvalidParameter, fmt.Sprintf("expected receipt's count>=1, obtain: %d", inputReceiptsLen))

	inputReceipt := g.sdk.Message().InputReceipts()[0]
	splitKey := strings.Split(string(inputReceipt.Key), "/")

	var receipt std.Receipt
	err := jsoniter.Unmarshal(inputReceipt.Value, &receipt)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	switch splitKey[2] {
	case "governance.newValidator":
		g.newValidatorForNotify(receipt)
	case "governance.setPower":
		g.setPowerForNotify(receipt)
	}

}

func (g *Governance) checkCalledByIBC() bool {
	origins := g.sdk.Message().Origins()
	if len(origins) < 1 {
		return false
	}

	lastContract := g.sdk.Helper().ContractHelper().ContractOfAddress(origins[len(origins)-1])
	if lastContract.Name() == "ibc" && lastContract.OrgID() == g.sdk.Helper().GenesisHelper().OrgID() {
		return true
	}
	return false
}
