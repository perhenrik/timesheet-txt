package report

import (
	"testing"
	"time"
)

func Test_dateInRange(t *testing.T) {
	type args struct {
		current time.Time
		start   time.Time
		end     time.Time
	}
	tests := []struct {
		name       string
		args       args
		wantResult bool
	}{
		{name: "InRange", args: args{
			current: time.Date(2018, 5, 3, 0, 0, 0, 0, time.UTC),
			start:   time.Date(2018, 5, 2, 0, 0, 0, 0, time.UTC),
			end:     time.Date(2018, 5, 4, 0, 0, 0, 0, time.UTC),
		}, wantResult: true},
		{name: "InRangeSameDate", args: args{
			current: time.Date(2018, 5, 3, 0, 0, 0, 0, time.UTC),
			start:   time.Date(2018, 5, 3, 0, 0, 0, 0, time.UTC),
			end:     time.Date(2018, 5, 3, 0, 0, 0, 0, time.UTC),
		}, wantResult: true},
		{name: "NotInRangeDateBefore", args: args{
			current: time.Date(2018, 5, 2, 0, 0, 0, 0, time.UTC),
			start:   time.Date(2018, 5, 3, 0, 0, 0, 0, time.UTC),
			end:     time.Date(2018, 5, 3, 0, 0, 0, 0, time.UTC),
		}, wantResult: false},
		{name: "NotInRangeDateAfter", args: args{
			current: time.Date(2018, 5, 4, 0, 0, 0, 0, time.UTC),
			start:   time.Date(2018, 5, 3, 0, 0, 0, 0, time.UTC),
			end:     time.Date(2018, 5, 3, 0, 0, 0, 0, time.UTC),
		}, wantResult: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := dateInRange(tt.args.current, tt.args.start, tt.args.end); gotResult != tt.wantResult {
				t.Errorf("inRange() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
