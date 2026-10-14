package billing

import "testing"

func TestCompute(t *testing.T) {
	tests := []struct {
		name    string
		cents   int64
		bps     int
		want    Quote
		wantErr error
	}{
		{"no tax", 10000, 0, Quote{SubtotalCents: 10000, TaxCents: 0, TotalCents: 10000}, nil},
		{"7 percent", 10000, 700, Quote{SubtotalCents: 10000, TaxCents: 700, TotalCents: 10700}, nil},
		{"zero amount", 0, 700, Quote{SubtotalCents: 0, TaxCents: 0, TotalCents: 0}, nil},
		{"half cent rounds up", 10, 500, Quote{SubtotalCents: 10, TaxCents: 1, TotalCents: 11}, nil},
		{"negative amount", -1, 700, Quote{}, ErrNegativeAmount},
		{"rate too high", 10000, 10001, Quote{}, ErrBadTaxRate},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Compute(tt.cents, tt.bps)
			if err != tt.wantErr {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
