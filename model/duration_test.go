package model

import "testing"

func TestParseDuration(t *testing.T) {
	type args struct {
		duration string
	}
	tests := []struct {
		name      string
		args      args
		wantHours float64
		wantErr   bool
	}{
		{name: "2hours", args: args{duration: "2h"}, wantHours: 2.0, wantErr: false},
		{name: "1day", args: args{duration: "1.0d"}, wantHours: 24.0, wantErr: false},
		{name: "2weeks", args: args{duration: "2w"}, wantHours: 336.0, wantErr: false},
		{name: "1month", args: args{duration: "1m"}, wantHours: 5040.0, wantErr: false},
		{name: "invalidDuration", args: args{duration: "2e"}, wantHours: 0.0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHours, err := ParseDuration(tt.args.duration)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotHours != tt.wantHours {
				t.Errorf("ParseDuration() = %v, want %v", gotHours, tt.wantHours)
			}
		})
	}
}
