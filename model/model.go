package model

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type Book struct {
	Id        int     `json:"id" gorm:"primaryKey;autoIncrement"`
	Title     string  `json:"title" gorm:"unique"`
	Author    string  `json:"author" gorm:"size:255"`
	Price     float64 `json:"price"`
	CreatedAt string  `json:"created_at"`
}

type Result struct {
	Value any   `json:"value"`
	Error error `json:"error,omitempty"`
}

type FleetHealthStatus struct {
	Networking bool `json:"networking"`
	DataCenter bool `json:"data_center"`
	Kubernetes bool `json:"kubernetes"`
}

type Region = string

type Filters[T any] struct {
	Model  *T     `json:"model"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	SortBy string `json:"sort_by"`
	Order  string `json:"order"`
}
