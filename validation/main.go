package main

import (
	"log"
	"net/http"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

var Users map[string]*User

func main() {
	Users = make(map[string]*User)

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/users", getUserHandle)

	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}
