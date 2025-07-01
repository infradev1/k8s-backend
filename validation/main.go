package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

var users map[uuid.UUID]*User

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

	userId := uuid.New()
	users[userId] = &user
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintf(w, "User %v created successfully", userId); err != nil {
		log.Fatal(err)
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

func main() {
	users = make(map[uuid.UUID]*User)

	http.HandleFunc("/register", registerHandler)

	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}
