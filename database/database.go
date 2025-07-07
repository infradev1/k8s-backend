package database

import (
	"fmt"
	"log"
	"log/slog"
	"sync"

	m "k8s-backend/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitDatabase() {
	dsn := "host:localhost user=postgres password=postgres dbname=booksdb port=5432 sslmode=disable"

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		slog.Error(err.Error())
		log.Fatal(fmt.Errorf("Failed to connect to the database: %w", err))
	}

	if err = db.AutoMigrate(&m.Book{}); err != nil {
		log.Fatal(fmt.Errorf("Failed to migrate database schema: %w", err))
	}

	initBooks := []m.Book{
		{Title: "QM", Author: "Bohr", Price: 10.99},
		{Title: "QFT", Author: "Dirac", Price: 11.99},
		{Title: "GR", Author: "Einstein", Price: 12.99},
	}

	for i, book := range initBooks {
		var existingBook m.Book
		result := db.First(&existingBook, i+1)
		if result.RowsAffected == 0 {
			if r := db.Create(&book); r.Error != nil {
				slog.Error(r.Error.Error())
			}
		}
	}

	slog.Info("Database connection established")
}

type Database[T any] interface {
	Get(id string) (*T, error)
	Insert(id string, element *T) error
	Update(id string, element *T) error
	Delete(id string) error
}

type Cache[T any] struct {
	Data map[string]*T
	sync.Mutex
}

func (c *Cache[T]) Get(id string) (*T, error) {
	c.Lock()
	defer c.Unlock()
	element := c.Data[id]
	if element == nil {
		return nil, fmt.Errorf("%s not found", id)
	}
	return element, nil
}

func (c *Cache[T]) Insert(id string, element *T) error {
	c.Lock()
	defer c.Unlock()
	if e := c.Data[id]; e == nil {
		c.Data[id] = element
		return nil
	}
	return fmt.Errorf("%s already exists", id)
}

func (c *Cache[T]) Update(id string, element *T) error {
	c.Lock()
	defer c.Unlock()
	if e := c.Data[id]; e == nil {
		return fmt.Errorf("%s does not exist", id)
	}
	c.Data[id] = element
	return nil
}

func (c *Cache[T]) Delete(id string) error {
	c.Lock()
	defer c.Unlock()
	if e := c.Data[id]; e == nil {
		return fmt.Errorf("%s does not exist", id)
	}
	delete(c.Data, id)
	return nil
}
