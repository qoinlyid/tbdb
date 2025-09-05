package tbdb

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"strconv"
	"strings"
	"sync"
)

// currency represents a currency definition with its code and decimal precision.
type currency struct {
	code    string
	decimal uint8
}

// === Begin Ledger Implement ===

const (
	minCurrencyCodeChar = 1
	maxCurrencyCodeChar = 6
	maxDecimal          = 99
)

// Compile-time check if *currency implements Ledger interface.
var _ Ledger = (*currency)(nil)

// Built-in currencies with their decimal precision.
var (
	// Vietnamese Dong (no decimals).
	VND *currency = newCurrency("VND", 0)
	// Indonesian Rupiah.
	IDR *currency = newCurrency("IDR", 2)
	// Malaysian Ringgit.
	MYR *currency = newCurrency("MYR", 2)
	// Singapore Dollar.
	SGD *currency = newCurrency("SGD", 2)
	// Thai Baht.
	THB *currency = newCurrency("THB", 2)
	// Philippine Peso.
	PHP *currency = newCurrency("PHP", 2)
	// US Dollar.
	USD *currency = newCurrency("USD", 2)
	// Euro.
	EUR *currency = newCurrency("EUR", 2)

	// Tether (6 decimals).
	USDT *currency = newCurrency("USDT", 6)
	// Bitcoin (8 decimals).
	BTC *currency = newCurrency("BTC", 8)
	// Binance Coin (12 decimals).
	BNB *currency = newCurrency("BNB", 12)
	// Ethereum (18 decimals).
	ETH *currency = newCurrency("ETH", 18)
)

// hashCurrencyCode creates a deterministic hash for currency codes > 3 characters.
// Uses FNV-1a algorithm to ensure consistent results.
func hashCurrencyCode(code string) uint32 {
	// FNV-1a offset basis.
	hash := uint32(2166136261)
	for _, c := range code {
		hash ^= uint32(c)
		// FNV-1a prime.
		hash *= 16777619
	}
	// Constrain to 6 digits (0-999999)
	return hash % 1000000
}

// newCurrency creates a new currency without validation.
// Used internally for built-in currencies.
func newCurrency(code string, decimal uint8) *currency {
	code = strings.ToUpper(strings.TrimSpace(code))
	return &currency{code: code, decimal: decimal}
}

// NewCurrency creates a new currency with validation.
// Returns error if code length or decimal precision is invalid.
//
//	cur, err := tbdb.NewCurrency("DOGE", 8)
//	if err != nil {
//		log.Println(err)
//	}
func NewCurrency(code string, decimal uint8) (*currency, error) {
	// Validate decimal range (0-99).
	if decimal > maxDecimal {
		return nil, fmt.Errorf("supported decimal is 0-%d", maxDecimal)
	}
	// Validate code length (1-6 characters).
	codelen := len(strings.TrimSpace(code))
	if codelen < minCurrencyCodeChar || codelen > maxCurrencyCodeChar {
		return nil, fmt.Errorf("supported currency char length is %d-%d", minCurrencyCodeChar, maxCurrencyCodeChar)
	}
	return newCurrency(code, decimal), nil
}

// EncodeLedger generates a unique LedgerCode for TigerBeetle.
// Format: [3-digit decimal + 100][6-digit currency encoding]
//
// Encoding strategy:
//   - Codes ≤3 chars: Direct base-100 encoding (reversible)
//   - Codes >3 chars: FNV hash + offset (deterministic but not reversible)
func (c *currency) EncodeLedger() LedgerCode {
	// Decimal part: 100-199 (represents 0-99 decimal precision).
	decimalPart := uint32(c.decimal) + 100

	var codeVal uint32
	if len(c.code) <= 3 {
		// Direct encoding for short codes (≤3 chars).
		// Each character encoded as 2 digits in base-100.
		// Result range: 0-353535 (for direct encoding).
		for i := range 3 {
			var n uint32
			if i < len(c.code) {
				char := c.code[i]
				switch {
				case char >= 'A' && char <= 'Z':
					n = uint32(char-'A') + 10 // A=10, B=11, ..., Z=35
				case char >= '0' && char <= '9':
					n = uint32(char - '0') // 0=0, 1=1, ..., 9=9
				default:
					n = 0 // Invalid chars treated as padding
				}
			} else {
				n = 0 // Padding for shorter codes
			}
			codeVal = codeVal*100 + n // Base-100 untuk ensure fit
		}
	} else {
		// Hash encoding for long codes (>3 chars).
		// Result range: 500000-999999 (hash + offset to avoid collision).
		hash := hashCurrencyCode(c.code)
		codeVal = 500000 + (hash % 500000)
	}

	// Combine: [3 digits][6 digits] = 9 digits total.
	return LedgerCode(decimalPart*1000000 + codeVal)
}

// DecodeLedger converts the LedgerCode back to human-readable format.
// Returns: "decimal=<precision>, currency=<code>"
//
// Decoding behavior:
//   - Direct encoded (value < 500000): Reverses the encoding to get original code
//   - Hash encoded (value ≥ 500000): Uses stored currency code (not reversible)
func (c *currency) DecodeLedger() string {
	ledgerCode := c.EncodeLedger()

	// Extract decimal and currency parts.
	decimal := int(ledgerCode/1000000) - 100
	codeValue := uint32(ledgerCode % 1000000)

	var currencyCode string
	if codeValue < 500000 {
		// Direct encoded: reverse the base-100 encoding.
		temp := codeValue
		chars := make([]uint32, 3)
		// Extract each character value (reverse order).
		for i := 2; i >= 0; i-- {
			chars[i] = temp % 100
			temp = temp / 100
		}

		// Convert character values back to letters/digits.
		var builder strings.Builder
		for _, charValue := range chars {
			if charValue == 0 {
				continue // Skip padding zeros
			}

			var ch byte
			switch {
			case charValue >= 1 && charValue <= 9:
				ch = byte('0' + charValue) // Numbers: 1-9 → '1'-'9'
			case charValue >= 10 && charValue <= 35:
				ch = byte('A' + charValue - 10) // Letters: 10-35 → 'A'-'Z'
			}
			if ch != 0 {
				builder.WriteByte(ch)
			}
		}
		currencyCode = builder.String()
	} else {
		// Hash encoded: use original currency code.
		// (Hash is not reversible, so we return the stored value).
		currencyCode = c.code
	}

	return fmt.Sprintf("decimal=%d, currency=%s", decimal, currencyCode)
}

// === End Ledger Implement ===

// === Begin Amount Implement ===

type amountCurrency struct {
	curr       *currency
	float64Val float64
	uint128Val Uint128
}

// NewMonetary creates new currency monetary.
func (c *currency) NewMonetary() *amountCurrency {
	return &amountCurrency{
		curr: c,
	}
}

// NewAmountFromFloat64 creates new currency amount from float64 value.
func (c *currency) NewAmountFromFloat64(val float64) *amountCurrency {
	return &amountCurrency{
		curr:       c,
		float64Val: val,
	}
}

// Compile-time check to ensure *currency implements Amount interface.
var _ Amount = (*amountCurrency)(nil)

// Pre-computed scale values for common decimal precisions (0-19).
// Provides O(1) lookup performance for frequently used decimal places.
var scaleCache = [31]uint64{
	1,                    // 0 decimals
	10,                   // 1 decimal
	100,                  // 2 decimals
	1000,                 // 3 decimals
	10000,                // 4 decimals
	100000,               // 5 decimals
	1000000,              // 6 decimals
	10000000,             // 7 decimals
	100000000,            // 8 decimals
	1000000000,           // 9 decimals
	10000000000,          // 10 decimals
	100000000000,         // 11 decimals
	1000000000000,        // 12 decimals
	10000000000000,       // 13 decimals
	100000000000000,      // 14 decimals
	1000000000000000,     // 15 decimals
	10000000000000000,    // 16 decimals
	100000000000000000,   // 17 decimals
	1000000000000000000,  // 18 decimals
	10000000000000000000, // 19 decimals
	// Extended for edge cases
	1000000000000000000, // 20 decimals (will overflow, but we handle it)
	1000000000000000000, // 21+ decimals (fallback to safe value)
	1000000000000000000, // ...continuing with safe fallback
	1000000000000000000,
	1000000000000000000,
	1000000000000000000,
	1000000000000000000,
	1000000000000000000,
	1000000000000000000,
	1000000000000000000,
	1000000000000000000,
}

// Pool for reusing big.Float instances
var bigFloatPool = sync.Pool{
	New: func() any {
		return new(big.Float)
	},
}

// Pool for big.Int operations
var bigIntPool = sync.Pool{
	New: func() any {
		return new(big.Int)
	},
}

// Pool for string builders to reduce GC pressure
var stringBuilderPool = sync.Pool{
	New: func() any {
		return &strings.Builder{}
	},
}

// scaleFromDecimals returns the scaling factor (10^decimals) for the given decimal precision.
// Uses pre-computed cache for common values (0-19) for optimal performance.
func scaleFromDecimals(decimals uint8) uint64 {
	if decimals < 31 {
		return scaleCache[decimals] // O(1) lookup, extended range
	}

	// Fallback for extreme cases (>30 decimals) - return safe value
	log.Printf("[tbdb] Warning: decimal precision %d exceeds cache, using fallback", decimals)
	return 1000000000000000000 // 10^18 as reasonable max
}

// formatWithCommas adds thousand separators to an integer value.
// Optimized for performance with minimal memory allocations.
func formatWithCommas(value int64) string {
	// Convert to string
	str := strconv.FormatInt(value, 10)

	// No formatting needed for numbers with 3 digits or less.
	if len(str) <= 3 {
		return str
	}

	// Calculate exact size needed for result
	commaCount := (len(str) - 1) / 3
	resultSize := len(str) + commaCount

	// Single allocation with exact capacity
	result := make([]byte, 0, resultSize)

	// Build result with commas in single pass
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(digit))
	}
	return string(result)
}

// formatCurrency formats a float64 value as a currency string with thousand separators.
// Handles decimal precision according to the specified decimal places.
func formatCurrency(value float64, decimal uint8) string {
	// Handle special floating-point values.
	if math.IsInf(value, 0) || math.IsNaN(value) {
		return "0"
	}

	// Separate integer and fractional parts.
	intPart := int64(value)
	fracPart := value - float64(intPart)

	// Format integer part with thousand separators.
	intStr := formatWithCommas(intPart)

	// Return integer-only format if no decimal places required.
	if decimal == 0 {
		return intStr
	}

	// Calculate and format fractional part with proper rounding.
	scale := scaleFromDecimals(decimal)
	fracScaled := int64(math.Round(fracPart * float64(scale)))

	// Handle overflow in fractional part
	if fracScaled >= int64(scale) {
		intPart++
		fracScaled = 0
		intStr = formatWithCommas(intPart) // Recalculate if carry-over
	}

	// Pre-calculate result size to minimize allocations
	resultSize := len(intStr) + 1 + int(decimal) // intStr + "." + decimals

	// Build result with single allocation
	result := make([]byte, 0, resultSize)
	result = append(result, intStr...)
	result = append(result, '.')

	// Format fractional part with leading zeros
	fracStr := strconv.FormatInt(fracScaled, 10)

	// Add leading zeros if needed
	for i := len(fracStr); i < int(decimal); i++ {
		result = append(result, '0')
	}

	result = append(result, fracStr...)
	return string(result)
}

func (a *amountCurrency) clone() *amountCurrency {
	c := *a
	return &c
}

// SetUint128Value sets Uint128 value to the monetary.
func (a *amountCurrency) SetUint128Value(val Uint128) Amount {
	c := a.clone()
	c.uint128Val = val
	return c
}

// SetFloat64Value sets float64 value to the monetary.
func (a *amountCurrency) SetFloat64Value(val float64) Amount {
	c := a.clone()
	c.float64Val = val
	return c
}

// Float64ToUint128 converts a human-readable float64 to TigerBeetle's Uint128 format.
// Scales the value according to the currency's decimal precision.
//
// Example: For IDR (2 decimals), 20000.50 becomes 2000050 (scaled by 100).
func (a *amountCurrency) Float64ToUint128() (Uint128, error) {
	// TigerBeetle doesn't support negative amounts.
	val := a.float64Val
	if val < 0 {
		val = 0
	}

	// Handle invalid floating-point values.
	if math.IsInf(val, 0) || math.IsNaN(val) {
		return Uint128{}, nil // Return zero value
	}

	// Apply scaling with rounding to handle floating-point precision issues.
	scale := scaleFromDecimals(a.curr.decimal)
	scaledVal := val * float64(scale)
	roundedVal := math.Round(scaledVal)

	// Check for values exceeding uint64 maximum.
	if roundedVal > float64(math.MaxUint64) {
		return Uint128{}, fmt.Errorf("value %f too large for Uint128", val)
	}

	// Convert to uint64 dan then to Uint128
	intVal := uint64(roundedVal)
	return Uint128FromUint64(intVal), nil
}

// ToUint128FromValue returns Uint128 valu type depends on value that has been set.
//
// The priority is float64Val, if greather than 0 will parse it to Uint128.
// Otherwise will use uint128Val.
func (a *amountCurrency) ToUint128FromValue() Uint128 {
	if a.float64Val > 0 {
		uint128, err := a.Float64ToUint128()
		if err != nil {
			return Uint128{}
		}
		return uint128
	}
	return a.uint128Val
}

// Uint128ToFloat64 converts TigerBeetle's Uint128 to human-readable float64.
// Reverses the scaling applied in Float64ToUint128.
//
// Example: For IDR (2 decimals), 2000050 becomes 20000.50.
func (a *amountCurrency) Uint128ToFloat64() float64 {
	val := a.uint128Val
	bigInt := val.BigInt()

	// Fast path 1: Zero value
	if bigInt.Sign() == 0 {
		return 0.0
	}

	scale := scaleFromDecimals(a.curr.decimal)

	// Fast path 2: Small values that fit in uint64 (most common case)
	if bigInt.IsUint64() {
		intVal := bigInt.Uint64()

		// Ultra-fast path for no decimals
		if a.curr.decimal == 0 {
			return float64(intVal)
		}

		// Fast division for common decimal cases
		switch a.curr.decimal {
		case 1:
			return float64(intVal) / 10.0
		case 2:
			return float64(intVal) / 100.0
		case 6:
			return float64(intVal) / 1000000.0
		case 8:
			return float64(intVal) / 100000000.0
		case 18:
			return float64(intVal) / 1000000000000000000.0
		default:
			return float64(intVal) / float64(scale)
		}
	}

	// Slow path: Use pooled big.Float for large values
	floatVal := bigFloatPool.Get().(*big.Float)
	scaleFloat := bigFloatPool.Get().(*big.Float)
	result := bigFloatPool.Get().(*big.Float)
	defer func() {
		floatVal.SetInt64(0) // Reset before returning to pool
		scaleFloat.SetInt64(0)
		result.SetInt64(0)
		bigFloatPool.Put(floatVal)
		bigFloatPool.Put(scaleFloat)
		bigFloatPool.Put(result)
	}()

	floatVal.SetInt(bigInt)
	scaleFloat.SetUint64(scale)
	result.Quo(floatVal, scaleFloat)
	humanValue, _ := result.Float64()
	return humanValue
}

// Uint128ToString formats a Uint128 amount as a complete currency string.
// Includes currency code, thousand separators, and proper decimal formatting.
//
// Example: For IDR, 2000050 becomes "IDR 20,000.50".
func (a *amountCurrency) Uint128ToString() string {
	// Get pooled string builder
	builder := stringBuilderPool.Get().(*strings.Builder)
	defer func() {
		builder.Reset() // Clear before returning to pool
		stringBuilderPool.Put(builder)
	}()

	// Fast path for zero
	val := a.uint128Val
	bigInt := val.BigInt()
	if bigInt.Sign() == 0 {
		builder.Grow(len(a.curr.code) + 10) // Pre-allocate for "CODE 0.00"
		builder.WriteString(a.curr.code)
		builder.WriteByte(' ')
		builder.WriteByte('0')

		if a.curr.decimal > 0 {
			builder.WriteByte('.')
			for i := uint8(0); i < a.curr.decimal; i++ {
				builder.WriteByte('0')
			}
		}

		return builder.String()
	}

	// Get human value using optimized conversion.
	humanValue := a.Uint128ToFloat64()

	// Pre-allocate with estimated size.
	estimatedSize := len(a.curr.code) + 25 // Conservative estimate
	builder.Grow(estimatedSize)

	// Build result
	builder.WriteString(a.curr.code)
	builder.WriteByte(' ')
	builder.WriteString(formatCurrency(humanValue, a.curr.decimal))
	return builder.String()
}

// Add safely adds two Uint128 amounts with overflow protection.
// Uses fast path for small values, falls back to big.Int for large values.
func (a *amountCurrency) Add(b Uint128) (Uint128, error) {
	x := a.uint128Val
	aBigInt := x.BigInt()
	bBigInt := b.BigInt()

	// Fast path: both values fit in uint64 and no overflow risk.
	if aBigInt.IsUint64() && bBigInt.IsUint64() {
		aVal := aBigInt.Uint64()
		bVal := bBigInt.Uint64()
		if aVal <= math.MaxUint64-bVal {
			return Uint128FromUint64(aVal + bVal), nil
		}
	}

	// Slow path: with pooled big.Int.
	result := bigIntPool.Get().(*big.Int)
	defer bigIntPool.Put(result)
	result.Add(aBigInt, bBigInt)

	// Direct overflow check without string allocation
	if result.BitLen() > 128 {
		return Uint128{}, errors.New("addition overflow")
	}
	return Uint128FromBigInt(result)
}

// Sub safely subtracts b from a with underflow protection.
// Returns error if the result would be negative.
func (a *amountCurrency) Sub(b Uint128) (Uint128, error) {
	x := a.uint128Val
	aBigInt := x.BigInt()
	bBigInt := b.BigInt()

	// Fast path: both values fit in uint64.
	if aBigInt.IsUint64() && bBigInt.IsUint64() {
		aVal := aBigInt.Uint64()
		bVal := bBigInt.Uint64()
		if aVal < bVal {
			return Uint128{}, ErrNegativeOrNilBigInt
		}
		return Uint128FromUint64(aVal - bVal), nil
	}

	// Slow path: use big.Int arithmetic.
	if aBigInt.Cmp(bBigInt) < 0 {
		return Uint128{}, ErrNegativeOrNilBigInt
	}
	result := new(big.Int).Sub(aBigInt, bBigInt)
	return Uint128FromBigInt(result)
}

// Mul multiplies an amount by a floating-point multiplier.
// Commonly used for calculating fees or applying exchange rates.
func (a *amountCurrency) Mul(multiplier float64) (Uint128, error) {
	if multiplier < 0 {
		multiplier = 0
	}
	amount := a.uint128Val
	amountBigInt := amount.BigInt()

	// Get pooled big.Float instances
	multiplierBig := bigFloatPool.Get().(*big.Float)
	amountFloat := bigFloatPool.Get().(*big.Float)
	result := bigFloatPool.Get().(*big.Float)
	defer func() {
		multiplierBig.SetFloat64(0)
		amountFloat.SetInt64(0)
		result.SetFloat64(0)
		bigFloatPool.Put(multiplierBig)
		bigFloatPool.Put(amountFloat)
		bigFloatPool.Put(result)
	}()

	multiplierBig.SetFloat64(multiplier)
	amountFloat.SetInt(amountBigInt)
	result.Mul(amountFloat, multiplierBig)
	resultInt, _ := result.Int(nil)

	// Optimized overflow check
	if resultInt.BitLen() > 128 {
		return Uint128{}, errors.New("multiplication overflow")
	}
	return Uint128FromBigInt(resultInt)
}

// Div divides an amount by a floating-point divisor.
// Returns error for invalid divisors (zero, infinity, NaN).
func (a *amountCurrency) Div(divisor float64) (Uint128, error) {
	if divisor == 0 || math.IsInf(divisor, 0) || math.IsNaN(divisor) {
		return Uint128{}, errors.New("invalid divisor")
	}
	amount := a.uint128Val
	amountBigInt := amount.BigInt()

	// Get pooled big.Float instances
	divisorBig := bigFloatPool.Get().(*big.Float)
	amountFloat := bigFloatPool.Get().(*big.Float)
	result := bigFloatPool.Get().(*big.Float)
	defer func() {
		divisorBig.SetFloat64(0)
		amountFloat.SetInt64(0)
		result.SetFloat64(0)
		bigFloatPool.Put(divisorBig)
		bigFloatPool.Put(amountFloat)
		bigFloatPool.Put(result)
	}()

	divisorBig.SetFloat64(divisor)
	amountFloat.SetInt(amountBigInt)
	result.Quo(amountFloat, divisorBig)
	resultInt, _ := result.Int(nil)

	// Ensure result is non-negative.
	if resultInt.Sign() < 0 {
		return Uint128{}, ErrNegativeOrNilBigInt
	}
	return Uint128FromBigInt(resultInt)
}

// Percentage calculates a percentage of the given amount.
//
// Example: Percentage(1000, 0.7) returns 0.7% of 1000 = 7.
func (a *amountCurrency) Percentage(percent float64) (Uint128, error) {
	return a.Mul(percent / 100.0)
}

// Compare compares two Uint128 amounts.
// Uses fast path for values that fit in uint64.
// Returns:
//   - -1 if a < b
//   - 0 if a == b
//   - 1 if a > b.
func (a *amountCurrency) Compare(b Uint128) int {
	x := a.uint128Val
	aBigInt := x.BigInt()
	bBigInt := b.BigInt()

	// Fast path: both values fit in uint64.
	if aBigInt.IsUint64() && bBigInt.IsUint64() {
		aVal := aBigInt.Uint64()
		bVal := bBigInt.Uint64()
		switch {
		case aVal < bVal:
			return -1
		case aVal > bVal:
			return 1
		default:
			return 0
		}
	}

	// Slow path: full big.Int comparison.
	return aBigInt.Cmp(bBigInt)
}

// IsZero checks if the Uint128 value represents zero or negative (treated as zero).
func (a *amountCurrency) IsZero() bool {
	val := a.uint128Val
	bigInt := val.BigInt()
	return bigInt.Sign() <= 0
}

// Equal checks if two Uint128 amounts are equal.
func (a *amountCurrency) Equal(b Uint128) bool {
	return a.Compare(b) == 0
}

// GreaterThan checks if a > b.
func (a *amountCurrency) GreaterThan(b Uint128) bool {
	return a.Compare(b) > 0
}

// GreaterThanOrEqual checks if a >= b.
func (a *amountCurrency) GreaterThanOrEqual(b Uint128) bool {
	return a.Compare(b) >= 0
}

// LessThan checks if a < b.
func (a *amountCurrency) LessThan(b Uint128) bool {
	return a.Compare(b) < 0
}

// LessThanOrEqual checks if a <= b.
func (a *amountCurrency) LessThanOrEqual(b Uint128) bool {
	return a.Compare(b) <= 0
}

// Min returns the smaller of two Uint128 amounts.
func (a *amountCurrency) Min(b Uint128) Uint128 {
	if a.LessThan(b) {
		return a.uint128Val
	}
	return b
}

// Max returns the larger of two Uint128 amounts.
func (a *amountCurrency) Max(b Uint128) Uint128 {
	if a.GreaterThan(b) {
		return a.uint128Val
	}
	return b
}

// === End Amount Implement ===
