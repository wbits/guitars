package domain

import (
	"fmt"
	"strings"
)

// Currency is a closed enumeration of the currencies in which a guitar's price
// may be expressed. Keeping this small intentionally: the collection lives in
// the EUR/USD world only.
type Currency string

const (
	EUR Currency = "EUR"
	USD Currency = "USD"
)

func (c Currency) valid() bool {
	switch c {
	case EUR, USD:
		return true
	default:
		return false
	}
}

// Money is an immutable value object representing an amount in a specific
// currency. Amounts are modelled in minor units (cents) to avoid floating
// point rounding issues in a financial context.
type Money struct {
	amount   int64
	currency Currency
}

// NewMoney constructs a Money value object. The amount is given in minor units
// (e.g. cents). Negative amounts are rejected because a guitar in the
// collection cannot have a negative price.
func NewMoney(amount int64, currency Currency) (Money, error) {
	if amount < 0 {
		return Money{}, newValidationError("price.amount", "must be zero or positive")
	}
	c := Currency(strings.ToUpper(string(currency)))
	if !c.valid() {
		return Money{}, newValidationError("price.currency", fmt.Sprintf("unsupported currency %q", string(currency)))
	}
	return Money{amount: amount, currency: c}, nil
}

// Amount returns the price expressed in minor units.
func (m Money) Amount() int64 { return m.amount }

// Currency returns the ISO-like currency code (EUR or USD).
func (m Money) Currency() Currency { return m.currency }

// Equals reports whether two Money values are identical.
func (m Money) Equals(other Money) bool {
	return m.amount == other.amount && m.currency == other.currency
}

// String renders the value as e.g. "1234.56 EUR" for logging and debugging.
func (m Money) String() string {
	return fmt.Sprintf("%d.%02d %s", m.amount/100, abs(m.amount%100), m.currency)
}

func abs(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}
