package tokenissue

import (
	"fmt"
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/jsoniter"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
	"strings"
)

//@:public:ibc:gas[0]
func (ti *TokenIssue) Recast(ibcHash types.Hash) bool {
	if !ti.isCallByIBC() {
		return false
	}

	lockAssetChange, lockTransfer, bResult := ti.checkReceiptsForRecast()
	if !bResult {
		return bResult
	}

	return ti.recast(lockAssetChange, lockTransfer)
}

//@:public:ibc:gas[0]
func (ti *TokenIssue) Confirm(ibcHash types.Hash) {
	sdk.Require(ti.isCallByIBC(),
		types.ErrNoAuthorization, "")

	recastAssetChange := ti.checkReceiptsForConfirm()

	ti.confirm(recastAssetChange)
}

//@:public:ibc:gas[0]
func (ti *TokenIssue) Cancel(ibcHash types.Hash) {
	sdk.Require(ti.isCallByIBC(),
		types.ErrNoAuthorization, "")

	lockAssetChange, lockTransfer := ti.checkReceiptsForCancel()

	ti.unlock(lockAssetChange, lockTransfer)
}

//@:public:ibc:gas[0]
func (ti *TokenIssue) TryRecast(ibcHash types.Hash) bool {
	sdk.Require(ti.isCallByIBC(),
		types.ErrNoAuthorization, "")

	lockAssetChange, bResult := ti.checkReceiptsForTryRecast()
	if bResult == false {
		return bResult
	}

	return ti.tryRecast(lockAssetChange)
}

//@:public:ibc:gas[0]
func (ti *TokenIssue) ConfirmRecast(ibcHash types.Hash) {
	sdk.Require(ti.isCallByIBC(),
		types.ErrNoAuthorization, "")

	recastAssetChange := ti.checkReceiptsForConfirmRecast()

	ti.confirmRecast(recastAssetChange)
}

//@:public:ibc:gas[0]
func (ti *TokenIssue) CancelRecast(ibcHash types.Hash) {
	sdk.Require(ti.isCallByIBC(),
		types.ErrNoAuthorization, "")

	lockAssetChange := ti.checkReceiptsForCancelRecast()

	ti.cancelRecast(lockAssetChange)
}

//@:public:ibc:gas[0]
func (ti *TokenIssue) Notify(ibcHash types.Hash) {
	sdk.Require(ti.isCallByIBC(),
		types.ErrNoAuthorization, "")

	inputReceiptsLen := len(ti.sdk.Message().InputReceipts())
	sdk.Require(inputReceiptsLen >= 1,
		types.ErrInvalidParameter, fmt.Sprintf("expected receipt's count>=1, obtain: %d", inputReceiptsLen))

	inputReceipt := ti.sdk.Message().InputReceipts()[0]
	splitKey := strings.Split(string(inputReceipt.Key), "/")

	var receipt std.Receipt
	err := jsoniter.Unmarshal(inputReceipt.Value, &receipt)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	switch splitKey[2] {
	case "std::setGasPrice":
		ti.setGasPriceForNotify(receipt)
	case "std::setOwner":
		ti.setOwnerForNotify(receipt)
	case "tokenissue.Activate":
		ti.activateForNotify(receipt)
	default:
		sdk.Require(false, types.ErrInvalidParameter, "invalid receipts")
	}
}
