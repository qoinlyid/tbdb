package tbdb

import (
	"math/big"
	"testing"

	"github.com/qoinlyid/qore"
	"github.com/stretchr/testify/assert"
)

func TestCreatePendingTransfers(t *testing.T) {
	// Setup environment and instance
	t.Setenv(qore.CONFIG_USED_KEY, "./.env")
	i := New()
	err := i.Open()
	assert.NoError(t, err)
	defer i.Close()

	// Parse account IDs from hex
	debitAccount, err := Uint128FromHex("198c8cbea858163e614bda026cc5848")
	assert.NoError(t, err)
	creditAccount, err := Uint128FromHex("198c8caa0ed1b230d56b39be8add93c")
	assert.NoError(t, err)

	result, err := i.CreatePendingTransfers([]TransferData{
		{
			DebitAccountID:  debitAccount,
			CreditAccountID: creditAccount,
			Amount:          IDR.NewAmountFromFloat64(50_000_000),
			Ledger:          IDR,
		},
		{
			DebitAccountID:  debitAccount,
			CreditAccountID: creditAccount,
			Amount:          IDR.NewAmountFromFloat64(3000),
			Ledger:          IDR,
		},
		{
			DebitAccountID:  debitAccount,
			CreditAccountID: creditAccount,
			Amount:          IDR.NewAmountFromFloat64(1_452_000),
			Ledger:          IDR,
		},
		{
			DebitAccountID:  debitAccount,
			CreditAccountID: creditAccount,
			Amount:          IDR.NewAmountFromFloat64(3000),
			Ledger:          IDR,
		},
		// {
		// 	DebitAccountID:  debitAccount,
		// 	CreditAccountID: creditAccount,
		// 	Amount:          IDR.NewAmountFromFloat64(12300),
		// 	Ledger:          IDR,
		// },
	}, true, 1001)
	assert.NoError(t, err)
	assert.True(t, result.FailedCount == 0)
	for _, result := range result.Results {
		t.Logf(`
{
  "index": "%d",
  "id": "%s",
  "id_hex": "%s",
  "result": "%s"
}
		`,
			result.Index,
			result.ID.BigInt(),
			result.ID.String(),
			result.Result,
		)
	}
}

func TestResolvePendingTransfers(t *testing.T) {
	// Setup environment and instance
	t.Setenv(qore.CONFIG_USED_KEY, "./.env")
	i := New()
	err := i.Open()
	assert.NoError(t, err)
	defer i.Close()

	raw := []struct {
		id     string
		amount float64
	}{
		{id: "2122903876979618333636925725593202514", amount: 50_000_000.00},
		{id: "2122903876979618333636925725593202515", amount: 3000.00},
	}
	pendings := make([]PendingTransfer, 0, len(raw))
	for _, r := range raw {
		b := new(big.Int)
		b.SetString(r.id, 10)
		pendingID, _ := Uint128FromBigInt(b)
		pendings = append(pendings, PendingTransfer{
			State:     ResolvePendingStateVoid,
			PendingID: pendingID,
			Amount:    IDR.NewAmountFromFloat64(r.amount),
		})
	}

	result, err := i.ResolvePendingTransfers(pendings)
	assert.NoError(t, err)
	assert.True(t, result.FailedCount == 0)
	for _, result := range result.Results {
		t.Logf(`
{
  "index": "%d",
  "id": "%s",
  "id_hex": "%s",
  "result": "%s"
}
		`,
			result.Index,
			result.ID.BigInt(),
			result.ID.String(),
			result.Result,
		)
	}
}
