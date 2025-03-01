package misc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove ym:*: prefix",
			input:    "ym:some:key",
			expected: "key",
		},
		{
			name:     "convert camelCase to snake_case",
			input:    "someKeyName",
			expected: "some_key_name",
		},
		{
			name:     "exclusion for watchIDs",
			input:    "watchIDs",
			expected: "watch_ids",
		},
		{
			name:     "exclusion for iFrame",
			input:    "iFrame",
			expected: "iframe",
		},
		{
			name:     "no transformation needed",
			input:    "already_snake_case",
			expected: "already_snake_case",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ProcessKey(tt.input))
		})
	}
}
