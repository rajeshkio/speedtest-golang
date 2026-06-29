package main

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		waitErr  bool
	}{
		{"30m", 30 * time.Minute, false},
		{"2h", 2 * time.Hour, false},
		{"1d", 1 * 24 * time.Hour, false},
		{"1w", 1 * 7 * 24 * time.Hour, false},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		got, err := parseDuration(tt.input)
		if tt.waitErr && err == nil {
			t.Errorf("input %s: expected error but got none", tt.input)
		}
		if !tt.waitErr && got != tt.expected {
			t.Errorf("input %s: expected %v got %v", tt.input, tt.expected, got)
		}
	}
}
