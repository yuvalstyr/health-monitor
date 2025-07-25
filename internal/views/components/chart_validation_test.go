package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateUpdateInterval(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid seconds",
			input:    "10s",
			expected: "10s",
		},
		{
			name:     "valid minutes",
			input:    "5m",
			expected: "5m",
		},
		{
			name:     "valid hours",
			input:    "2h",
			expected: "2h",
		},
		{
			name:     "valid large number",
			input:    "120s",
			expected: "120s",
		},
		{
			name:     "invalid - no unit",
			input:    "10",
			expected: "30s",
		},
		{
			name:     "invalid - wrong unit",
			input:    "10x",
			expected: "30s",
		},
		{
			name:     "invalid - starts with zero",
			input:    "0s",
			expected: "30s",
		},
		{
			name:     "invalid - negative number",
			input:    "-5s",
			expected: "30s",
		},
		{
			name:     "invalid - decimal number",
			input:    "1.5s",
			expected: "30s",
		},
		{
			name:     "invalid - empty string",
			input:    "",
			expected: "30s",
		},
		{
			name:     "invalid - malicious input",
			input:    "10s; alert('xss')",
			expected: "30s",
		},
		{
			name:     "invalid - contains spaces",
			input:    "10 s",
			expected: "30s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateUpdateInterval(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}