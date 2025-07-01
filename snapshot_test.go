package main

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	var tests = []struct {
		d    time.Duration
		want string
	}{
		{time.Duration(0), "0s"},
		{time.Duration(12), "12ns"},
		{time.Duration(123), "123ns"},
		{time.Duration(1234), "1.23µs"},
		{time.Duration(12345), "12.3µs"},
		{time.Duration(123456), "123µs"},
		{time.Duration(12345678912), "12.35s"},
		{time.Duration(1*time.Minute + 1*time.Second), "1m1s"},
		{time.Duration(1*time.Minute + 1*time.Second + 1*time.Millisecond), "1m1s"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := formatDuration(tt.d); got != tt.want {
				t.Errorf("FormatDuration(%v) = %v, want %v", tt.d, got, tt.want)
			}
		})
	}

}
