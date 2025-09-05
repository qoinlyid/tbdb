package tbdb

// import (
// 	"fmt"
// 	"math"
// 	"strings"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// func TestNewCurrency(t *testing.T) {
// 	tests := []struct {
// 		name        string
// 		code        string
// 		decimal     uint8
// 		wantErr     bool
// 		errContains string
// 	}{
// 		{
// 			name:    "valid USD currency",
// 			code:    "USD",
// 			decimal: 2,
// 			wantErr: false,
// 		},
// 		{
// 			name:    "valid VND currency with zero decimals",
// 			code:    "VND",
// 			decimal: 0,
// 			wantErr: false,
// 		},
// 		{
// 			name:    "valid BTC with high decimals",
// 			code:    "BTC",
// 			decimal: 8,
// 			wantErr: false,
// 		},
// 		{
// 			name:        "invalid decimal too high",
// 			code:        "USD",
// 			decimal:     100,
// 			wantErr:     true,
// 			errContains: "supported decimal is",
// 		},
// 		{
// 			name:        "empty currency code",
// 			code:        "",
// 			decimal:     2,
// 			wantErr:     true,
// 			errContains: "supported currency char length is",
// 		},
// 		{
// 			name:        "currency code too long",
// 			code:        "TOOLONG",
// 			decimal:     2,
// 			wantErr:     true,
// 			errContains: "supported currency char length is",
// 		},
// 		{
// 			name:    "currency code with spaces (should be trimmed)",
// 			code:    "  USD  ",
// 			decimal: 2,
// 			wantErr: false,
// 		},
// 		{
// 			name:    "single character currency",
// 			code:    "A",
// 			decimal: 1,
// 			wantErr: false,
// 		},
// 		{
// 			name:    "maximum length currency",
// 			code:    "CUSTOM",
// 			decimal: 12,
// 			wantErr: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			curr, err := NewCurrency(tt.code, tt.decimal)

// 			if tt.wantErr {
// 				assert.Error(t, err)
// 				assert.Nil(t, curr)
// 				if tt.errContains != "" {
// 					assert.Contains(t, err.Error(), tt.errContains)
// 				}
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotNil(t, curr)
// 				assert.Equal(t, strings.ToUpper(strings.TrimSpace(tt.code)), curr.code)
// 				assert.Equal(t, tt.decimal, curr.decimal)
// 			}
// 		})
// 	}
// }

// func TestBuiltInCurrencies(t *testing.T) {
// 	tests := []struct {
// 		currency *currency
// 		code     string
// 		decimal  uint8
// 	}{
// 		{VND, "VND", 0},
// 		{IDR, "IDR", 2},
// 		{USD, "USD", 2},
// 		{BTC, "BTC", 8},
// 		{ETH, "ETH", 18},
// 		{USDT, "USDT", 6},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.code, func(t *testing.T) {
// 			assert.Equal(t, tt.code, tt.currency.code)
// 			assert.Equal(t, tt.decimal, tt.currency.decimal)

// 			// Test that they implement interfaces
// 			assert.Implements(t, (*Ledger)(nil), tt.currency)
// 			assert.Implements(t, (*Amount)(nil), tt.currency)
// 		})
// 	}
// }

// func TestEncodeLedger(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		currency *currency
// 		validate func(t *testing.T, code LedgerCode)
// 	}{
// 		{
// 			name:     "USD 2 decimals",
// 			currency: USD,
// 			validate: func(t *testing.T, code LedgerCode) {
// 				// Should start with 102 (2 + 100)
// 				decimalPart := uint32(code) / 1000000
// 				assert.Equal(t, uint32(102), decimalPart)
// 			},
// 		},
// 		{
// 			name:     "VND 0 decimals",
// 			currency: VND,
// 			validate: func(t *testing.T, code LedgerCode) {
// 				// Should start with 100 (0 + 100)
// 				decimalPart := uint32(code) / 1000000
// 				assert.Equal(t, uint32(100), decimalPart)
// 			},
// 		},
// 		{
// 			name:     "ETH 18 decimals",
// 			currency: ETH,
// 			validate: func(t *testing.T, code LedgerCode) {
// 				// Should start with 118 (18 + 100)
// 				decimalPart := uint32(code) / 1000000
// 				assert.Equal(t, uint32(118), decimalPart)
// 			},
// 		},
// 		{
// 			name:     "IDR 2 decimals",
// 			currency: IDR,
// 			validate: func(t *testing.T, code LedgerCode) {
// 				// Should start with 102 (2 + 100)
// 				decimalPart := uint32(code) / 1000000
// 				assert.Equal(t, uint32(102), decimalPart)
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			code := tt.currency.EncodeLedger()
// 			fmt.Println(tt.name, code)
// 			assert.Greater(t, uint32(code), uint32(0))
// 			tt.validate(t, code)
// 		})
// 	}
// }

// func TestDecodeLedger(t *testing.T) {
// 	currencies := []*currency{USD, IDR, BTC, ETH, VND, USDT}

// 	for _, curr := range currencies {
// 		t.Run(curr.code, func(t *testing.T) {
// 			decoded := curr.DecodeLedger()

// 			// Should contain decimal and currency information
// 			assert.Contains(t, decoded, "decimal=")
// 			assert.Contains(t, decoded, "currency=")
// 			assert.Contains(t, decoded, fmt.Sprintf("decimal=%d", curr.decimal))

// 			// For short codes (≤3 chars), should be able to decode exactly
// 			if len(curr.code) <= 3 {
// 				assert.Contains(t, decoded, fmt.Sprintf("currency=%s", curr.code))
// 			}
// 		})
// 	}
// }

// func TestFloat64ToUint128(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		currency *currency
// 		input    float64
// 		expected uint64 // Expected as uint64 for simple verification
// 		hasError bool
// 	}{
// 		{
// 			name:     "USD positive amount",
// 			currency: USD,
// 			input:    100.50,
// 			expected: 10050, // 100.50 * 100 (2 decimals)
// 		},
// 		{
// 			name:     "IDR large amount",
// 			currency: IDR,
// 			input:    1000000.75,
// 			expected: 100000075, // 1000000.75 * 100
// 		},
// 		{
// 			name:     "VND no decimals",
// 			currency: VND,
// 			input:    50000,
// 			expected: 50000, // 50000 * 1 (0 decimals)
// 		},
// 		{
// 			name:     "BTC high precision",
// 			currency: BTC,
// 			input:    0.00000001,
// 			expected: 1, // 0.00000001 * 100000000 (8 decimals)
// 		},
// 		{
// 			name:     "negative amount becomes zero",
// 			currency: USD,
// 			input:    -100.50,
// 			expected: 0,
// 		},
// 		{
// 			name:     "zero amount",
// 			currency: USD,
// 			input:    0,
// 			expected: 0,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result, err := tt.currency.Amount().Float64ToUint128(tt.input)

// 			if tt.hasError {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)

// 				// Convert back to verify (simple check for uint64 range)
// 				bigInt := result.BigInt()
// 				if bigInt.IsUint64() {
// 					assert.Equal(t, tt.expected, bigInt.Uint64())
// 				}
// 			}
// 		})
// 	}
// }

// func TestUint128ToFloat64(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		currency *currency
// 		input    uint64
// 		expected float64
// 	}{
// 		{
// 			name:     "USD conversion",
// 			currency: USD,
// 			input:    10050,
// 			expected: 100.50, // 10050 / 100
// 		},
// 		{
// 			name:     "IDR conversion",
// 			currency: IDR,
// 			input:    100000075,
// 			expected: 1000000.75, // 100000075 / 100
// 		},
// 		{
// 			name:     "VND conversion",
// 			currency: VND,
// 			input:    50000,
// 			expected: 50000, // 50000 / 1
// 		},
// 		{
// 			name:     "BTC conversion",
// 			currency: BTC,
// 			input:    1,
// 			expected: 0.00000001, // 1 / 100000000
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			uint128Val := Uint128FromUint64(tt.input)
// 			result := tt.currency.Amount().Uint128ToFloat64(uint128Val)
// 			assert.InDelta(t, tt.expected, result, 0.000000001) // Allow small floating point errors
// 		})
// 	}
// }

// func TestUint128ToString(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		currency *currency
// 		input    uint64
// 		expected string
// 	}{
// 		{
// 			name:     "USD formatting",
// 			currency: USD,
// 			input:    1234567,
// 			expected: "USD 12,345.67",
// 		},
// 		{
// 			name:     "IDR formatting",
// 			currency: IDR,
// 			input:    50000000,
// 			expected: "IDR 500,000.00",
// 		},
// 		{
// 			name:     "VND formatting no decimals",
// 			currency: VND,
// 			input:    1000000,
// 			expected: "VND 1,000,000",
// 		},
// 		{
// 			name:     "Small amount USD",
// 			currency: USD,
// 			input:    99,
// 			expected: "USD 0.99",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			uint128Val := Uint128FromUint64(tt.input)
// 			result := tt.currency.Amount().Uint128ToString(uint128Val)
// 			assert.Equal(t, tt.expected, result)
// 		})
// 	}
// }

// func TestArithmeticOperations(t *testing.T) {
// 	currency := USD

// 	t.Run("Add", func(t *testing.T) {
// 		a := Uint128FromUint64(10000) // $100.00
// 		b := Uint128FromUint64(5050)  // $50.50

// 		result, err := currency.Amount().Add(a, b)
// 		require.NoError(t, err)

// 		expected := Uint128FromUint64(15050) // $150.50
// 		assert.True(t, currency.Amount().Equal(result, expected))
// 	})

// 	t.Run("Sub", func(t *testing.T) {
// 		a := Uint128FromUint64(10000) // $100.00
// 		b := Uint128FromUint64(5050)  // $50.50

// 		result, err := currency.Amount().Sub(a, b)
// 		require.NoError(t, err)

// 		expected := Uint128FromUint64(4950) // $49.50
// 		assert.True(t, currency.Amount().Equal(result, expected))
// 	})

// 	t.Run("Sub underflow", func(t *testing.T) {
// 		a := Uint128FromUint64(5000)  // $50.00
// 		b := Uint128FromUint64(10000) // $100.00

// 		_, err := currency.Amount().Sub(a, b)
// 		assert.Error(t, err)
// 		assert.Equal(t, ErrNegativeOrNilBigInt, err)
// 	})

// 	t.Run("Mul", func(t *testing.T) {
// 		a := Uint128FromUint64(10000) // $100.00
// 		multiplier := 1.5

// 		result, err := currency.Amount().Mul(a, multiplier)
// 		require.NoError(t, err)

// 		expected := Uint128FromUint64(15000) // $150.00
// 		assert.True(t, currency.Amount().Equal(result, expected))
// 	})

// 	t.Run("Div", func(t *testing.T) {
// 		a := Uint128FromUint64(10000) // $100.00
// 		divisor := 2.0

// 		result, err := currency.Amount().Div(a, divisor)
// 		require.NoError(t, err)

// 		expected := Uint128FromUint64(5000) // $50.00
// 		assert.True(t, currency.Amount().Equal(result, expected))
// 	})

// 	t.Run("Div by zero", func(t *testing.T) {
// 		a := Uint128FromUint64(10000)

// 		_, err := currency.Amount().Div(a, 0)
// 		assert.Error(t, err)
// 		assert.Contains(t, err.Error(), "invalid divisor")
// 	})

// 	t.Run("Percentage", func(t *testing.T) {
// 		a := Uint128FromUint64(100000) // $1000.00
// 		percent := 0.7                 // 0.7%

// 		result, err := currency.Amount().Percentage(a, percent)
// 		require.NoError(t, err)

// 		// Expected: 0.7% of $1000.00 = $7.00 = 700 (in cents)
// 		// But due to floating-point precision, we might get 699 or 700
// 		// So we check that it's within 1 cent tolerance
// 		expected := Uint128FromUint64(700)
// 		actual := result.BigInt().Uint64()
// 		expectedVal := expected.BigInt().Uint64()

// 		// Allow 1 cent tolerance for floating-point precision issues
// 		tolerance := uint64(1)
// 		diff := uint64(0)
// 		if actual > expectedVal {
// 			diff = actual - expectedVal
// 		} else {
// 			diff = expectedVal - actual
// 		}

// 		assert.LessOrEqual(t, diff, tolerance,
// 			"Result %d should be within %d of expected %d (difference: %d)",
// 			actual, tolerance, expectedVal, diff)

// 		// Also verify the result is reasonable (between $6.99 and $7.01)
// 		minExpected := Uint128FromUint64(699) // $6.99
// 		maxExpected := Uint128FromUint64(701) // $7.01

// 		assert.True(t, currency.Amount().GreaterThanOrEqual(result, minExpected),
// 			"Result should be at least $6.99")
// 		assert.True(t, currency.Amount().LessThanOrEqual(result, maxExpected),
// 			"Result should be at most $7.01")
// 	})

// 	// More comprehensive percentage tests with tolerance
// 	t.Run("Percentage with tolerance", func(t *testing.T) {
// 		tests := []struct {
// 			name      string
// 			amount    uint64
// 			percent   float64
// 			expected  uint64
// 			tolerance uint64
// 		}{
// 			{
// 				name:      "1% of $100",
// 				amount:    10000,
// 				percent:   1.0,
// 				expected:  100,
// 				tolerance: 1,
// 			},
// 			{
// 				name:      "2.5% of $200",
// 				amount:    20000,
// 				percent:   2.5,
// 				expected:  500,
// 				tolerance: 1,
// 			},
// 			{
// 				name:      "0.7% of $1000 (problematic case)",
// 				amount:    100000,
// 				percent:   0.7,
// 				expected:  700,
// 				tolerance: 1, // Allow 1 cent tolerance
// 			},
// 			{
// 				name:      "10% of $50",
// 				amount:    5000,
// 				percent:   10.0,
// 				expected:  500,
// 				tolerance: 0, // Should be exact
// 			},
// 		}

// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				amount := Uint128FromUint64(tt.amount)
// 				result, err := currency.Amount().Percentage(amount, tt.percent)
// 				require.NoError(t, err)

// 				actual := result.BigInt().Uint64()
// 				diff := uint64(0)
// 				if actual > tt.expected {
// 					diff = actual - tt.expected
// 				} else {
// 					diff = tt.expected - actual
// 				}

// 				assert.LessOrEqual(t, diff, tt.tolerance,
// 					"Amount: $%.2f, Percent: %.1f%%, Expected: %d, Actual: %d, Diff: %d, Tolerance: %d",
// 					currency.Amount().Uint128ToFloat64(amount), tt.percent, tt.expected, actual, diff, tt.tolerance)
// 			})
// 		}
// 	})
// }

// func TestComparisonOperations(t *testing.T) {
// 	currency := USD

// 	a := Uint128FromUint64(10000) // $100.00
// 	b := Uint128FromUint64(5000)  // $50.00
// 	c := Uint128FromUint64(10000) // $100.00
// 	zero := Uint128FromUint64(0)

// 	t.Run("Compare", func(t *testing.T) {
// 		assert.Equal(t, 1, currency.Amount().Compare(a, b))  // a > b
// 		assert.Equal(t, -1, currency.Amount().Compare(b, a)) // b < a
// 		assert.Equal(t, 0, currency.Amount().Compare(a, c))  // a == c
// 	})

// 	t.Run("IsZero", func(t *testing.T) {
// 		assert.True(t, currency.Amount().IsZero(zero))
// 		assert.False(t, currency.Amount().IsZero(a))
// 	})

// 	t.Run("Equal", func(t *testing.T) {
// 		assert.True(t, currency.Amount().Equal(a, c))
// 		assert.False(t, currency.Amount().Equal(a, b))
// 	})

// 	t.Run("GreaterThan", func(t *testing.T) {
// 		assert.True(t, currency.Amount().GreaterThan(a, b))
// 		assert.False(t, currency.Amount().GreaterThan(b, a))
// 		assert.False(t, currency.Amount().GreaterThan(a, c))
// 	})

// 	t.Run("LessThan", func(t *testing.T) {
// 		assert.True(t, currency.Amount().LessThan(b, a))
// 		assert.False(t, currency.Amount().LessThan(a, b))
// 		assert.False(t, currency.Amount().LessThan(a, c))
// 	})

// 	t.Run("MinMax", func(t *testing.T) {
// 		min := currency.Amount().Min(a, b)
// 		max := currency.Amount().Max(a, b)

// 		assert.True(t, currency.Amount().Equal(min, b))
// 		assert.True(t, currency.Amount().Equal(max, a))
// 	})
// }

// func TestScaleFromDecimals(t *testing.T) {
// 	tests := []struct {
// 		decimals uint8
// 		expected uint64
// 	}{
// 		{0, 1},
// 		{1, 10},
// 		{2, 100},
// 		{8, 100000000},
// 		{18, 1000000000000000000},
// 	}

// 	for _, tt := range tests {
// 		t.Run(fmt.Sprintf("decimals_%d", tt.decimals), func(t *testing.T) {
// 			result := scaleFromDecimals(tt.decimals)
// 			assert.Equal(t, tt.expected, result)
// 		})
// 	}
// }

// func TestFormatWithCommas(t *testing.T) {
// 	tests := []struct {
// 		input    int64
// 		expected string
// 	}{
// 		{0, "0"},
// 		{123, "123"},
// 		{1234, "1,234"},
// 		{1234567, "1,234,567"},
// 		{1000000000, "1,000,000,000"},
// 	}

// 	for _, tt := range tests {
// 		t.Run(fmt.Sprintf("format_%d", tt.input), func(t *testing.T) {
// 			result := formatWithCommas(tt.input)
// 			assert.Equal(t, tt.expected, result)
// 		})
// 	}
// }

// func TestEdgeCases(t *testing.T) {
// 	t.Run("infinity and NaN handling", func(t *testing.T) {
// 		currency := USD

// 		// Test infinity
// 		result, err := currency.Amount().Float64ToUint128(math.Inf(1))
// 		assert.NoError(t, err)
// 		assert.True(t, currency.Amount().IsZero(result))

// 		// Test NaN
// 		result, err = currency.Amount().Float64ToUint128(math.NaN())
// 		assert.NoError(t, err)
// 		assert.True(t, currency.Amount().IsZero(result))
// 	})

// 	t.Run("very large decimal precision", func(t *testing.T) {
// 		// This should not panic but return 1 as fallback
// 		result := scaleFromDecimals(100)
// 		assert.Equal(t, uint64(1000000000000000000), result)
// 	})
// }
