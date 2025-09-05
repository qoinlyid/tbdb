package tbdb

import (
	"errors"
	"fmt"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type Transfer = types.Transfer

// CreateTransferResult represents the result of transfer creation operations.
type CreateTransferResult = types.CreateTransferResult

// TransferData defines transfer argument for create TigerBeetle's transfer event.
type TransferData struct {
	// ID is unique account identifier (auto-generated if zero).
	ID Uint128
	// DebitAccountID is debit account id, required.
	DebitAccountID Uint128
	// CreditAccountID is credit account id, required.
	CreditAccountID Uint128
	// Amount must be greater than 0.
	Amount Amount
	// UserData128 is 128-bit user-defined data.
	UserData128 Uint128
	// UserData64 is 64-bit user-defined data.
	UserData64 uint64
	// UserData32 is 32-bit user-defined data.
	UserData32 uint32
	// Ledger this account belongs to.
	Ledger Ledger

	// Transfer flags.
	flags TransferFlags
	// For post/void pending.
	pendingID Uint128
	code      uint16
	// For set transfer as pending.
	timeout uint32
	// For import.
	timestamp uint64
}

// TransferEventResult defines result data for create transfer event.
type TransferEventResult struct {
	// Index of the transfer in the original request.
	Index uint32
	// ID is unique transfer identifier (auto-generated if zero).
	ID Uint128
	// Result is creation result status.
	Result CreateTransferResult
	// Err is error creation.
	Err error
}

// TransferEventResults defines result data for create transfer batch event.
type TransferEventResults struct {
	// Create transfer success & failed count.
	SuccessCount, FailedCount int
	// Silce of each result.
	Results []TransferEventResult
}

func (i *Instance) doTransfers(transfers []TransferData) (TransferEventResults, error) {
	// Validate.
	countTransfer := len(transfers)
	if countTransfer == 0 {
		return TransferEventResults{}, errors.New("no transfers given")
	}
	// Get TigerBeetle client instance.
	cln, err := i.Client()
	if err != nil {
		return TransferEventResults{}, err
	}

	// Convert TransferData structs to TigerBeetle's native Transfer format.
	tbTransfers := make([]types.Transfer, 0, countTransfer)
	generatedIDs := make([]Uint128, countTransfer)
	for idx, transfer := range transfers {
		var id Uint128
		if transfer.ID.IsZero() {
			id = NewID()
		} else {
			id = transfer.ID
		}
		generatedIDs[idx] = id

		var ledger uint32
		if transfer.Ledger != nil {
			ledger = uint32(transfer.Ledger.EncodeLedger())
		}

		tbTransfers = append(tbTransfers, types.Transfer{
			ID:              toBinding(id),
			DebitAccountID:  toBinding(transfer.DebitAccountID),
			CreditAccountID: toBinding(transfer.CreditAccountID),
			Amount:          toBinding(transfer.Amount.ToUint128FromValue()),
			PendingID:       toBinding(transfer.pendingID),
			UserData128:     toBinding(transfer.UserData128),
			UserData64:      transfer.UserData64,
			UserData32:      transfer.UserData32,
			Timeout:         transfer.timeout,
			Ledger:          ledger,
			Code:            transfer.code,
			Flags:           transfer.flags.ToUint16(),
			Timestamp:       transfer.timestamp,
		})
	}

	// Execute batch transfer creation.
	tbResults, err := cln.CreateTransfers(tbTransfers)
	if err != nil {
		return TransferEventResults{}, err
	}
	result := TransferEventResults{
		SuccessCount: countTransfer,
		Results:      make([]TransferEventResult, countTransfer),
	}

	// Initialize all results as successful first (most common case).
	for idx := range transfers {
		result.Results[idx] = TransferEventResult{
			Index:  uint32(idx),
			ID:     generatedIDs[idx],
			Result: types.TransferOK,
			Err:    nil,
		}
	}

	// Update only the failed transfers from TigerBeetle results.
	// TigerBeetle only returns results for failed transfers.
	for _, tbResult := range tbResults {
		if tbResult.Index < uint32(countTransfer) {
			var resultErr error
			var errCount int
			if tbResult.Result != types.TransferOK {
				resultErr = fmt.Errorf("account creation failed: %s", tbResult.Result.String())
				errCount = 1
			}

			result.SuccessCount = result.SuccessCount - errCount
			result.FailedCount = result.FailedCount + errCount
			result.Results[tbResult.Index] = TransferEventResult{
				Index:  tbResult.Index,
				ID:     generatedIDs[tbResult.Index],
				Result: tbResult.Result,
				Err:    resultErr,
			}
		}
	}
	return result, nil
}

// CreateTransfers .
func (i *Instance) CreateTransfers(transfers []TransferData) (TransferEventResults, error) {
	return i.doTransfers(transfers)
}

// CreatePendingTransfers .
func (i *Instance) CreatePendingTransfers(
	transfers []TransferData,
	linked bool,
	code uint16,
) (TransferEventResults, error) {
	flags := TransferFlags{Pending: true}
	flags.Linked = linked
	for i := range transfers {
		if i == len(transfers)-1 && flags.Linked {
			flags.Linked = false
		}
		transfers[i].code = code
		transfers[i].flags = flags
	}
	return i.doTransfers(transfers)
}

type ResolvePendingState uint8

const (
	ResolvePendingStatePost ResolvePendingState = 1
	ResolvePendingStateVoid ResolvePendingState = 2
)

// PendingTransfer defines pending transfer data that want to resolved by TigerBeetle's transfer event.
type PendingTransfer struct {
	// ID is unique account identifier (auto-generated if zero).
	ID Uint128
	// Pending transfer id that must reference to transfer id with pending state & required.
	PendingID Uint128
	// Amount that want to resolved from pending transfer.
	//
	// For post-pending state:
	//	- Amount less than actual pending transfer amount, then only this amount is posted & the remainder
	//	is restored to its original account.
	//	- Amount equal to actual pending transfer amount, then full pending amount will be posted
	//	- Amount greater than actual pending transfer amount, then will be failed.
	//
	// For void-pending state: Amount can be zero value.
	Amount Amount
	// UserData32 is 32-bit user-defined data.
	UserData32 uint32
	// State is type of resolve, either will posted or voided.
	State ResolvePendingState
}

// ResolvePendingTransfers resolves pending transfers based on given state, either its posted or voided.
func (i *Instance) ResolvePendingTransfers(pendings []PendingTransfer) (TransferEventResults, error) {
	var flags TransferFlags
	transfers := make([]TransferData, 0, len(pendings))
	for _, pending := range pendings {
		if pending.State == ResolvePendingStatePost {
			flags.PostPendingTransfer = true
		} else {
			flags.VoidPendingTransfer = true
		}

		transfers = append(transfers, TransferData{
			ID:         pending.ID,
			Amount:     pending.Amount,
			pendingID:  pending.PendingID,
			UserData32: pending.UserData32,
			flags:      flags,
		})
	}
	return i.doTransfers(transfers)
}
