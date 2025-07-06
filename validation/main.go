package main

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func main() {
	server := &Server{
		Database: make(map[string]*User),
	}
	server.Run()
}
