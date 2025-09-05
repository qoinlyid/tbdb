package tbdb

import (
	"errors"
	"fmt"
	"log"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// CreateAccountResult represents the result of account creation operations.
type CreateAccountResult = types.CreateAccountResult

// AccountCategory defines account category that underlying the TigerBeetle's account code.
type AccountCategory uint16

// Enum of account category.
const (
	// Account category used for account that is source and/or destination.
	AccountCategoryControl AccountCategory = 1000
	// Account category used for account that is asset account & respected 'DebitsMustNotExceedCredits' flags.
	AccountCategoryBalance AccountCategory = 1001
	// Account category used for account that is income account & respected 'DebitsMustNotExceedCredits' flags.
	AccountCategoryIncome AccountCategory = 1002
	// Account category used for account that is liabilities account.
	AccountCategoryLiabilities AccountCategory = 1003
	// Account category used for account that is testing account.
	AccountCategoryTesting AccountCategory = 9999
)

// CreateAccount defines account argument for create single account event.
type CreateAccount struct {
	// ID is unique account identifier (auto-generated if zero).
	ID Uint128
	// UserData128 is 128-bit user-defined data.
	UserData128 Uint128
	// UserData64 is 64-bit user-defined data.
	UserData64 uint64
	// UserData32 is 32-bit user-defined data.
	UserData32 uint32
	// Ledger this account belongs to.
	Ledger Ledger
}

// CreateAccounts defines account argument for create many account event.
type CreateAccounts struct {
	CreateAccount

	// Code is account type code.
	Code uint16
	// Flags is account behavior flags.
	Flags AccountFlags
	// Timestamp is account creation timestamp.
	Timestamp uint64
}

// AccountEventResult defines result data for create single account event.
type AccountEventResult struct {
	// Index of the account in the original request.
	Index uint32
	// ID is unique account identifier (auto-generated if zero).
	ID Uint128
	// Result is creation result status.
	Result CreateAccountResult
	// Err is error creation.
	Err error
}

// AccountEventResults defines result data for create many account event.
type AccountEventResults struct {
	// Create account success & failed count.
	SuccessCount, FailedCount int
	// Silce of each result.
	Results []AccountEventResult
}

// CreateAccountBatch wraps TigerBeetle's CreateAccounts method to create one or many account(s).
func (i *Instance) CreateAccountBatch(accounts []CreateAccounts) (AccountEventResults, error) {
	// Validate.
	countAccount := len(accounts)
	if countAccount == 0 {
		return AccountEventResults{}, errors.New("no accounts given")
	}
	// Get TigerBeetle client instance.
	cln, err := i.Client()
	if err != nil {
		return AccountEventResults{}, err
	}

	// Convert CreateAccount structs to TigerBeetle's native Account format.
	tbAccounts := make([]types.Account, 0, countAccount)
	generatedIDs := make([]Uint128, countAccount)
	for idx, account := range accounts {
		var id Uint128
		if account.ID.IsZero() {
			id = NewID()
		} else {
			id = account.ID
		}
		generatedIDs[idx] = id

		tbAccounts = append(tbAccounts, types.Account{
			ID:          toBinding(id),
			UserData128: toBinding(account.UserData128),
			UserData64:  account.UserData64,
			UserData32:  account.UserData32,
			Ledger:      uint32(account.Ledger.EncodeLedger()),
			Code:        account.Code,
			Flags:       account.Flags.ToUint16(),
			Timestamp:   account.Timestamp,
		})
	}

	// Execute batch account creation.
	tbResults, err := cln.CreateAccounts(tbAccounts)
	if err != nil {
		return AccountEventResults{}, err
	}
	result := AccountEventResults{
		SuccessCount: countAccount,
		Results:      make([]AccountEventResult, countAccount),
	}

	// Initialize all results as successful first (most common case).
	for idx := range accounts {
		result.Results[idx] = AccountEventResult{
			Index:  uint32(idx),
			ID:     generatedIDs[idx],
			Result: types.AccountOK,
			Err:    nil,
		}
	}

	// Update only the failed accounts from TigerBeetle results.
	// TigerBeetle only returns results for failed accounts.
	for _, tbResult := range tbResults {
		if tbResult.Index < uint32(countAccount) {
			var resultErr error
			var errCount int
			if tbResult.Result != types.AccountOK && tbResult.Result != types.AccountExists {
				resultErr = fmt.Errorf("account creation failed: %s", tbResult.Result.String())
				errCount = 1
			}

			result.SuccessCount = result.SuccessCount - errCount
			result.FailedCount = result.FailedCount + errCount
			result.Results[tbResult.Index] = AccountEventResult{
				Index:  tbResult.Index,
				ID:     generatedIDs[tbResult.Index],
				Result: tbResult.Result,
				Err:    resultErr,
			}
		}
	}
	return result, nil
}

// CreateAccount creates single TigerBeetle's account.
// This is simple method to just create a single account.
// Use 'CreateAccountBatch' to take advantage for high throughput instead.
func (i *Instance) CreateAccount(account CreateAccount, code uint16, flags AccountFlags) (AccountEventResult, error) {
	// Validate.
	switch {
	case account == (CreateAccount{}):
		return AccountEventResult{}, errors.New("account must be not empty")
	case flags.DebitsMustNotExceedCredits && flags.CreditsMustNotExceedDebits:
		return AccountEventResult{}, errors.New(
			"flags DebitsMustNotExceedCredits & CreditsMustNotExceedDebits are mutually exclusive",
		)
	case code == 0:
		return AccountEventResult{}, errors.New("code must be not zero")
	}
	flags.Linked = false
	flags.Imported = false

	result, err := i.CreateAccountBatch([]CreateAccounts{{
		CreateAccount: account,
		Code:          code,
		Flags:         flags,
	}})
	if err != nil {
		return AccountEventResult{}, err
	} else if len(result.Results) == 0 {
		return AccountEventResult{}, errors.New("create account event results is empty")
	}
	return result.Results[0], nil
}

// CreateAccountsWithCategory creates TigerBeetle's account with built-in category from this tbdb package.
// It will return 'ErrUnknownCategory' if given category is unknown.
// Some built-in category have their respected account flags, as shown on the list bellow:
//
//   - AccountCategoryControl: History
//   - AccountCategoryBalance: History, DebitsMustNotExceedCredits
//   - AccountCategoryIncome: History, DebitsMustNotExceedCredits
//   - AccountCategoryTesting: None
func (i *Instance) CreateAccountsWithCategory(category AccountCategory, accounts ...CreateAccount) (AccountEventResults, error) {
	// Validate.
	if len(accounts) == 0 {
		return AccountEventResults{}, errors.New("at least give 1 account")
	}
	if category != AccountCategoryControl && category != AccountCategoryBalance && category != AccountCategoryIncome &&
		category != AccountCategoryLiabilities && category != AccountCategoryTesting {
		return AccountEventResults{}, ErrUnknownCategory
	}

	// Set proper account flags based on category.
	flags := AccountFlags{History: true}
	switch category {
	case AccountCategoryBalance, AccountCategoryIncome:
		flags.DebitsMustNotExceedCredits = true
	case AccountCategoryTesting:
		flags.History = false
	}

	// Create single account.
	if len(accounts) == 1 {
		result, err := i.CreateAccount(accounts[0], uint16(category), flags)
		if err != nil {
			return AccountEventResults{}, err
		}
		success, failed := 1, 0
		if result.Err != nil {
			success, failed = 0, 1
		}
		return AccountEventResults{
			SuccessCount: success,
			FailedCount:  failed,
			Results:      []AccountEventResult{result},
		}, nil
	}

	// Params accounts must be greather than 1.
	// Set account's flags.Linked to true, so create accounts must be success all.
	batchAccount := make([]CreateAccounts, 0, len(accounts))
	flags.Linked = true
	for idx, account := range accounts {
		// In the end of iteration, the account's flags.Linked must be set to false.
		if idx == len(accounts)-1 {
			flags.Linked = false
		}

		batchAccount = append(batchAccount, CreateAccounts{
			CreateAccount: account,
			Code:          uint16(category),
			Flags:         flags,
		})
	}
	return i.CreateAccountBatch(batchAccount)
}

// Account defines TigerBeetle's account.
type Account struct {
	// ID is unique account identifier.
	ID Uint128
	// DebitsPending.
	DebitsPending Amount
	// DebitsPosted.
	DebitsPosted Amount
	// CreditsPending.
	CreditsPending Amount
	// CreditsPosted.
	CreditsPosted Amount
	// UserData128 is 128-bit user-defined data.
	UserData128 Uint128
	// UserData64 is 64-bit user-defined data.
	UserData64 uint64
	// UserData32 is 32-bit user-defined data.
	UserData32 uint32
	// Reserved may be used for additional data in the future.
	Reserved uint32
	// Ledger this account belongs to.
	Ledger uint32
	// Code is account type code.
	Code uint16
	// Flags is account behavior flags.
	Flags AccountFlags
	// Timestamp is account creation timestamp.
	Timestamp uint64
}

// AccountLookup defines lookup account parameters.
type AccountLookup struct {
	// ID is unique account identifier.
	ID Uint128
	// Monetary represent account monetary type.
	Monetary Amount
}

// LookupAccounts fetchs one or more accounts by their ids alongside the monetary.
func (i *Instance) LookupAccounts(lookups []AccountLookup) ([]Account, error) {
	// Validate.
	countLookup := len(lookups)
	if countLookup > int(TigerBeetleMaxBatch) {
		return nil, ErrExceedsMaxTigerBeetleBatch
	}
	// Get TigerBeetle client instance.
	cln, err := i.Client()
	if err != nil {
		return nil, err
	}

	// Uint128 to TigerBeetle's Uint128.
	mapMonetary := make(map[string]Amount, countLookup)
	tbIds := make([]types.Uint128, 0, len(lookups))
	for _, lookup := range lookups {
		tbIds = append(tbIds, toBinding(lookup.ID))
		mapMonetary[lookup.ID.String()] = lookup.Monetary
	}
	defer func() {
		mapMonetary = nil
		tbIds = nil
	}()

	// Perform TigerBeetle LookupAccounts.
	tbAccounts, err := cln.LookupAccounts(tbIds)
	if err != nil {
		return nil, err
	}
	defer func() { tbAccounts = nil }()

	// Convert TigerBeetle's Account to Account.
	accounts := make([]Account, 0, len(tbAccounts))
	for _, account := range tbAccounts {
		// Get monetary.
		monetary, exists := mapMonetary[account.ID.String()]
		if !exists {
			log.Printf("[tbdb] Warning: monetary map for %s does not exists", account.ID.String())
			continue
		}

		// To TBDB's account.
		flags := account.AccountFlags()
		accounts = append(accounts, Account{
			ID:             fromBinding(account.ID),
			DebitsPending:  monetary.SetUint128Value(fromBinding(account.DebitsPending)),
			DebitsPosted:   monetary.SetUint128Value(fromBinding(account.DebitsPosted)),
			CreditsPending: monetary.SetUint128Value(fromBinding(account.CreditsPending)),
			CreditsPosted:  monetary.SetUint128Value(fromBinding(account.CreditsPosted)),
			UserData128:    fromBinding(account.UserData128),
			UserData64:     account.UserData64,
			UserData32:     account.UserData32,
			Reserved:       account.Reserved,
			Ledger:         account.Ledger,
			Code:           account.Code,
			Flags: AccountFlags{
				Linked:                     flags.Linked,
				DebitsMustNotExceedCredits: flags.DebitsMustNotExceedCredits,
				CreditsMustNotExceedDebits: flags.CreditsMustNotExceedDebits,
				History:                    flags.History,
				Imported:                   flags.Imported,
				Closed:                     flags.Closed,
			},
			Timestamp: account.Timestamp,
		})
	}
	return accounts, nil
}
