package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateUser(t *testing.T) {
	// table-driven tests for the various cases
	tests := []struct {
		name          string
		user          *User
		expectedError bool
	}{
		{
			name:          "Valid user",
			user:          &User{"John", "john@work.com", 35},
			expectedError: false,
		},
		{
			name:          "Empty user",
			user:          &User{},
			expectedError: true,
		},
		{
			name:          "Invalid user name",
			user:          &User{"H", "john@work.com", 35},
			expectedError: true,
		},
		{
			name:          "Invalid email",
			user:          &User{"John", "work.com", 35},
			expectedError: true,
		},
		{
			name:          "Invalid age",
			user:          &User{"John", "john@work.com", 15},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedError {
				require.Error(t, validateUser(tt.user))
			} else {
				require.NoError(t, validateUser(tt.user))
			}
		})
	}
}
