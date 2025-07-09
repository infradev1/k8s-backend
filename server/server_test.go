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

	// start server in separate goroutine
	go func() {
		server := &Server{
			Port:     ":8081",
			Services: []Service{userSvc},
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
