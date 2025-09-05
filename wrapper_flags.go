package tbdb

import (
	"strings"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type AccountFlags struct {
	Linked                     bool
	DebitsMustNotExceedCredits bool
	CreditsMustNotExceedDebits bool
	History                    bool
	Imported                   bool
	Closed                     bool
}

func (f AccountFlags) ToUint16() uint16 {
	tbAccountFlag := types.AccountFlags{
		Linked:                     f.Linked,
		DebitsMustNotExceedCredits: f.DebitsMustNotExceedCredits,
		CreditsMustNotExceedDebits: f.CreditsMustNotExceedDebits,
		History:                    f.History,
		Imported:                   f.Imported,
		Closed:                     f.Closed,
	}
	return tbAccountFlag.ToUint16()
}

func (f AccountFlags) String() string {
	var flags []string
	if f.Linked {
		flags = append(flags, "linked")
	}
	if f.DebitsMustNotExceedCredits {
		flags = append(flags, "debits_must_not_exceed_credits")
	}
	if f.CreditsMustNotExceedDebits {
		flags = append(flags, "credits_must_not_exceed_debits")
	}
	if f.History {
		flags = append(flags, "history")
	}
	if f.Imported {
		flags = append(flags, "imported")
	}
	if f.Closed {
		flags = append(flags, "closed")
	}
	return strings.Join(flags, ",")
}

type TransferFlags struct {
	Linked              bool
	Pending             bool
	PostPendingTransfer bool
	VoidPendingTransfer bool
	BalancingDebit      bool
	BalancingCredit     bool
	ClosingDebit        bool
	ClosingCredit       bool
	Imported            bool
}

func (f TransferFlags) ToUint16() uint16 {
	tbTransferFlag := types.TransferFlags{
		Linked:              f.Linked,
		Pending:             f.Pending,
		PostPendingTransfer: f.PostPendingTransfer,
		VoidPendingTransfer: f.VoidPendingTransfer,
		BalancingDebit:      f.BalancingDebit,
		BalancingCredit:     f.BalancingCredit,
		ClosingDebit:        f.ClosingDebit,
		ClosingCredit:       f.ClosingCredit,
		Imported:            f.Imported,
	}
	return tbTransferFlag.ToUint16()
}

func (f TransferFlags) String() string {
	var flags []string
	if f.Linked {
		flags = append(flags, "linked")
	}
	if f.Pending {
		flags = append(flags, "pending")
	}
	if f.PostPendingTransfer {
		flags = append(flags, "post_pending_transfer")
	}
	if f.VoidPendingTransfer {
		flags = append(flags, "void_pending_transfer")
	}
	if f.BalancingDebit {
		flags = append(flags, "balancing_debit")
	}
	if f.BalancingCredit {
		flags = append(flags, "balancing_credit")
	}
	if f.ClosingDebit {
		flags = append(flags, "closing_debit")
	}
	if f.ClosingCredit {
		flags = append(flags, "closing_credit")
	}
	if f.Imported {
		flags = append(flags, "imported")
	}
	return strings.Join(flags, ",")
}

type AccountFilterFlags struct {
	Debits   bool
	Credits  bool
	Reversed bool
}

func (f AccountFilterFlags) ToUint32() uint32 {
	tbAccountFilterFlag := types.AccountFilterFlags{
		Debits:   f.Debits,
		Credits:  f.Credits,
		Reversed: f.Reversed,
	}
	return tbAccountFilterFlag.ToUint32()
}
