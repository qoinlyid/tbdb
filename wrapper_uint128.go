package tbdb

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// Uint128 represents a 128-bit number with two 64-bit parts.
type Uint128 struct {
	Hi uint64
	Lo uint64
}

// NewID generates a unique ID from the TigerBeetle binding.
func NewID() Uint128 {
	raw := types.ID()
	return fromBinding(raw)
}

// SafeNewID calls NewID but wrapped up with panic recover.
func SafeNewID() (id Uint128, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("new Uint128 id generation failed: %v", r)
		}
	}()
	id = NewID()
	return id, nil
}

// Uint128FromParts constructs Uint128 from hi and lo parts.
func Uint128FromParts(hi, lo uint64) Uint128 { return Uint128{Hi: hi, Lo: lo} }

// Uint128FromUint64 constructs Uint128 from a uint64 value.
func Uint128FromUint64(v uint64) Uint128 { return Uint128{Hi: 0, Lo: v} }

// Uint128FromBytesLE constructs Uint128 from 16-byte little-endian slice.
func Uint128FromBytesLE(b []byte) (Uint128, error) {
	if len(b) != 16 {
		return Uint128{}, fmt.Errorf("%w: want 16, got %d", ErrInvalidLength, len(b))
	}
	lo := binary.LittleEndian.Uint64(b[0:8])
	hi := binary.LittleEndian.Uint64(b[8:16])
	return Uint128{Hi: hi, Lo: lo}, nil
}

// Uint128FromHex decodes a hex string (big-endian, up to 32 chars) into Uint128.
func Uint128FromHex(s string) (Uint128, error) {
	if len(s) == 0 {
		return Uint128{}, nil
	}
	if len(s) > 32 {
		return Uint128{}, errors.Join(ErrHexTooLong, errors.New(">32 chars"))
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	var buf [16]byte
	n, err := hex.Decode(buf[16-len(s)/2:], []byte(s))
	if err != nil {
		return Uint128{}, err
	}
	_ = n
	// Buf is big-endian; convert to little-endian parts
	hi := binary.BigEndian.Uint64(buf[0:8])
	lo := binary.BigEndian.Uint64(buf[8:16])
	// Convert to little-endian logical value.
	return Uint128{
		Hi: hi,
		Lo: lo,
	}, nil
}

// Uint128FromBigInt converts a non-negative big.Int < 2^128 to Uint128.
func Uint128FromBigInt(b *big.Int) (Uint128, error) {
	if b == nil || b.Sign() < 0 {
		return Uint128{}, ErrNegativeOrNilBigInt
	}
	if b.BitLen() > 128 {
		return Uint128{}, errors.Join(ErrBigIntOverflow, errors.New("128 bits"))
	}
	var buf [16]byte
	b.FillBytes(buf[:]) // Big-endian
	return Uint128{
		Hi: binary.BigEndian.Uint64(buf[0:8]),
		Lo: binary.BigEndian.Uint64(buf[8:16]),
	}, nil
}

// BytesLE returns 16-byte little-endian representation.
func (u Uint128) BytesLE() [16]byte {
	var b [16]byte
	binary.LittleEndian.PutUint64(b[0:8], u.Lo)
	binary.LittleEndian.PutUint64(b[8:16], u.Hi)
	return b
}

// Hex returns big-endian hex string without leading zeros.
func (u Uint128) Hex() string {
	// Big-endian 16 bytes.
	var be [16]byte
	binary.BigEndian.PutUint64(be[0:8], u.Hi)
	binary.BigEndian.PutUint64(be[8:16], u.Lo)
	s := hex.EncodeToString(be[:])

	// Trim leading zeros but keep at least one digit.
	i := 0
	for i < len(s)-1 && s[i] == '0' {
		i++
	}
	return s[i:]
}

// AppendHex appends Uint128 hex string to dst.
func (u Uint128) AppendHex(dst []byte) []byte {
	hexStr := u.Hex()
	return append(dst, hexStr...)
}

// String implements fmt.Stringer by delegating to Hex.
func (u Uint128) String() string { return u.Hex() }

// BigInt returns a big.Int representation (heap allocation).
func (u Uint128) BigInt() *big.Int {
	var be [16]byte
	binary.BigEndian.PutUint64(be[0:8], u.Hi)
	binary.BigEndian.PutUint64(be[8:16], u.Lo)
	z := new(big.Int)
	return z.SetBytes(be[:])
}

// MarshalText implements encoding.TextMarshaler.
func (u Uint128) MarshalText() ([]byte, error) { return []byte(u.Hex()), nil }

// UnmarshalText implements encoding.TextUnmarshaler.
func (u *Uint128) UnmarshalText(b []byte) error {
	v, err := Uint128FromHex(string(b))
	if err != nil {
		return err
	}
	*u = v
	return nil
}

// MarshalJSON implements json.Marshaler as string to avoid precision issues.
func (u Uint128) MarshalJSON() ([]byte, error) { return json.Marshal(u.Hex()) }

// UnmarshalJSON implements json.Unmarshaler.
func (u *Uint128) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return u.UnmarshalText([]byte(s))
}

// IsZero returns true if Uint128 is zero.
func (u Uint128) IsZero() bool { return u.Hi == 0 && u.Lo == 0 }

// Compare compares two Uint128 values: -1 if u<v, 0 if u==v, 1 if u>v.
func (u Uint128) Compare(v Uint128) int {
	if u.Hi < v.Hi {
		return -1
	}
	if u.Hi > v.Hi {
		return 1
	}
	if u.Lo < v.Lo {
		return -1
	}
	if u.Lo > v.Lo {
		return 1
	}
	return 0
}

func toBinding(u Uint128) types.Uint128 {
	le := u.BytesLE()
	return types.BytesToUint128(le)
}

func fromBinding(b types.Uint128) Uint128 {
	le := b.Bytes()
	return Uint128{
		Lo: binary.LittleEndian.Uint64(le[0:8]),
		Hi: binary.LittleEndian.Uint64(le[8:16]),
	}
}
