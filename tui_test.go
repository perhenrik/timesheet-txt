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

func TestPeriodValidationHint(t *testing.T) {
	text, isError := periodValidationHint("5d")
	if isError || text != "looks good" {
		t.Fatalf("periodValidationHint(5d) = (%q, %v), want (looks good, false)", text, isError)
	}

	text, isError = periodValidationHint("5x")
	if !isError {
		t.Fatalf("periodValidationHint(5x) should be error")
	}
}

func TestDateValidationHint(t *testing.T) {
	text, isError := dateValidationHint("2026-05-13")
	if isError || text != "looks good" {
		t.Fatalf("dateValidationHint(valid) = (%q, %v), want (looks good, false)", text, isError)
	}

	text, isError = dateValidationHint("2026-5")
	if isError || text != "continue typing" {
		t.Fatalf("dateValidationHint(partial) = (%q, %v), want (continue typing, false)", text, isError)
	}

	text, isError = dateValidationHint("2026-99-99")
	if !isError {
		t.Fatalf("dateValidationHint(invalid) should be error")
	}
}

func TestCanCycleFocusedOptions(t *testing.T) {
	m := newTUIModel()

	cycleIndexes := []int{2, 3, 5, 6}
	for _, idx := range cycleIndexes {
		m.focusIndex = idx
		if !m.canCycleFocusedOptions() {
			t.Fatalf("focusIndex %d should allow cycling", idx)
		}
	}

	nonCycleIndexes := []int{0, 1, 4, 7}
	for _, idx := range nonCycleIndexes {
		m.focusIndex = idx
		if m.canCycleFocusedOptions() {
			t.Fatalf("focusIndex %d should not allow cycling", idx)
		}
	}
}

func TestBuildManualTaskUsesEnteredDate(t *testing.T) {
	task, normalizedDate, hours, label, err := buildManualTask("2026-05-01", "project", "db", "2.5")
	if err != nil {
		t.Fatalf("buildManualTask() error = %v", err)
	}

	if normalizedDate != "2026-05-01" {
		t.Fatalf("normalizedDate = %q, want 2026-05-01", normalizedDate)
	}

	if !task.Completed {
		t.Fatalf("task should be completed")
	}

	if got := task.CreatedDate.Format("2006-01-02"); got != "2026-05-01" {
		t.Fatalf("CreatedDate = %q, want 2026-05-01", got)
	}

	if got := task.CompletedDate.Format("2006-01-02"); got != "2026-05-01" {
		t.Fatalf("CompletedDate = %q, want 2026-05-01", got)
	}

	if task.CompletedDate.Hour() != 0 || task.CompletedDate.Minute() != 0 || task.CompletedDate.Second() != 0 {
		t.Fatalf("CompletedDate should be midnight UTC, got %s", task.CompletedDate)
	}

	if hours != 2.5 {
		t.Fatalf("hours = %v, want 2.5", hours)
	}

	if label != "project.db" {
		t.Fatalf("label = %q, want project.db", label)
	}
}

func TestBuildManualTaskRejectsInvalidDate(t *testing.T) {
	_, _, _, _, err := buildManualTask("2026-13-01", "project", "db", "2.5")
	if err == nil {
		t.Fatalf("expected error for invalid date")
	}
}

func TestBuildManualTaskRejectsInvalidHours(t *testing.T) {
	_, _, _, _, err := buildManualTask("2026-05-01", "project", "db", "0")
	if err == nil {
		t.Fatalf("expected error for invalid hours")
	}
}

func TestBuildManualTaskAcceptsFlexibleDate(t *testing.T) {
	task, normalizedDate, _, _, err := buildManualTask("2026-5-3", "project", "db", "1")
	if err != nil {
		t.Fatalf("buildManualTask() error = %v", err)
	}

	if normalizedDate != "2026-05-03" {
		t.Fatalf("normalizedDate = %q, want 2026-05-03", normalizedDate)
	}

	if got := task.CompletedDate.Format("2006-01-02"); got != "2026-05-03" {
		t.Fatalf("CompletedDate = %q, want 2026-05-03", got)
	}
}

func TestBuildManualTaskOriginalContainsManualDate(t *testing.T) {
	task, _, _, _, err := buildManualTask("2026-05-01", "project", "db", "1")
	if err != nil {
		t.Fatalf("buildManualTask() error = %v", err)
	}

	if task.Original == "" {
		t.Fatalf("task.Original should be set")
	}

	if task.CompletedDate.Format("2006-01-02") != "2026-05-01" {
		t.Fatalf("task should keep manual completed date")
	}

	if task.CreatedDate.Format("2006-01-02") != "2026-05-01" {
		t.Fatalf("task should keep manual created date")
	}

	if task.CompletedDate.IsZero() {
		t.Fatalf("task should have completed date")
	}

	if task.CreatedDate.IsZero() {
		t.Fatalf("task should have created date")
	}

	if got := task.Original; got == "" {
		t.Fatalf("task.Original should not be empty")
	}
}
