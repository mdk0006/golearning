package healthcheck

import (
	"testing"
)

func TestCheckCPU(t *testing.T) {
	tests := []struct {
		name     string
		usage    float64
		expected string
	}{
		{"healthy", 45.0, "OK"},
		{"boundary", 90.0, "OK"},
		{"critical", 95.0, "CRITICAL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckCPU(tt.usage)
			if got != tt.expected {
				t.Errorf("got %s, want %s", got, tt.expected)
			}
		})
	}
}
func TestCheckDisk(t *testing.T) {
	tests := []struct {
		name     string
		freeGB   float64
		expected string
	}{
		{"healthy", 20.0, "OK"},
		{"boundary", 10.0, "OK"},
		{"warning", 5.0, "WARNING"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckDisk(tt.freeGB)
			if got != tt.expected {
				t.Errorf("got %s wants %s", got, tt.expected)
			}

		})
	}
}
