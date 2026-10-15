package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/paved-path/internal/billing"
)

// Handler-level tests for /quote: query parsing and error mapping, without a
// live server. The happy path asserts the exact integer-cents result; the
// error cases assert both status and message, because a 400 with an opaque
// body is not a clean API.
func TestQuote(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantBody   billing.Quote // checked when wantStatus == 200
		wantErr    string        // substring of the body otherwise
	}{
		{"valid", "/quote?cents=10000&bps=700", 200,
			billing.Quote{SubtotalCents: 10000, TaxCents: 700, TotalCents: 10700}, ""},
		{"zero tax", "/quote?cents=500&bps=0", 200,
			billing.Quote{SubtotalCents: 500, TaxCents: 0, TotalCents: 500}, ""},
		{"negative amount", "/quote?cents=-5&bps=700", 400, billing.Quote{},
			"amount must be non-negative"},
		{"tax rate too high", "/quote?cents=100&bps=10001", 400, billing.Quote{},
			"tax rate must be between 0 and 10000"},
		{"missing cents", "/quote?bps=700", 400, billing.Quote{},
			"cents and bps are required"},
		{"non-numeric", "/quote?cents=abc&bps=700", 400, billing.Quote{},
			"cents and bps are required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			routes().ServeHTTP(rec, httptest.NewRequest("GET", tt.query, nil))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus == 200 {
				var got billing.Quote
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatalf("bad JSON: %v", err)
				}
				if got != tt.wantBody {
					t.Fatalf("quote = %+v, want %+v", got, tt.wantBody)
				}
			} else if !strings.Contains(rec.Body.String(), tt.wantErr) {
				t.Fatalf("body %q does not contain %q", rec.Body.String(), tt.wantErr)
			}
		})
	}
}
