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
