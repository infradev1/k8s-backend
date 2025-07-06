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
			name:          "Missing user name",
			user:          &User{},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, validateUser(tt.user))
		})
	}
}
