package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDuration(t *testing.T) {
	tests := []struct {
		name         string
		totalMinutes int
		formatted    string
	}{
		{name: "zero", totalMinutes: 0, formatted: "0h 0m"},
		{name: "minutes only", totalMinutes: 45, formatted: "0h 45m"},
		{name: "exact hours", totalMinutes: 120, formatted: "2h 0m"},
		{name: "hours and minutes", totalMinutes: 155, formatted: "2h 35m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDuration(tt.totalMinutes)

			assert.Equal(t, tt.totalMinutes, got.TotalMinutes)
			assert.Equal(t, tt.formatted, got.Formatted)
		})
	}
}
