# tbdb
A wrapper for TigerBeetle database client Go with enhanced interfaces for currency handling, ledger management, and financial operations. Can be used standalone or as a Qore dependency.

[![Go Report Card](https://goreportcard.com/badge/github.com/qoinlyid/tbdb)](https://goreportcard.com/report/github.com/qoinlyid/tbdb)
[![Go Reference](https://pkg.go.dev/badge/github.com/qoinlyid/tbdb.svg)](https://pkg.go.dev/github.com/qoinlyid/tbdb)

## Features

- **Interface-based Architecture**: Flexible `Ledger` and `Amount` interfaces for custom implementations
- **Built-in Currency Support**: Pre-configured support for major currencies (USD, EUR, etc.) and cryptocurrencies (BTC, ETH, USDT)
- **High Performance**: Optimized with sync.Pool, pre-computed scaling factors, and efficient `Uint128` operations
- **Account Management**: Simplified account creation with categories (Control, Balance, Income, Liabilities, Testing)
- **Transaction Support**: Standard transfers, pending transfers, and resolution mechanisms
- **Statement Generation**: Historical balances, transfer records, and account statements with custom enrichment
- **Batch Operations**: Leverage TigerBeetle's high throughput with batch processing (up to 8,189 items)

## Installation

Requirements:
- Go **1.21+**
- TigerBeetle Cluster Release **v0.16.57**

```bash
go get github.com/qoinlyid/tbdb
```

## Quick Start

### Standalone Usage

```go
package main

import (
  "log"
  "github.com/qoinlyid/tbdb"
)

func main() {
  // Create instance.
  instance := tbdb.New()

  // Open connection (new session).
  if err := instance.Open(); err != nil {
    log.Fatal(err)
  }
  defer func() { _ = instance.Close() }()

  // Create accounts with built-in category.
  result, err := instance.CreateAccountsWithCategory(
    tbdb.AccountCategoryBalance,
    tbdb.CreateAccount{
      Ledger: tbdb.USD,
      UserData128: tbdb.Uint128FromUint64(12345),
    },
  )
  if err != nil {
    log.Fatal(err)
  }

  // Create transfer.
  debitAccountID, err := tbdb.Uint128FromHex("198c8caa0ed1b230d56b39be8add93b")
  if err != nil {
    log.Fatal(err)
  }
  creditAccountID, err := tbdb.Uint128FromHex("198c8cbea858163e614bda026cc5848")
  if err != nil {
    log.Fatal(err)
  }
  transferResult, err := instance.CreateTransfers([]tbdb.TransferData{{
    DebitAccountID: debitAccountID,
    CreditAccountID: creditAccountID,
    Amount: tbdb.USD.NewAmountFromFloat64(150.00),
    Ledger: USD,
  }})
  // And much more...
}
```

### As Qore Dependency

```go
import "github.com/qoinlyid/tbdb"

type Module struct {
  Tbdb *tbdb.Instance
}

func InitModule() *Module {
  m := &Module{
    Tbdb: tbdb.New(),
  }
  // Another stuff...
  return m
}

// Open/Close handled automatically by Qore
```

## Configuration

### Configuration Reference

| Key | Description | Default |
|---------------------|-------------|---------|
| `TBDB_DEPENDENCY_PRIORITY` | Dependency priority for open/close order | `10` |
| `TBDB_CLUSTER_ID` | TigerBeetle cluster ID | `0` |
| `TBDB_ADDRESSES` | TigerBeetle node addresses (comma-separated for multi node replica) | `""` |

### Configuration Files

You can use `OS environment variables`, dotenv file `.env`, `.json` file, `.toml` file or `.yaml` file.
Example file:

**JSON Example (.json):**
```json
{
  "TBDB_CLUSTER_ID": 1,
  "TBDB_ADDRESSES": "127.0.0.1:3000"
}
```

**YAML Example (.yaml):**
```yaml
TBDB_CLUSTER_ID: 1
TBDB_ADDRESSES: "127.0.0.1:3000"
```

**Environment file (.env):**
```env
TBDB_CLUSTER_ID=1
TBDB_ADDRESSES=127.0.0.1:3000
```

To use a specific config file (Standalone mode):
```go
os.Setenv("QORE_CONFIG_USED", "./.env.json")
cache := cache.New()
```

## API Reference

### Currencies

Built-in currencies with proper decimal handling:

```go
// Fiat currencies
tbdb.USD  // 2 decimals
tbdb.SGD  // 2 decimals
tbdb.IDR  // 2 decimals

// Cryptocurrencies  
tbdb.BTC  // 8 decimals
tbdb.ETH  // 18 decimals
tbdb.USDT // 6 decimals

// Custom currency
cur, err := tbdb.NewCurrency("DOGE", 8)
```

### Amount Operations

```go
// Create amount
amount := tbdb.USD.NewAmountFromFloat64(1000.50)

// Operations
sum, err := amount.Add(otherAmount)
diff, err := amount.Sub(otherAmount) 
scaled, err := amount.Mul(1.5)
portion, err := amount.Percentage(10) // 10% = 100.05

// Comparisons
if amount.GreaterThan(minimum) {
  // Process payment
}

// Formatting
fmt.Println(amount.Uint128ToString()) // "USD 1,000.50"
```

### Account Categories

Pre-defined categories with appropriate flags:

- `AccountCategoryControl`: System account to used as controll account for source/destination accounts (History flag)
- `AccountCategoryBalance`: Asset accounts (History + DebitsMustNotExceedCredits)
- `AccountCategoryIncome`: Revenue accounts (History + DebitsMustNotExceedCredits)
- `AccountCategoryLiabilities`: Liability accounts (History flag)
- `AccountCategoryTesting`: Test accounts (no flags)

### Transfers

```go
// Standard transfer
instance.CreateTransfers([]tbdb.TransferData
  DebitAccountID:  fromAccount
  CreditAccountID: toAccount
  Amount:         amount
  Ledger:         tbdb.USD,
}})

// Pending transfer (2-phase commit)
code := 1
linked := true
instance.CreatePendingTransfers(transfers, linked, code)

// Resolve pending
instance.ResolvePendingTransfers([]tbdb.PendingTransfer{{
  PendingID: pendingID,
  Amount:    finalAmount,
  State:     tbdb.ResolvePendingStatePost, // or StateVoid
}})
```

### Querying

```go
// Account lookup
accounts, err := instance.LookupAccounts([]tbdb.AccountLookup{{
  ID:       accountID,
  Monetary: tbdb.USD.NewMonetary(),
}})

// Historical data
filter := tbdb.AccountTransferFilter{
  AccountID: accountID,
  TimeMin:   startTime,
  TimeMax:   endTime,
  Monetary:  tbdb.USD.NewMonetary(),
  Limit:     100,
}

balances, err := instance.GetHistoricalBalances(filter)
transfers, err := instance.GetAccountTransfers(filter)
statements, err := instance.GetAccountStatements(filter, enrichmentFunc)
```

## Custom Implementations

### Custom Ledger

```go
type CustomLedger struct {
  category string
  version  uint8
}

func (l *CustomLedger) EncodeLedger() tbdb.LedgerCode {
  // Your encoding logic...
  return tbdb.LedgerCode(encoded)
}

func (l *CustomLedger) DecodeLedger() string {
  return fmt.Sprintf("category=%s, version=%d", l.category, l.version)
}
```

### Custom Amount

```go
type TokenAmount struct {
  decimals uint8
  symbol   string
}

func (t *TokenAmount) Float64ToUint128() (tbdb.Uint128, error) {
  // Custom conversion logic...
}

func (t *TokenAmount) Uint128ToString() string {
  // Custom formatting...
}

// Implement remaining Amount interface methods...
```

## Performance Optimizations

- **Object Pooling**: Reuses `big.Int`, `big.Float`, and `strings.Buildere` instances
- **Pre-computed Scales**: O(1) lookup for decimal scaling factors
- **Fast Paths**: Optimized paths for uint64-sized values
- **Batch Processing**: Process up to 8,189 items in a single operation
- **Efficient Uint128**: Custom implementation avoiding unnecessary heap allocations.

## Error Handling

The package provides detailed error types:

```go
var
  ErrClientNil                  // TigerBeetle client not
  ErrMonetaryMustNotBeNil       // Missing monetary
  ErrExceedsMaxTigerBeetleBatch // Batch size > 8,
  ErrNegativeOrNilBigInt        // Invalid amount value
  // ... more specific errors
)
```
