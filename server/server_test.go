package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	db "k8s-backend/database"
	m "k8s-backend/model"

	"github.com/stretchr/testify/require"
)

const url = "http://localhost:8081"
const registerRoute = "/register"

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
				require.Error(t, validateUser(tt.user))
			} else {
				require.NoError(t, validateUser(tt.user))
			}
		})
	}
}

func TestRegisterHandler(t *testing.T) {
	// start server in separate goroutine
	go func() {
		server := &UserServer{
			Port: ":8081",
			DB:   &db.Cache[m.User]{},
		}
		server.Run()
	}()
	time.Sleep(5 * time.Second)

	user := &m.User{Name: "John", Email: "john@work.com", Age: 35}
	data, err := json.Marshal(user)
	if err != nil {
		t.Error(err)
	}
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		url+registerRoute,
		bytes.NewReader(data),
	)
	if err != nil {
		t.Error(err)
	}

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
	}
	require.Equal(t, 200, rsp.StatusCode)
}
