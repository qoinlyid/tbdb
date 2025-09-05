package tbdb

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewID(t *testing.T) {
	id := NewID()
	assert.NotZero(t, id, "NewID() should return a non-zero Uint128")
	assert.LessOrEqual(t, id.Lo, uint64(^uint64(0)), "Low part of NewID() exceeds uint64 max")
	assert.LessOrEqual(t, id.Hi, uint64(^uint64(0)), "High part of NewID() exceeds uint64 max")
}

func TestUint128FromParts(t *testing.T) {
	tests := []struct {
		name string
		hi   uint64
		lo   uint64
		want Uint128
	}{
		{"zero", 0, 0, Uint128{Hi: 0, Lo: 0}},
		{"only lo", 0, 12345, Uint128{Hi: 0, Lo: 12345}},
		{"only hi", 6789, 0, Uint128{Hi: 6789, Lo: 0}},
		{"both", 1, 2, Uint128{Hi: 1, Lo: 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Uint128FromParts(tt.hi, tt.lo)
			assert.Equal(t, tt.want, got, "Uint128FromParts(%d, %d) returned unexpected result", tt.hi, tt.lo)
		})
	}
}

func TestUint128FromUint64(t *testing.T) {
	tests := []struct {
		name string
		val  uint64
		want Uint128
	}{
		{"zero", 0, Uint128{Hi: 0, Lo: 0}},
		{"small", 42, Uint128{Hi: 0, Lo: 42}},
		{"max uint64", ^uint64(0), Uint128{Hi: 0, Lo: ^uint64(0)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Uint128FromUint64(tt.val)
			assert.Equal(t, tt.want, got, "Uint128FromUint64(%d) returned unexpected result", tt.val)
		})
	}
}

func TestUint128FromBytesLE(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    Uint128
		wantErr bool
	}{
		{"valid bytes", []byte{
			0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		}, Uint128{Hi: 2, Lo: 1}, false},
		{"invalid length", []byte{0x01, 0x02}, Uint128{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Uint128FromBytesLE(tt.input)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for input %v", tt.input)
			} else {
				assert.NoError(t, err, "Unexpected error for input %v", tt.input)
				assert.Equal(t, tt.want, got, "Uint128FromBytesLE(%v) returned unexpected result", tt.input)
			}
		})
	}
}

func TestUint128FromHex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Uint128
		wantErr bool
	}{
		{"empty", "", Uint128{}, false},
		{"simple hex", "0102", Uint128{Hi: 0, Lo: 0x0102}, false},
		{"leading zeros", "00000001", Uint128{Hi: 0, Lo: 1}, false},
		{"max 128-bit", "ffffffffffffffffffffffffffffffff", Uint128{Hi: 0xffffffffffffffff, Lo: 0xffffffffffffffff}, false},
		{"too long", "1ffffffffffffffffffffffffffffffff", Uint128{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Uint128FromHex(tt.input)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for input hex: %s", tt.input)
			} else {
				assert.NoError(t, err, "Unexpected error for input hex: %s", tt.input)
				assert.Equal(t, tt.want, got, "Uint128FromHex(%s) returned unexpected result", tt.input)
			}
		})
	}
}

func TestUint128FromBigInt(t *testing.T) {
	tests := []struct {
		name    string
		input   *big.Int
		want    Uint128
		wantErr bool
	}{
		{"nil", nil, Uint128{}, true},
		{"negative", big.NewInt(-1), Uint128{}, true},
		{"zero", big.NewInt(0), Uint128{Hi: 0, Lo: 0}, false},
		{"small", big.NewInt(42), Uint128{Hi: 0, Lo: 42}, false},
		{"max 128-bit", func() *big.Int {
			b := new(big.Int)
			b.SetString("340282366920938463463374607431768211455", 10) // 2^128-1
			return b
		}(), Uint128{Hi: 0xffffffffffffffff, Lo: 0xffffffffffffffff}, false},
		{"overflow", func() *big.Int {
			b := new(big.Int)
			b.SetString("340282366920938463463374607431768211456", 10) // 2^128
			return b
		}(), Uint128{}, true},
		{"tb real", func() *big.Int {
			b := new(big.Int)
			b.SetString("2122268224272085121281158996020055632", 10) // 2^128
			return b
		}(), Uint128{Hi: 0x198bbe6eb9b35ae, Lo: 0x66392201c924e50}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Uint128FromBigInt(tt.input)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for input BigInt: %v", tt.input)
			} else {
				assert.NoError(t, err, "Unexpected error for input BigInt: %v", tt.input)
				assert.Equal(t, tt.want, got, "Uint128FromBigInt(%v) returned unexpected result", tt.input)
			}
		})
	}
}
