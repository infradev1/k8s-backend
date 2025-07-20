package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	db "k8s-backend/database"
	m "k8s-backend/model"
	svc "k8s-backend/services"

	"github.com/stretchr/testify/require"
)

const url = "http://localhost:8081"
const registerRoute = "/register"

func TestRegisterHandler(t *testing.T) {
	userSvc := &svc.UserService{
		DB: &db.Cache[m.User]{},
	}
	userSvc.Init()
	defer userSvc.DB.Close()

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

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(userSvc.RegisterUserHandler)
	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	t.Log(rr.Body.String())
}
