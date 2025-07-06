package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const url = "http://localhost:8081"
const registerRoute = "/register"

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

func TestRegisterHandler(t *testing.T) {
	// start server in separate goroutine
	go func() {
		server := &Server{
			Port: ":8081",
			DB: &Cache[User]{
				Data: make(map[string]*User),
			},
		}
		server.Run()
	}()
	time.Sleep(5 * time.Second)

	user := &User{"John", "john@work.com", 35}
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
