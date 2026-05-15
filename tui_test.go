package main

import "testing"

func TestFormatDateDigits(t *testing.T) {
	tests := []struct {
		name   string
		digits string
		want   string
	}{
		{name: "empty", digits: "", want: ""},
		{name: "year", digits: "2026", want: "2026-"},
		{name: "yearMonth", digits: "202605", want: "2026-05-"},
		{name: "fullDate", digits: "20260513", want: "2026-05-13"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDateDigits(tt.digits); got != tt.want {
				t.Fatalf("formatDateDigits() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeDateDigits(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{name: "monthSingleDigitGetsPadded", input: "20265", output: "202605"},
		{name: "daySingleDigitNotAutoPadded", input: "2026053", output: "2026053"},
		{name: "validFullDateUnchanged", input: "20260513", output: "20260513"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeDateDigits(tt.input); got != tt.output {
				t.Fatalf("normalizeDateDigits() = %q, want %q", got, tt.output)
			}
		})
	}
}

func TestIsPotentialDateDigits(t *testing.T) {
	tests := []struct {
		name   string
		digits string
		want   bool
	}{
		{name: "validPrefix", digits: "20260", want: true},
		{name: "invalidMonthTens", digits: "20262", want: false},
		{name: "invalidMonth", digits: "202613", want: false},
		{name: "invalidDay", digits: "20260432", want: false},
		{name: "validLeapDay", digits: "20240229", want: true},
		{name: "invalidLeapDay", digits: "20230229", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPotentialDateDigits(tt.digits); got != tt.want {
				t.Fatalf("isPotentialDateDigits() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateDateInput(t *testing.T) {
	if err := validateDateInput("2026-05-13"); err != nil {
		t.Fatalf("validateDateInput() error = %v, want nil", err)
	}

	if err := validateDateInput("2026-02-30"); err == nil {
		t.Fatalf("validateDateInput() expected error for invalid date")
	}

	if err := validateDateInput("2026-0"); err == nil {
		t.Fatalf("validateDateInput() expected error for incomplete date")
	}
}

func TestValidateDateInputAcceptsHyphenSeparated(t *testing.T) {
	if err := validateDateInput("2026-5-3"); err != nil {
		t.Fatalf("validateDateInput() error = %v, want nil", err)
	}
}
