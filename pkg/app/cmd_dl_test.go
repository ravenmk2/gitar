package app

import (
	"testing"
	"time"
)

func TestCalcMailRetryDelay(t *testing.T) {
	tests := []struct {
		num  int
		want time.Duration
	}{
		{1, time.Second},
		{2, 8 * time.Second},
		{3, 27 * time.Second},
		{13, 2197 * time.Second},
		{20, 7200 * time.Second},
		{100, 7200 * time.Second},
	}
	for _, tt := range tests {
		if got := calcMailRetryDelay(tt.num); got != tt.want {
			t.Errorf("calcMailRetryDelay(%d) = %s, want %s", tt.num, got, tt.want)
		}
	}
}
