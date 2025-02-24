package misc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixArrayDateTime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected any
	}{
		{
			name:     "empty string",
			input:    ``,
			expected: ``,
		},
		{
			name:     "empty array",
			input:    `[]`,
			expected: `[]`,
		},
		{
			name:     "one-length array",
			input:    `[\\'2024-12-31 19:53:19\\']`,
			expected: []string{"2024-12-31 19:53:19"},
		},
		{
			name:     "two-length array with space",
			input:    `[\\'2024-12-31 19:53:19\\', \\'2024-12-31 19:53:19\\']`,
			expected: []string{"2024-12-31 19:53:19", "2024-12-31 19:53:19"},
		},
		{
			name:     "two-length array without space",
			input:    `[\\'2024-12-31 19:53:19\\',\\'2024-12-31 19:53:19\\']`,
			expected: []string{"2024-12-31 19:53:19", "2024-12-31 19:53:19"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, FixArrayDateTime(tt.input))
		})
	}
}

func TestFixWatchIDs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected any
	}{
		{
			name:     "empty string",
			input:    ``,
			expected: ``,
		},
		{
			name:     "empty array",
			input:    `[]`,
			expected: `[]`,
		},
		{
			name:     "one-length array",
			input:    `[18023332624550854749]`,
			expected: []json.Number{"18023332624550854749"},
		},
		{
			name:     "two-length array with space",
			input:    `[18023332624550854749, 18023347689297543250, -424224913253989848]`,
			expected: []json.Number{"18023332624550854749", "18023347689297543250", "18022519160455561768"},
		},
		{
			name:     "two-length array without space",
			input:    `[18023332624550854749, 18023347689297543250,-424224913253989848]`,
			expected: []json.Number{"18023332624550854749", "18023347689297543250", "18022519160455561768"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, FixWatchIDs(tt.input))
		})
	}
}
