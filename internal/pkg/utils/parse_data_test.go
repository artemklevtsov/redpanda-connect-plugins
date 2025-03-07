package utils

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDate(t *testing.T) {
	now := time.Now()
	dateLayout := "2006-01-02"

	tests := []struct {
		name        string
		input       string
		expected    string
		expectedErr error
	}{
		{
			name:        "today",
			input:       "today",
			expected:    now.Format(dateLayout),
			expectedErr: nil,
		},
		{
			name:        "TODAY",
			input:       "TODAY",
			expected:    now.Format(dateLayout),
			expectedErr: nil,
		},
		{
			name:        "yesterday",
			input:       "yesterday",
			expected:    now.AddDate(0, 0, -1).Format(dateLayout),
			expectedErr: nil,
		},
		{
			name:        "YESTERDAY",
			input:       "YESTERDAY",
			expected:    now.AddDate(0, 0, -1).Format(dateLayout),
			expectedErr: nil,
		},
		{
			name:        "7daysago",
			input:       "7daysago",
			expected:    now.AddDate(0, 0, -7).Format(dateLayout),
			expectedErr: nil,
		},
		{
			name:        "1daysago",
			input:       "1daysago",
			expected:    now.AddDate(0, 0, -1).Format(dateLayout),
			expectedErr: nil,
		},
		{
			name:        "10DAYSAGO",
			input:       "10DAYSAGO",
			expected:    now.AddDate(0, 0, -10).Format(dateLayout),
			expectedErr: nil,
		},
		{
			name:        "2023-10-27",
			input:       "2023-10-27",
			expected:    "2023-10-27",
			expectedErr: nil,
		},
		{
			name:        "invalid date format",
			input:       "27-10-2023",
			expected:    "",
			expectedErr: errors.New(`cannot parse "27-10-2023": invalid date format (YYYY-MM-DD)`),
		},
		{
			name:        "empty date daysago",
			input:       "daysago",
			expected:    "",
			expectedErr: errors.New(`cannot parse "daysago": invalid daysago format (NdaysAgo)`),
		},
		{
			name:        "invalid daysago format",
			input:       "xxdaysago",
			expected:    "",
			expectedErr: errors.New(`cannot parse "xxdaysago": invalid daysago format (NdaysAgo)`),
		},
		{
			name:        "invalid",
			input:       "invalid",
			expected:    "",
			expectedErr: errors.New(`cannot parse "invalid": invalid date format (YYYY-MM-DD)`),
		},
		{
			name:        "empty",
			input:       "",
			expected:    "",
			expectedErr: errors.New(`cannot parse "": invalid date format (YYYY-MM-DD)`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDate(tt.input)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
