package tbdb

import (
	"bufio"
	"encoding/json"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/qoinlyid/qore"
	"github.com/stretchr/testify/assert"
)

func BenchmarkCreateTransferPendings_Many(b *testing.B) {
	// Setup environment and instance
	b.Setenv(qore.CONFIG_USED_KEY, "./.env")
	instance := New()
	instance.Open()
	defer instance.Close()

	// Parse account IDs from hex
	debitAccount, err := Uint128FromHex("198c8caa0ed1b230d56b39be8add93b") // Control account
	assert.NoError(b, err)
	creditAccount, err := Uint128FromHex("198c8cbea858163e614bda026cc5848") // Credit account
	assert.NoError(b, err)

	// Random source for amounts (IDR 10,000 - 232,500)
	randSource := rand.New(rand.NewSource(time.Now().UnixNano()))
	const (
		minAmountIDR = 10_000.0  // IDR 10,000
		maxAmountIDR = 232_500.0 // IDR 232,500
	)

	// Base transfer template with your actual values
	baseTransfer := TransferData{
		DebitAccountID:  debitAccount,
		CreditAccountID: creditAccount,
		Ledger:          IDR,
	}

	// Helper function to generate transfer with random amount only
	generateTransfer := func() TransferData {
		transfer := baseTransfer
		transfer.ID = NewID() // Generate unique ID

		// Only randomize the Amount field
		randomAmount := minAmountIDR + randSource.Float64()*(maxAmountIDR-minAmountIDR)
		transfer.Amount = IDR.NewAmountFromFloat64(randomAmount)

		b.Logf(`{"id": %s, "amount": %.2f}`, transfer.ID.BigInt(), randomAmount)
		return transfer
	}

	// Constants for 10M transfers
	const (
		totalTransfers = 1_500_000
		batchSize      = 1_000
		totalBatches   = totalTransfers / batchSize
	)

	// Pre-allocate batch slice to reduce allocations
	batch := make([]TransferData, batchSize)

	// Start timing after setup
	b.ResetTimer()
	b.ReportAllocs()

	// Track progress
	var processed int64
	lastReport := time.Now()
	startTime := time.Now()

	// Process exactly 10M transfers in batches using modern range over int
	for batchNum := range totalBatches {
		// Generate batch of transfers with random amounts using modern range
		for j := range batchSize {
			batch[j] = generateTransfer()
		}

		// Execute CreateTransferPendings
		result, err := instance.CreatePendingTransfers(
			batch,
			false, // not linked
			1001,  // transfer code
		)
		if err != nil {
			b.Fatalf("CreateTransferPendings failed at batch %d: %v", batchNum, err)
		} else if result.FailedCount > 0 {
			b.Logf("Succeed: %d, Failed: %d", result.SuccessCount, result.FailedCount)
		}

		processed += batchSize

		// Progress reporting every 1M transfers
		if processed%1_000_000 == 0 {
			now := time.Now()
			elapsed := now.Sub(lastReport)
			rate := float64(1_000_000) / elapsed.Seconds()

			b.Logf("Processed %dM transfers (%.0f transfers/sec, total elapsed: %v)",
				processed/1_000_000, rate, now.Sub(startTime))
			lastReport = now
		}
	}

	// Final statistics
	b.StopTimer()
	totalElapsed := time.Since(startTime)
	overallRate := float64(processed) / totalElapsed.Seconds()

	b.Logf("Benchmark completed: %d total transfers processed in %v (%.0f transfers/sec overall)",
		processed, totalElapsed, overallRate)
}

func BenchmarkResolvePendingTransfers_Many(b *testing.B) {
	// Inner type os JSON dataset.
	type jsonRecord struct {
		PendingID string  `json:"id"`
		Amount    float64 `json:"amount"`
	}

	// Reader.
	readJSONBatch := func(scanner *bufio.Scanner, batchSize int, state ResolvePendingState) ([]PendingTransfer, error) {
		batch := make([]PendingTransfer, 0, batchSize)
		for len(batch) < batchSize && scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || line == "[" || line == "]" || line == "{" || line == "}" {
				continue
			}
			line = strings.TrimSuffix(line, ",")

			var record jsonRecord
			if err := json.Unmarshal([]byte(line), &record); err != nil {
				continue
			}

			bi := new(big.Int)
			bi.SetString(record.PendingID, 10)
			pendingID, err := Uint128FromBigInt(bi)
			if err != nil {
				continue
			}
			pending := PendingTransfer{
				State:     state,
				PendingID: pendingID,
				Amount:    IDR.NewAmountFromFloat64(record.Amount),
			}
			batch = append(batch, pending)
		}

		return batch, scanner.Err()
	}

	// Open dataset.
	dataset := "tb_dataset.json"
	f, err := os.Open(dataset)
	if err != nil {
		b.Fatalf("Cannot open %s: %v", dataset, err)
	}
	defer f.Close()
	if finfo, _ := f.Stat(); finfo != nil {
		b.Logf("Processing JSON file: %.2f MB", float64(finfo.Size())/(1024*1024))
	}

	// Setup scanner with large buffer
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	b.ResetTimer()
	b.ReportAllocs()

	// Perform.
	const batchSize = 1000
	var totalProcessed int64
	startTime := time.Now()
	lastReport := startTime
	b.Setenv(qore.CONFIG_USED_KEY, "./.env")
	i := New()
	if err := i.Open(); err != nil {
		b.Fatalf("Failed to open TigerBeetle connection: %s", err.Error())
	}
	defer i.Close()
	for {
		batch, err := readJSONBatch(scanner, batchSize, ResolvePendingStatePost)
		if err != nil {
			b.Fatalf("Error reading JSON: %v", err)
		}
		if len(batch) == 0 {
			break
		}

		// Execute ResolvePendingTransfers.
		result, err := i.ResolvePendingTransfers(batch)
		if err != nil {
			b.Fatalf("ResolvePendingTransfers failed: %v", err)
		} else if result.FailedCount > 0 {
			b.Logf("Got failed resolve %d", result.FailedCount)
		}

		totalProcessed += int64(len(batch))
		if totalProcessed%100_000 == 0 {
			now := time.Now()
			elapsed := now.Sub(lastReport)
			rate := float64(100_000) / elapsed.Seconds()

			b.Logf("Processed %dK transfers (%.0f transfers/sec)",
				totalProcessed/1000, rate)
			lastReport = now
		}
	}

	// Final stats
	b.StopTimer()
	totalElapsed := time.Since(startTime)
	overallRate := float64(totalProcessed) / totalElapsed.Seconds()

	b.Logf("Completed: %d transfers in %v (%.0f transfers/sec overall)",
		totalProcessed, totalElapsed, overallRate)
}
