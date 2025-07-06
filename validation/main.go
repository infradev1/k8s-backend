package main

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func main() {
	server := &Server{
		Port:     ":8081",
		Database: make(map[string]*User),
	}
	server.Run()
}
