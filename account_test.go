package tbdb

import (
	"testing"

	"github.com/qoinlyid/qore"
	"github.com/stretchr/testify/assert"
)

func TestCreateCreateAccountsWithCategory(t *testing.T) {
	t.Setenv(qore.CONFIG_USED_KEY, "./.env")
	i := New()
	err := i.Open()
	assert.NoError(t, err)
	defer i.Close()

	res, err := i.CreateAccountsWithCategory(
		// AccountCategoryControl,
		// CreateAccount{Ledger: IDR, UserData32: 0},
		// CreateAccount{Ledger: IDR, UserData32: 1},
		AccountCategoryBalance,
		CreateAccount{Ledger: IDR, UserData64: 1},
		CreateAccount{Ledger: IDR, UserData64: 2},
		// CreateAccount{Ledger: IDR, UserData64: 3},
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.True(t, res.FailedCount == 0)
	for _, result := range res.Results {
		t.Logf(
			"Idx: %d, IDRaw: %s, IDStr: %s",
			result.Index, result.ID.BigInt(), result.ID.String(),
		)
	}
}

func TestLookupAccounts(t *testing.T) {
	t.Setenv(qore.CONFIG_USED_KEY, "./.env")
	i := New()
	err := i.Open()
	assert.NoError(t, err)
	defer i.Close()

	accIDHexs := []string{
		"198c8caa0ed1b230d56b39be8add93b",
		"198c8caa0ed1b230d56b39be8add93c",
		"198c8cbea858163e614bda026cc5848",
		"198c8cbea858163e614bda026cc5849",
	}
	parseAccIDFn := func() (looks []AccountLookup) {
		for _, hex := range accIDHexs {
			id, err := Uint128FromHex(hex)
			if err != nil {
				continue
			}
			looks = append(looks, AccountLookup{ID: id, Monetary: IDR.NewMonetary()})
		}
		return
	}
	accounts, err := i.LookupAccounts(parseAccIDFn())
	assert.NoError(t, err)
	for _, account := range accounts {
		t.Logf(`
{
  "id": "%s",
  "debits_pending": "%s",
  "debits_posted": "%s",
  "credits_pending": "%s",
  "credits_posted": "%s",
  "credits_posted_f64": "%.2f",
  "user_data_128": "%s",
  "user_data_64": "%d",
  "user_data_32": "%d",
  "ledger": "%d",
  "code": "%d",
  "flags": "%v",
  "timestamp": "%d",
}
		`,
			account.ID.BigInt(),
			account.DebitsPending.Uint128ToString(),
			account.DebitsPosted.Uint128ToString(),
			account.CreditsPending.Uint128ToString(),
			account.CreditsPosted.Uint128ToString(),
			account.CreditsPosted.Uint128ToFloat64(),
			account.UserData128.BigInt(),
			account.UserData64,
			account.UserData32,
			account.Ledger,
			account.Code,
			account.Flags,
			account.Timestamp,
		)
	}
}
