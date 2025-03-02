package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStructToMap(t *testing.T) {
	type EmbeddedStruct struct {
		EmbeddedName string `json:"embedded_name"`
		EmbeddedVal  int    `json:"embedded_val"`
	}

	testCases := []struct {
		name          string
		input         any
		expected      map[string]any
		expectedError error
	}{
		{
			name: "Simple struct",
			input: struct {
				Name  string `json:"name"`
				Value int    `json:"value"`
			}{
				Name:  "test",
				Value: 123,
			},
			expected: map[string]any{
				"name":  "test",
				"value": 123,
			},
			expectedError: nil,
		},
		{
			name: "Struct with slice",
			input: struct {
				Items []string `json:"items"`
			}{
				Items: []string{"item1", "item2"},
			},
			expected: map[string]any{
				"items": []string{"item1", "item2"},
			},
			expectedError: nil,
		},
		{
			name: "Struct with empty slice",
			input: struct {
				Items []string `json:"items"`
			}{
				Items: []string{},
			},
			expected: map[string]any{
				"items": []string{},
			},
			expectedError: nil,
		},
		{
			name: "Struct with embedded struct",
			input: struct {
				EmbeddedStruct `json:",squash"`
				Other          string `json:"other"`
			}{
				EmbeddedStruct: EmbeddedStruct{
					EmbeddedName: "embedded",
					EmbeddedVal:  789,
				},
				Other: "other",
			},
			expected: map[string]any{
				"embedded_name": "embedded",
				"embedded_val":  789,
				"other":         "other",
			},
			expectedError: nil,
		},
		{
			name:          "Nil input",
			input:         nil,
			expected:      map[string]any{},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := StructToMap(tc.input)

			if tc.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}
