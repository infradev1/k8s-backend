package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "User registration requires POST", http.StatusBadRequest)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Request body must contain name, email, and age", http.StatusBadRequest)
		return
	}

	if err := validateUser(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userId := uuid.NewString()
	Users[userId] = &user

	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintf(w, "User %v created successfully", userId); err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
}

func validateUser(user *User) error {
	if user.Name == "" {
		return fmt.Errorf("User name must not be empty")
	}
	if !strings.Contains(user.Email, "@") {
		return fmt.Errorf("User email must be valid")
	}
	if user.Age <= 21 {
		return fmt.Errorf("User age must be greater than 21")
	}
	return nil
}

func getUserHandle(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Query parameter 'id' must be provided", http.StatusBadRequest)
		return
	}

	user := Users[id]
	if user == nil {
		http.Error(w, fmt.Sprintf("User %s not found", id), http.StatusNotFound)
		return
	}

	data, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		http.Error(w, "Error marshaling struct into JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		http.Error(w, "Error writing response JSON", http.StatusInternalServerError)
		return
	}
}
