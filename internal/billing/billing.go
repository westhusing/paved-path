// Package billing computes invoice quotes.
// Money is integer cents and tax rates are basis points (1 bp = 0.01%),
// so the money path is exact integer arithmetic — no float64 in sight.
package billing

import "errors"

var (
	ErrNegativeAmount = errors.New("amount must be non-negative")
	ErrBadTaxRate     = errors.New("tax rate must be between 0 and 10000 basis points")
)

type Quote struct {
	SubtotalCents int64 `json:"subtotal_cents"`
	TaxCents      int64 `json:"tax_cents"`
	TotalCents    int64 `json:"total_cents"`
}

// Compute returns the quote for amountCents at taxRateBps.
// Tax is rounded half-up to the nearest cent.
func Compute(amountCents int64, taxRateBps int) (Quote, error) {
	if amountCents < 0 {
		return Quote{}, ErrNegativeAmount
	}
	if taxRateBps < 0 || taxRateBps > 10000 {
		return Quote{}, ErrBadTaxRate
	}
	tax := (amountCents*int64(taxRateBps) + 5000) / 10000 // half-up
	return Quote{SubtotalCents: amountCents, TaxCents: tax, TotalCents: amountCents + tax}, nil
}
