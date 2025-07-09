package services

import (
	"testing"

	m "k8s-backend/model"

	"github.com/stretchr/testify/require"
)

func TestValidateUser(t *testing.T) {
	// table-driven tests for the various cases
	tests := []struct {
		name          string
		user          *m.User
		expectedError bool
	}{
		{
			name:          "Valid user",
			user:          &m.User{Name: "John", Email: "john@work.com", Age: 35},
			expectedError: false,
		},
		{
			name:          "Empty user",
			user:          &m.User{},
			expectedError: true,
		},
		{
			name:          "Invalid user name",
			user:          &m.User{Name: "H", Email: "john@work.com", Age: 35},
			expectedError: true,
		},
		{
			name:          "Invalid email",
			user:          &m.User{Name: "John", Email: "work.com", Age: 35},
			expectedError: true,
		},
		{
			name:          "Invalid age",
			user:          &m.User{Name: "John", Email: "john@work.com", Age: 15},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedError {
				require.Error(t, ValidateUser(tt.user))
			} else {
				require.NoError(t, ValidateUser(tt.user))
			}
		})
	}
}
