package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type Server struct {
	Port string
	DB   Database[User]
}

func (s *Server) RegisterHandler(w http.ResponseWriter, r *http.Request) {
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
	if err := s.DB.Insert(userId, &user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintf(w, "User %v created successfully", userId); err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
}

func validateUser(user *User) error {
	if len(user.Name) < 3 {
		return fmt.Errorf("User name must have 3+ characters")
	}
	if !strings.Contains(user.Email, "@") {
		return fmt.Errorf("User email must be valid")
	}
	if user.Age <= 21 {
		return fmt.Errorf("User age must be greater than 21")
	}
	return nil
}

func (s *Server) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Query parameter 'id' must be provided", http.StatusBadRequest)
		return
	}

	user, err := s.DB.Get(id)
	if err != nil {
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

func (s *Server) Run() {
	http.HandleFunc("/register", s.RegisterHandler)
	http.HandleFunc("/users", s.GetUserHandler)

	if err := http.ListenAndServe(s.Port, nil); err != nil {
		log.Fatal(err)
	}
}
