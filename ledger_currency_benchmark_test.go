package tbdb

// import (
// 	"math/rand"
// 	"testing"
// 	"time"
// )

// var (
// 	benchCurrency = USD

// 	// Pre-generate random values for realistic benchmarks
// 	randomAmounts []Uint128
// 	randomFloats  []float64
// )

// func init() {
// 	// Initialize random data for benchmarks using modern approach
// 	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

// 	randomAmounts = make([]Uint128, 1000)
// 	randomFloats = make([]float64, 1000)

// 	for i := range 1000 {
// 		randomAmounts[i] = Uint128FromUint64(rng.Uint64() % 1000000000)
// 		randomFloats[i] = rng.Float64() * 1000000
// 	}
// }

// // === Ledger Encoding/Decoding Benchmarks ===

// func BenchmarkEncodeLedger(b *testing.B) {
// 	currencies := []*currency{USD, IDR, BTC, ETH, VND, USDT}
// 	for i := 0; b.Loop(); i++ {
// 		curr := currencies[i%len(currencies)]
// 		_ = curr.EncodeLedger()
// 	}
// }

// func BenchmarkDecodeLedger(b *testing.B) {
// 	currencies := []*currency{USD, IDR, BTC, ETH, VND, USDT}
// 	for i := 0; b.Loop(); i++ {
// 		curr := currencies[i%len(currencies)]
// 		_ = curr.DecodeLedger()
// 	}
// }

// // === Conversion Benchmarks ===

// func BenchmarkFloat64ToUint128(b *testing.B) {
// 	for i := 0; b.Loop(); i++ {
// 		val := randomFloats[i%len(randomFloats)]
// 		_, _ = benchCurrency.Amount().Float64ToUint128(val)
// 	}
// }

// func BenchmarkUint128ToFloat64(b *testing.B) {
// 	for i := 0; b.Loop(); i++ {
// 		amount := randomAmounts[i%len(randomAmounts)]
// 		_ = benchCurrency.Amount().Uint128ToFloat64(amount)
// 	}
// }

// func BenchmarkUint128ToString(b *testing.B) {
// 	for i := 0; b.Loop(); i++ {
// 		amount := randomAmounts[i%len(randomAmounts)]
// 		_ = benchCurrency.Amount().Uint128ToString(amount)
// 	}
// }

// // === Arithmetic Operations Benchmarks ===

// func BenchmarkAdd(b *testing.B) {
// 	for i := 0; b.Loop(); i++ {
// 		a := randomAmounts[i%len(randomAmounts)]
// 		c := randomAmounts[(i+1)%len(randomAmounts)]
// 		_, _ = benchCurrency.Amount().Add(a, c)
// 	}
// }

// func BenchmarkSub(b *testing.B) {
// 	for i := 0; b.Loop(); i++ {
// 		a := randomAmounts[i%len(randomAmounts)]
// 		c := randomAmounts[(i+1)%len(randomAmounts)]
// 		_, _ = benchCurrency.Amount().Sub(a, c)
// 	}
// }

// func BenchmarkMul(b *testing.B) {
// 	multipliers := []float64{0.5, 1.5, 2.0, 0.7, 1.25}
// 	for i := 0; b.Loop(); i++ {
// 		amount := randomAmounts[i%len(randomAmounts)]
// 		multiplier := multipliers[i%len(multipliers)]
// 		_, _ = benchCurrency.Amount().Mul(amount, multiplier)
// 	}
// }

// func BenchmarkDiv(b *testing.B) {
// 	divisors := []float64{2.0, 3.5, 1.5, 4.0, 2.5}
// 	for i := 0; b.Loop(); i++ {
// 		amount := randomAmounts[i%len(randomAmounts)]
// 		divisor := divisors[i%len(divisors)]
// 		_, _ = benchCurrency.Amount().Div(amount, divisor)
// 	}
// }

// func BenchmarkPercentage(b *testing.B) {
// 	percentages := []float64{0.5, 1.0, 2.5, 0.7, 1.5}
// 	for i := 0; b.Loop(); i++ {
// 		amount := randomAmounts[i%len(randomAmounts)]
// 		percent := percentages[i%len(percentages)]
// 		_, _ = benchCurrency.Amount().Percentage(amount, percent)
// 	}
// }

// // === Comparison Operations Benchmarks ===

// func BenchmarkCompare(b *testing.B) {
// 	for i := 0; b.Loop(); i++ {
// 		a := randomAmounts[i%len(randomAmounts)]
// 		c := randomAmounts[(i+1)%len(randomAmounts)]
// 		_ = benchCurrency.Amount().Compare(a, c)
// 	}
// }

// func BenchmarkIsZero(b *testing.B) {
// 	for i := 0; b.Loop(); i++ {
// 		amount := randomAmounts[i%len(randomAmounts)]
// 		_ = benchCurrency.Amount().IsZero(amount)
// 	}
// }

// func BenchmarkEqual(b *testing.B) {
// 	for i := 0; b.Loop(); i++ {
// 		a := randomAmounts[i%len(randomAmounts)]
// 		c := randomAmounts[(i+1)%len(randomAmounts)]
// 		_ = benchCurrency.Amount().Equal(a, c)
// 	}
// }

// // === Utility Functions Benchmarks ===

// func BenchmarkScaleFromDecimals(b *testing.B) {
// 	decimals := []uint8{0, 2, 6, 8, 12, 18, 25} // Mix of cached and uncached values
// 	for i := 0; b.Loop(); i++ {
// 		decimal := decimals[i%len(decimals)]
// 		_ = scaleFromDecimals(decimal)
// 	}
// }

// func BenchmarkFormatWithCommas(b *testing.B) {
// 	values := []int64{123, 1234, 12345, 123456, 1234567, 12345678, 123456789}
// 	for i := 0; b.Loop(); i++ {
// 		val := values[i%len(values)]
// 		_ = formatWithCommas(val)
// 	}
// }

// func BenchmarkFormatCurrency(b *testing.B) {
// 	values := []float64{123.45, 1234.56, 12345.67, 123456.78}
// 	decimals := []uint8{0, 2, 6, 8}
// 	for i := 0; b.Loop(); i++ {
// 		val := values[i%len(values)]
// 		decimal := decimals[i%len(decimals)]
// 		_ = formatCurrency(val, decimal)
// 	}
// }

// // === Memory Allocation Benchmarks ===

// func BenchmarkUint128ToString_Allocs(b *testing.B) {
// 	amount := Uint128FromUint64(1234567890)

// 	b.ReportAllocs()
// 	for i := 0; b.Loop(); i++ {
// 		_ = benchCurrency.Amount().Uint128ToString(amount)
// 	}
// }

// func BenchmarkFormatWithCommas_Allocs(b *testing.B) {
// 	value := int64(1234567890)

// 	b.ReportAllocs()
// 	for i := 0; b.Loop(); i++ {
// 		_ = formatWithCommas(value)
// 	}
// }

// // === Comparative Benchmarks ===

// func BenchmarkCompare_FastPath_vs_SlowPath(b *testing.B) {
// 	// Small values (fast path)
// 	smallA := Uint128FromUint64(12345)
// 	smallB := Uint128FromUint64(67890)

// 	// Large values (slow path)
// 	largeA := randomAmounts[0]
// 	largeB := randomAmounts[1]

// 	b.Run("FastPath", func(b *testing.B) {
// 		for i := 0; b.Loop(); i++ {
// 			_ = benchCurrency.Amount().Compare(smallA, smallB)
// 		}
// 	})

// 	b.Run("SlowPath", func(b *testing.B) {
// 		for i := 0; b.Loop(); i++ {
// 			_ = benchCurrency.Amount().Compare(largeA, largeB)
// 		}
// 	})
// }

// // === Stress Test Benchmarks ===

// func BenchmarkMixedOperations(b *testing.B) {
// 	operations := []func(int){
// 		func(i int) {
// 			a := randomAmounts[i%len(randomAmounts)]
// 			c := randomAmounts[(i+1)%len(randomAmounts)]
// 			_, _ = benchCurrency.Amount().Add(a, c)
// 		},
// 		func(i int) {
// 			amount := randomAmounts[i%len(randomAmounts)]
// 			_ = benchCurrency.Amount().Uint128ToString(amount)
// 		},
// 		func(i int) {
// 			val := randomFloats[i%len(randomFloats)]
// 			_, _ = benchCurrency.Amount().Float64ToUint128(val)
// 		},
// 		func(i int) {
// 			a := randomAmounts[i%len(randomAmounts)]
// 			c := randomAmounts[(i+1)%len(randomAmounts)]
// 			_ = benchCurrency.Amount().Compare(a, c)
// 		},
// 	}

// 	for i := 0; b.Loop(); i++ {
// 		op := operations[i%len(operations)]
// 		op(i)
// 	}
// }
