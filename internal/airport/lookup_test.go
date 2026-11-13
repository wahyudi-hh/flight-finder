package airport

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookup(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
		ok   bool
	}{
		{name: "Jakarta", code: "CGK", want: "Jakarta", ok: true},
		{name: "Denpasar", code: "DPS", want: "Denpasar", ok: true},
		{name: "unknown airport", code: "XXX", ok: false},
		{name: "empty code", code: "", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Lookup(tt.code)

			assert.Equal(t, tt.ok, ok)
			if tt.ok {
				assert.Equal(t, tt.code, got.Code)
				assert.Equal(t, tt.want, got.City)
			} else {
				assert.Empty(t, got)
			}
		})
	}
}
