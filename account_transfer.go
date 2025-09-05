package tbdb

import (
	"slices"
	"time"

	tb "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// AccountTransferFilter defines filter account transfer parameters.
type AccountTransferFilter struct {
	// The minimum transfer time from which results will be returned, inclusive range; required.
	TimeMin time.Time
	// The maximum transfer time from which results will be returned, inclusive range. Required
	TimeMax time.Time
	// AccountID is unique identifier of account. Required or must not be zero.
	AccountID Uint128
	// Filter the results by 128-bit user-defined data on transfer event.
	// Optional; set to zero to disable the filter
	UserData128 Uint128
	// Filter the results by 64-bit user-defined data on transfer event.
	// Optional; set to zero to disable the filter
	UserData64 uint64
	// Filter the results by 32-bit user-defined data on transfer event.
	// Optional; set to zero to disable the filter
	UserData32 uint32
	// The maximum number of results that can be returned by this query.
	Limit uint32
	// Filter the results by Transfer.code.
	// Optional; set to zero to disable the filter
	Code uint16
	// Monetary represent account monetary type; required.
	Monetary Amount
	// To specifies querying behavior.
	Flags AccountFilterFlags
}

// accountTransferFilterValidate is helper method to validate and modify the AccountTransferFilter.
func (i *Instance) accountTransferFilterValidate(filter *AccountTransferFilter) (tb.Client, error) {
	// Validate.
	switch {
	case filter.AccountID.IsZero():
		return nil, ErrAccountIDMustNotBeZero
	case filter.TimeMin.IsZero():
		return nil, ErrTimeMinMustNotBeZero
	case filter.TimeMax.IsZero():
		return nil, ErrTimeMaxMustNotBeZero
	case filter.Monetary == nil:
		return nil, ErrMonetaryMustNotBeNil
	}
	// Get TigerBeetle client instance.
	cln, err := i.Client()
	if err != nil {
		return nil, err
	}

	// Unfortunately until v0.16.55 flags can not be zero
	if filter.Flags.ToUint32() == 0 {
		filter.Flags.Debits = true
		filter.Flags.Credits = true
	} else if filter.Flags.Reversed && (!filter.Flags.Debits && !filter.Flags.Credits) {
		filter.Flags.Debits = true
		filter.Flags.Credits = true
	}
	// Unfortunately until v0.16.55 limit can not be zero
	if filter.Limit == 0 {
		filter.Limit = 1
	} else if filter.Limit > uint32(TigerBeetleMaxBatch) {
		filter.Limit = uint32(TigerBeetleMaxBatch)
	}
	return cln, nil
}

// AccountBalance defines account's balance record.
type AccountBalance struct {
	// This is the time the account balance was updated, as nanoseconds since UNIX epoch.
	// The timestamp refers to the same Transfer.timestamp which changed the Account.
	Timestamp uint64
	// Amount of pending debits.
	DebitsPending Amount
	// Amount of posted debits.
	DebitsPosted Amount
	// Amount of pending credits.
	CreditsPending Amount
	// Amount of posted credits.
	CreditsPosted Amount
}

// GetHisotricalBalances fetchs the historical account balances.
// The max size of AccountBalance array result is equal to TigerBeetleMaxBatch,
// even the filter.Limit set to > TigerBeetleMaxBatch.
func (i *Instance) GetHisotricalBalances(filter AccountTransferFilter) ([]AccountBalance, error) {
	cln, err := i.accountTransferFilterValidate(&filter)
	if err != nil {
		return nil, err
	}

	// Perform TigerBeetle GetAccountBalances.
	tbAccountBalances, err := cln.GetAccountBalances(types.AccountFilter{
		AccountID:    toBinding(filter.AccountID),
		UserData128:  toBinding(filter.UserData128),
		UserData64:   filter.UserData64,
		UserData32:   filter.UserData32,
		Code:         filter.Code,
		TimestampMin: uint64(filter.TimeMin.UTC().UnixNano()),
		TimestampMax: uint64(filter.TimeMax.UTC().UnixNano()),
		Limit:        filter.Limit,
		Flags:        filter.Flags.ToUint32(),
	})
	if err != nil {
		return nil, err
	}

	// Convert TigerBeetle's AccountBalance to AccountBalance.
	accountBalances := make([]AccountBalance, 0, len(tbAccountBalances))
	for _, tbAccountBalance := range tbAccountBalances {
		accountBalances = append(accountBalances, AccountBalance{
			Timestamp:      tbAccountBalance.Timestamp,
			DebitsPending:  filter.Monetary.SetUint128Value(fromBinding(tbAccountBalance.DebitsPending)),
			DebitsPosted:   filter.Monetary.SetUint128Value(fromBinding(tbAccountBalance.DebitsPosted)),
			CreditsPending: filter.Monetary.SetUint128Value(fromBinding(tbAccountBalance.CreditsPending)),
			CreditsPosted:  filter.Monetary.SetUint128Value(fromBinding(tbAccountBalance.CreditsPosted)),
		})
	}
	return accountBalances, nil
}

// AccountTransfer defines account's transfer record.
type AccountTransfer struct {
	// ID is unique account identifier.
	ID Uint128
	// DebitAccountID is debit account id.
	DebitAccountID Uint128
	// CreditAccountID is credit account id.
	CreditAccountID Uint128
	// PendingID reference to pending transfer id.
	PendingID Uint128
	// Amount is the monetary amount.
	Amount Amount
	// UserData128 is 128-bit user-defined data.
	UserData128 Uint128
	// UserData64 is 64-bit user-defined data.
	UserData64 uint64
	// Timestamp is the time transfer was created; in nanoseconds since UNIX epoch.
	Timestamp uint64
	// UserData32 is 32-bit user-defined data.
	UserData32 uint32
	// Timeout is the interval of pending transfer; in seconds
	Timeout uint32
	// Ledger this account belongs to.
	Ledger uint32
	// This is a user-defined enum denoting the reason for (or category of) the transfer.
	Code uint16
	// Flags defines the behavior of transfer.
	Flags TransferFlags
}

// GetAccountTransfers fetchs the transfers record involving a given account.
// The max size of AccountTransfer array result is equal to TigerBeetleMaxBatch,
// even the filter.Limit set to > TigerBeetleMaxBatch.
func (i *Instance) GetAccountTransfers(filter AccountTransferFilter) ([]AccountTransfer, error) {
	cln, err := i.accountTransferFilterValidate(&filter)
	if err != nil {
		return nil, err
	}

	// Perform TigerBeetle GetAccountTransfers.
	tbAccountTransfers, err := cln.GetAccountTransfers(types.AccountFilter{
		AccountID:    toBinding(filter.AccountID),
		UserData128:  toBinding(filter.UserData128),
		UserData64:   filter.UserData64,
		UserData32:   filter.UserData32,
		Code:         filter.Code,
		TimestampMin: uint64(filter.TimeMin.UTC().UnixNano()),
		TimestampMax: uint64(filter.TimeMax.UTC().UnixNano()),
		Limit:        filter.Limit,
		Flags:        filter.Flags.ToUint32(),
	})
	if err != nil {
		return nil, err
	}

	// Convert TigerBeetle's Transfer to AccountTransfer.
	accountTransfers := make([]AccountTransfer, 0, len(tbAccountTransfers))
	for _, transfer := range tbAccountTransfers {
		flags := transfer.TransferFlags()
		accountTransfers = append(accountTransfers, AccountTransfer{
			ID:              fromBinding(transfer.ID),
			DebitAccountID:  fromBinding(transfer.DebitAccountID),
			CreditAccountID: fromBinding(transfer.CreditAccountID),
			Amount:          filter.Monetary.SetUint128Value(fromBinding(transfer.Amount)),
			PendingID:       fromBinding(transfer.PendingID),
			UserData128:     fromBinding(transfer.UserData128),
			UserData64:      transfer.UserData64,
			Timestamp:       transfer.Timestamp,
			UserData32:      transfer.UserData32,
			Timeout:         transfer.Timeout,
			Ledger:          transfer.Ledger,
			Code:            transfer.Code,
			Flags: TransferFlags{
				Linked:              flags.Linked,
				Pending:             flags.Pending,
				PostPendingTransfer: flags.PostPendingTransfer,
				VoidPendingTransfer: flags.VoidPendingTransfer,
				BalancingDebit:      flags.BalancingDebit,
				BalancingCredit:     flags.BalancingCredit,
				ClosingDebit:        flags.ClosingDebit,
				ClosingCredit:       flags.ClosingCredit,
				Imported:            flags.Imported,
			},
		})
	}
	return accountTransfers, nil
}

// AccountStatement defines account's statement record.
type AccountStatement struct {
	// ID is unique account identifier.
	ID Uint128
	// DebitAccountID is debit account id.
	DebitAccountID Uint128
	// CreditAccountID is credit account id.
	CreditAccountID Uint128
	// Debit is the monetary amount for debit transfer.
	Debit Amount
	// Credit is the monetary amount for credit transfer.
	Credit Amount
	// BalanceBefore is the monetary of balance before.
	BalanceBefore Amount
	// BalanceAfter is the monetary of balance after.
	BalanceAfter Amount
	// UserData128 is 128-bit user-defined data.
	UserData128 Uint128
	// UserData64 is 64-bit user-defined data.
	UserData64 uint64
	// Timestamp is the time transfer was created; in nanoseconds since UNIX epoch.
	Timestamp uint64
	// UserData32 is 32-bit user-defined data.
	UserData32 uint32
	// Ledger this account belongs to.
	Ledger uint32
	// This is a user-defined enum denoting the reason for (or category of) the transfer.
	Code uint16
	// Flags defines the behavior of transfer.
	Flags TransferFlags
	// Additional is user-defined data that be writen by given closure.
	Additional any
}

// StatementClosureFn is closure function of statement.
type StatementClosureFn func(
	id, userData128 Uint128,
	userData64 uint64,
	userData32 uint32,
	code uint16,
	flags TransferFlags,
	timestamp uint64,
) any

// GetAccountStatements fetchs the account statements record involving a given account
// The max size of AccountStatement array result is equal to TigerBeetleMaxBatch,
// even the filter.Limit set to > TigerBeetleMaxBatch.
func (i *Instance) GetAccountStatements(
	filter AccountTransferFilter,
	closureFn ...StatementClosureFn,
) ([]AccountStatement, error) {
	// Get balances.
	balances, err := i.GetHisotricalBalances(filter)
	if err != nil {
		return nil, err
	}

	// Get transfers.
	transfers, err := i.GetAccountTransfers(filter)
	if err != nil {
		return nil, err
	}

	// Make AccountStatement data.
	var clsFn StatementClosureFn
	if len(closureFn) > 0 {
		if closureFn[0] != nil {
			clsFn = closureFn[0]
		}
	}
	statements := make([]AccountStatement, 0, len(transfers))
	for _, transfer := range transfers {
		// Determine is debit or not.
		isDebit := transfer.DebitAccountID == filter.AccountID
		switch {
		// If transfer debit && state is void pending, should be credited back the amount.
		case isDebit && transfer.Flags.VoidPendingTransfer:
			isDebit = false

		// If transfer is debit && post post pending, skip data.
		case isDebit && transfer.Flags.PostPendingTransfer:
			continue
		}

		// Statement data.
		statement := AccountStatement{
			ID:              transfer.ID,
			DebitAccountID:  transfer.DebitAccountID,
			CreditAccountID: transfer.CreditAccountID,
			Debit:           filter.Monetary.SetFloat64Value(0),
			Credit:          filter.Monetary.SetFloat64Value(0),
			BalanceBefore:   filter.Monetary.SetFloat64Value(0),
			BalanceAfter:    filter.Monetary.SetFloat64Value(0),
			UserData128:     transfer.UserData128,
			UserData64:      transfer.UserData64,
			Timestamp:       transfer.Timestamp,
			UserData32:      transfer.UserData32,
			Ledger:          transfer.Ledger,
			Code:            transfer.Code,
			Flags:           transfer.Flags,
		}

		// Get balance element fron balances slice that Timestamp equal to transfer.Timestamp.
		balancesIdx := slices.IndexFunc(balances, func(b AccountBalance) bool {
			return b.Timestamp == statement.Timestamp
		})
		if balancesIdx < 0 || balancesIdx > len(balances) {
			continue
		}
		balance := balances[balancesIdx]

		// Balance after.
		debit, err := balance.DebitsPending.Add(balance.DebitsPosted.ToUint128FromValue())
		if err != nil {
			continue
		}
		after, err := balance.CreditsPosted.Sub(debit)
		if err != nil {
			continue
		}
		statement.BalanceAfter = filter.Monetary.SetUint128Value(after)

		// Balance before.
		if isDebit {
			statement.Debit = transfer.Amount
			before, err := statement.BalanceAfter.Add(transfer.Amount.ToUint128FromValue())
			if err != nil {
				continue
			}
			statement.BalanceBefore = filter.Monetary.SetUint128Value(before)
		} else {
			statement.Credit = transfer.Amount
			before, err := statement.BalanceAfter.Sub(transfer.Amount.ToUint128FromValue())
			if err != nil {
				continue
			}
			statement.BalanceBefore = filter.Monetary.SetUint128Value(before)
		}

		// Append.
		if clsFn != nil {
			statement.Additional = clsFn(
				statement.ID,
				statement.UserData128,
				statement.UserData64,
				statement.UserData32,
				statement.Code,
				statement.Flags,
				statement.Timestamp,
			)
		}
		statements = append(statements, statement)
	}
	return statements, nil
}
