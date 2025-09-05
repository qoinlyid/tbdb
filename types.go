package tbdb

type LedgerCode uint32

// Ledger defines the interface for encoding and decoding ledger codes.
// Used to generate unique identifiers for TigerBeetle ledgers based on currency properties.
type Ledger interface {
	// EncodeLedger generates a unique LedgerCode for TigerBeetle.
	// The code combines currency properties into a deterministic identifier.
	EncodeLedger() LedgerCode

	// DecodeLedger converts the LedgerCode back to human-readable format.
	// Returns a formatted string describing the currency and its properties.
	DecodeLedger() string
}

// Amount defines the interface for monetary amount operations and conversions.
// Implementers can define custom scaling, precision, and formatting rules
// while maintaining compatibility with TigerBeetle's Uint128 amount type.
// This allows for different currency systems, token standards, or custom monetary units.
type Amount interface {
	// SetUint128Value sets Uint128 value to the monetary.
	SetUint128Value(val Uint128) Amount

	// SetFloat64Value sets float64 value to the monetary.
	SetFloat64Value(val float64) Amount

	// Float64ToUint128 converts a human-readable float64 to TigerBeetle's Uint128 format.
	// Implementation should handle scaling and precision according to the specific
	// monetary unit's requirements (e.g., decimal places, rounding rules).
	Float64ToUint128() (Uint128, error)

	// ToUint128FromValue returns Uint128 valu type depends on value that has been set.
	ToUint128FromValue() Uint128

	// Uint128ToFloat64 converts TigerBeetle's Uint128 back to human-readable float64.
	// Implementation should reverse the scaling applied in Float64ToUint128.
	Uint128ToFloat64() float64

	// Uint128ToString formats a Uint128 amount as a human-readable string.
	// Implementation can define custom formatting rules (currency symbols,
	// thousand separators, decimal precision, etc.).
	Uint128ToString() string

	// Add safely adds two Uint128 amounts.
	// Returns error if the operation would cause overflow or violates
	// implementation-specific constraints.
	Add(b Uint128) (Uint128, error)

	// Sub safely subtracts b from a.
	// Returns error if the result would be invalid according to
	// implementation rules (e.g., negative amounts not allowed).
	Sub(b Uint128) (Uint128, error)

	// Mul multiplies an amount by a floating-point multiplier.
	// Implementation should handle precision and rounding according to
	// the monetary unit's requirements.
	Mul(multiplier float64) (Uint128, error)

	// Div divides an amount by a floating-point divisor.
	// Implementation should handle division precision and edge cases
	// (zero divisor, infinity, NaN) according to its requirements.
	Div(divisor float64) (Uint128, error)

	// Percentage calculates a percentage of the given amount.
	// Implementation should define how percentage calculation and rounding
	// are handled for the specific monetary unit.
	Percentage(percent float64) (Uint128, error)

	// Compare compares two Uint128 amounts according to implementation rules.
	// Returns:
	//	- -1 if a < b
	//	- 0 if a == b
	//	- 1 if a > b
	Compare(b Uint128) int

	// IsZero checks if the amount represents zero value according to
	// implementation-defined zero semantics.
	IsZero() bool

	// Equal checks if two amounts are considered equal by the implementation.
	Equal(b Uint128) bool

	// GreaterThan checks if a > b according to implementation comparison rules.
	GreaterThan(b Uint128) bool

	// GreaterThanOrEqual checks if a >= b according to implementation comparison rules.
	GreaterThanOrEqual(b Uint128) bool

	// LessThan checks if a < b according to implementation comparison rules.
	LessThan(b Uint128) bool

	// LessThanOrEqual checks if a <= b according to implementation comparison rules.
	LessThanOrEqual(b Uint128) bool

	// Min returns the smaller of two amounts according to implementation comparison rules.
	Min(b Uint128) Uint128

	// Max returns the larger of two amounts according to implementation comparison rules.
	Max(b Uint128) Uint128
}
