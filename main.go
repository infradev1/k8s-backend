package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type User struct {
	Name  string
	Age   int
	Email string
}

var users map[int]*User

func main() {
	users = make(map[int]*User)
	for i := range 5 {
		users[i] = &User{fmt.Sprintf("user-%d", i), 100, fmt.Sprintf("user-%d@email.com", i)}
	}

	// Register handlers
	http.HandleFunc("/", greet)
	http.HandleFunc("/users", userInfo)

	// Serve on port 8080
	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// TODO: Define greet and userInfo handler functions here

func greet(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Hello"))
	if err != nil {
		fmt.Printf("Failed to write greeting: %v\n", err)
	}
}

func userInfo(w http.ResponseWriter, r *http.Request) {
	// query parameter for user ID (http://localhost:8080/users?id=0)
	userId := r.URL.Query().Get("id")
	if userId == "" {
		w.Write([]byte("Please provide integer query parameter: id"))
		return
	}
	id, err := strconv.Atoi(userId)
	if err != nil {
		fmt.Fprintf(w, "Invalid user id: %s", userId)
		return
	}
	user := users[id]
	if user == nil {
		fmt.Fprintf(w, "Please provide a user id between 0 and %d", len(users)-1)
		return
	}

	data, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	w.Write(data)
}
