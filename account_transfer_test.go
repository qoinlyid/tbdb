package tbdb

import (
	"testing"
	"time"

	"github.com/qoinlyid/qore"
	"github.com/stretchr/testify/assert"
)

func TestGetHisotricalBalances_Asc(t *testing.T) {
	t.Setenv(qore.CONFIG_USED_KEY, "./.env")
	i := New()
	err := i.Open()
	assert.NoError(t, err)
	defer i.Close()

	// Prepare.
	loc, err := time.LoadLocation("Asia/Jakarta")
	assert.NoError(t, err)
	timeMin := time.Date(2025, 8, 21, 0, 0, 0, 0, loc)
	timeMax := time.Date(2025, 8, 21, 23, 59, 59, 999_999_999, loc)
	accountID, err := Uint128FromHex("198c8cbea858163e614bda026cc5848")
	assert.NoError(t, err)

	// Perform.
	balances, err := i.GetHisotricalBalances(AccountTransferFilter{
		TimeMin:    timeMin,
		TimeMax:    timeMax,
		AccountID:  accountID,
		UserData32: 1,
		Limit:      uint32(TigerBeetleMaxBatch),
		Monetary:   IDR.NewMonetary(),
	})
	assert.NoError(t, err)
	for _, balance := range balances {
		t.Logf(`
{
  "debits_pending": "%s",
  "debits_posted": "%s",
  "credits_pending": "%s",
  "credits_posted": "%s",
  "credits_posted_f64": "%.2f",
  "timestamp": "%d"
  "dateTime": "%s"
}
		`,
			balance.DebitsPending.Uint128ToString(),
			balance.DebitsPosted.Uint128ToString(),
			balance.CreditsPending.Uint128ToString(),
			balance.CreditsPosted.Uint128ToString(),
			balance.CreditsPosted.Uint128ToFloat64(),
			balance.Timestamp,
			time.Unix(0, int64(balance.Timestamp)).In(loc),
		)
	}
}

func TestGetHisotricalBalancesDesc_All(t *testing.T) {
	t.Setenv(qore.CONFIG_USED_KEY, "./.env")
	i := New()
	err := i.Open()
	assert.NoError(t, err)
	defer i.Close()

	// Prepare.
	loc, err := time.LoadLocation("Asia/Jakarta")
	assert.NoError(t, err)
	timeMin := time.Date(2025, 8, 21, 0, 0, 0, 0, loc)
	timeMax := time.Date(2025, 8, 21, 23, 59, 59, 999_999_999, loc)
	accountID, err := Uint128FromHex("198c8cbea858163e614bda026cc5848")
	assert.NoError(t, err)

	// Perform.
	runner := func(upper, lower time.Time, accountID Uint128) ([]AccountBalance, error) {
		return i.GetHisotricalBalances(AccountTransferFilter{
			TimeMin:    upper,
			TimeMax:    lower,
			AccountID:  accountID,
			UserData32: 1,
			Limit:      uint32(TigerBeetleMaxBatch),
			Monetary:   IDR.NewMonetary(),
			Flags:      AccountFilterFlags{Reversed: true},
		})
	}

	var (
		firstTimestamp int64
		lastTimestamp  int64
	)
	var countTotal int
	firstRun := true
	for {
		balances, err := runner(timeMin, timeMax, accountID)
		assert.NoError(t, err)
		if len(balances) == 0 {
			break
		}
		if firstRun {
			firstTimestamp = int64(balances[0].Timestamp)
		}
		lastTimestamp = int64(balances[len(balances)-1].Timestamp)
		countTotal = countTotal + len(balances)
		firstRun = false

		// Window modifier.
		timeMax = time.Unix(0, lastTimestamp-1)
	}

	// Result.
	t.Logf(`
{
  "first_timestamp": "%d",
  "first_date_time": "%s",
  "last_timestamp": "%d",
  "last_date_time": "%s",
  "total_count": "%d"
}
		`,
		firstTimestamp,
		time.Unix(0, firstTimestamp).In(loc),
		lastTimestamp,
		time.Unix(0, lastTimestamp).In(loc),
		countTotal,
	)
}

func TestGetAccountTransfers_Asc(t *testing.T) {
	t.Setenv(qore.CONFIG_USED_KEY, "./.env")
	i := New()
	err := i.Open()
	assert.NoError(t, err)
	defer i.Close()

	// Prepare.
	loc, err := time.LoadLocation("Asia/Jakarta")
	assert.NoError(t, err)
	timeMin := time.Date(2025, 8, 21, 0, 0, 0, 0, loc)
	timeMax := time.Date(2025, 8, 21, 23, 59, 59, 999_999_999, loc)
	accountID, err := Uint128FromHex("198c8cbea858163e614bda026cc5848")
	assert.NoError(t, err)

	// Perform.
	transfers, err := i.GetAccountTransfers(AccountTransferFilter{
		TimeMin:    timeMin,
		TimeMax:    timeMax,
		AccountID:  accountID,
		UserData32: 1,
		Limit:      uint32(TigerBeetleMaxBatch),
		Monetary:   IDR.NewMonetary(),
	})
	assert.NoError(t, err)
	for _, transfer := range transfers {
		t.Logf(`
{
  "id": "%s",
  "id_hex": "%s",
  "debit_account_id": "%s",
  "debit_account_id_hex": "%s",
  "credit_account_id": "%s",
  "credit_account_id_hex": "%s",
  "amount": "%s",
  "amount_f64": "%.2f",
  "pending_id": "%s",
  "pending_id_hex": "%s",
  "user_data_128": "%s",
  "user_data_64": "%d",
  "user_data_32": "%d",
  "timeout": "%d",
  "ledger": "%d",
  "code": "%d",
  "flags": "%s",
  "timestamp": "%d",
  "date_time": "%s",
}
		`,
			transfer.ID.BigInt(),
			transfer.ID.String(),
			transfer.DebitAccountID.BigInt(),
			transfer.DebitAccountID.String(),
			transfer.CreditAccountID.BigInt(),
			transfer.CreditAccountID.String(),
			transfer.Amount.Uint128ToString(),
			transfer.Amount.Uint128ToFloat64(),
			transfer.PendingID.BigInt(),
			transfer.PendingID.String(),
			transfer.UserData128.BigInt(),
			transfer.UserData64,
			transfer.UserData32,
			transfer.Timeout,
			transfer.Ledger,
			transfer.Code,
			transfer.Flags,
			transfer.Timestamp,
			time.Unix(0, int64(transfer.Timestamp)).In(loc),
		)
	}
}

func TestGetAccountStatements(t *testing.T) {
	t.Setenv(qore.CONFIG_USED_KEY, "./.env")
	i := New()
	err := i.Open()
	assert.NoError(t, err)
	defer i.Close()

	// Prepare.
	loc, err := time.LoadLocation("Asia/Jakarta")
	assert.NoError(t, err)
	timeMin := time.Date(2025, 8, 24, 0, 0, 0, 0, loc)
	timeMax := time.Date(2025, 8, 24, 23, 59, 59, 999_999_999, loc)
	accountID, err := Uint128FromHex("198c8cbea858163e614bda026cc5848")
	assert.NoError(t, err)

	// Perform.
	statements, err := i.GetAccountStatements(AccountTransferFilter{
		TimeMin:   timeMin,
		TimeMax:   timeMax,
		AccountID: accountID,
		// UserData32: 1,
		Limit: uint32(TigerBeetleMaxBatch),
		// Limit:    2,
		Monetary: IDR.NewMonetary(),
		// Flags:      AccountFilterFlags{Reversed: true},
	})
	assert.NoError(t, err)
	t.Logf("total statement count = %d", len(statements))
	for _, statement := range statements {
		t.Logf(`
{
  "id": "%s",
  "id_hex": "%s",
  "debit_account_id": "%s",
  "debit_account_id_hex": "%s",
  "credit_account_id": "%s",
  "credit_account_id_hex": "%s",
  "debit": "%.2f",
  "credit": "%.2f",
  "balance_before": "%.2f",
  "balance_after": "%.2f",
  "user_data_128": "%s",
  "user_data_64": "%d",
  "user_data_32": "%d",
  "ledger": "%d",
  "code": "%d",
  "flags": "%s",
  "timestamp": "%d",
  "date_time": "%s",
  "additional": %v
}
		`,
			statement.ID.BigInt(),
			statement.ID.String(),
			statement.DebitAccountID.BigInt(),
			statement.DebitAccountID.String(),
			statement.CreditAccountID.BigInt(),
			statement.CreditAccountID.String(),
			statement.Debit.Uint128ToFloat64(),
			statement.Credit.Uint128ToFloat64(),
			statement.BalanceBefore.Uint128ToFloat64(),
			statement.BalanceAfter.Uint128ToFloat64(),
			statement.UserData128.BigInt(),
			statement.UserData64,
			statement.UserData32,
			statement.Ledger,
			statement.Code,
			statement.Flags,
			statement.Timestamp,
			time.Unix(0, int64(statement.Timestamp)).In(loc),
			statement.Additional,
		)
	}
}
