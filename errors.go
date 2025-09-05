package tbdb

import (
	"errors"
	"fmt"
)

var (
	ErrOpenTBConnection    = errors.New("failed to open TigerBeetle connection")
	ErrClientNil           = errors.New("client TigeBeetle is nil")
	ErrInvalidLength       = errors.New("invalid length")
	ErrHexTooLong          = errors.New("hex too long")
	ErrNegativeOrNilBigInt = errors.New("negative or nil big.Int")
	ErrBigIntOverflow      = errors.New("big.Int overflows")

	// Operations.
	ErrMonetaryMustNotBeNil       = errors.New("account monetary must not be nil")
	ErrUnknownCategory            = errors.New("unknown account category")
	ErrCurrencyAmountMismatch     = errors.New("currency mismatch")
	ErrExceedsMaxTigerBeetleBatch = fmt.Errorf("exceeds maximum batch size %d", TigerBeetleMaxBatch)
	ErrAccountIDMustNotBeZero     = errors.New("account id must not be zero")
	ErrAccountIDMustNotBeIntMax   = errors.New("account id must not be 2^128 - 1")
	ErrTimeMinMustNotBeZero       = errors.New("account transfer filter time min must not be zero")
	ErrTimeMaxMustNotBeZero       = errors.New("account transfer filter time max must not be zero")
)
