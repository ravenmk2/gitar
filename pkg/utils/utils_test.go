package utils

import (
	"testing"
)

func TestHumanReadableSize(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1023, "1023 B"},
		{1024, "1.00 KiB"},
		{1536, "1.50 KiB"},
		{1048576, "1.00 MiB"},
		{1073741824, "1.00 GiB"},
	}
	for _, tt := range tests {
		if got := HumanReadableSize(tt.in); got != tt.want {
			t.Errorf("HumanReadableSize(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestStringSet(t *testing.T) {
	set := NewStringSet([]string{"a", "b"})
	if !set.Contains("a") || !set.Contains("b") {
		t.Error("expected set to contain a and b")
	}
	if set.Contains("c") {
		t.Error("expected set to not contain c")
	}

	empty := NewStringSet(nil)
	if empty.Contains("a") {
		t.Error("expected nil set to contain nothing")
	}

	zero := &StringSet{}
	if zero.Contains("a") {
		t.Error("expected zero set to contain nothing")
	}
}
