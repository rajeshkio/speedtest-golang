package main

import (
	"net/http"
	"net/http/httptest"
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

func TestIsSlow(t *testing.T) {
	slow, err := isSlow("50", 100)
	if err != nil {
		t.Errorf("did not expect an error, got %v", err)
	}
	if !slow {
		t.Errorf("expected 50 to be slow when the threshold is 100")
	}
}

func TestIsSlowBadInput(t *testing.T) {
	slow, err := isSlow("abc", 100)
	if err == nil {
		t.Errorf("expected an error for abc, got none")
	}
	if slow {
		t.Errorf("expected a bad value to not count as slow")
	}
}

func TestFetchAllResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"ID":1,"TimeStamp":"2026-11-24T01:00:00Z","DownloadSpeed":"109","UploadSpeed":"100","Latency":"5","PublicIp":"1.1.1.1","ISP":"x","Peers":"y"}]`))
	}))
	defer server.Close()

	results, err := fetchAllResults(server.URL)
	if err != nil {
		t.Fatalf("did not expect an error, got %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 row, got %d", len(results))
	}
}
