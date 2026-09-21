package web

import (
	"testing"
)

func TestQValues(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected bool
	}{
		{
			name:     "single_no_weight",
			header:   "gzip",
			expected: true,
		},
		{
			name:     "single_0_weight",
			header:   "gzip;q=0.0",
			expected: false,
		},
		{
			name:     "missing",
			header:   "deflate",
			expected: false,
		},
		{
			name:     "multiple_options_0_weight",
			header:   "deflate, gzip;q=0",
			expected: false,
		},
		{
			name:     "single_non_zero_weight",
			header:   "gzip;q=0.5",
			expected: true,
		},
		{
			name:     "single_non_zero_weight_whitespace",
			header:   "gzip;         q=0.5",
			expected: true,
		},
		{
			name:     "multiple_non_zero_weight",
			header:   "deflate, gzip;q=1",
			expected: true,
		},
		{
			name:     "case_insensitive",
			header:   "deflate, GZIP;q=1",
			expected: true,
		},
		{
			name:     "wildcard",
			header:   "deflate, *;q=0.5",
			expected: true,
		},
		{
			name:     "rejected_and_wildcard",
			header:   "deflate, gzip;q=0, *;q=0.5",
			expected: false,
		},
		{
			name:     "rejected_and_wildcard_reverse_order",
			header:   "deflate, *;q=0.5, gzip;q=0",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			encodings, err := parseQValues(tc.header)

			if err != nil {
				tt.Fatal(err)
			}

			ok := encodings.accepts("gzip")
			if ok != tc.expected {
				tt.Fatalf("encodings.accepts(\"gzip\") = %v, want %v", ok, tc.expected)
			}
		})
	}

}

func TestQValuesErr(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "weight_negative",
			value: "gzip;q=-0.5",
		},
		{
			name:  "weight_out_of_range",
			value: "gzip;q=5",
		},
		{
			name:  "invalid_weight",
			value: "gzip;q=foo",
		},
		{
			name:  "nan_weight",
			value: "gzip;q=NaN",
		},
		{
			name:  "not_q",
			value: "gzip;x=0.5",
		},
		{
			name:  "missing_weight",
			value: "gzip;q",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			_, err := parseQValues(tc.value)

			if err == nil {
				tt.Fatal("expected error")
			}
		})
	}

}
