package services

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"

	db "k8s-backend/database"
	m "k8s-backend/model"

	"github.com/google/uuid"
)

type UserService struct {
	DB db.Database[m.User]
}

func (s *UserService) Init() {
	if err := s.DB.Initialize(); err != nil {
		slog.Error(err.Error())
		log.Fatal(fmt.Errorf("failed to initialize database: %w", err))
	}
}

func (s *UserService) SetupEndpoints() {
	http.HandleFunc("/register", s.RegisterUserHandler)
	http.HandleFunc("/users", s.GetUserHandler)
}

func (s *UserService) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusBadRequest)
		return
	}

	var user m.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Request body must contain name, email, and age", http.StatusBadRequest)
		return
	}

	if err := ValidateUser(&user); err != nil {
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

func ValidateUser(user *m.User) error {
	if len(user.Name) < 3 {
		return fmt.Errorf("user name must have 3+ characters")
	}
	if !strings.Contains(user.Email, "@") {
		return fmt.Errorf("user email must be valid")
	}
	if user.Age <= 21 {
		return fmt.Errorf("user age must be greater than 21")
	}
	return nil
}

func (s *UserService) GetUserHandler(w http.ResponseWriter, r *http.Request) {
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
