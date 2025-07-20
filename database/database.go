package database

import (
	"fmt"
	"log/slog"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database[T any] interface {
	Initialize() error
	Close()
	Get(id string) (*T, error)
	GetAll() ([]*T, error)
	Insert(id string, element *T) error
	Update(id string, fields map[string]any) error
	Delete(id string) error
}

type Postgres[T any] struct {
	DB           *gorm.DB
	InitElements []T
}

func (p *Postgres[T]) Initialize() error {
	dsn := "host=localhost user=carloslara password=postgres dbname=postgres port=5432 sslmode=disable"

	var err error
	p.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	if err = p.DB.AutoMigrate(new(T)); err != nil {
		return err
	}

	for i, e := range p.InitElements {
		var existing T
		result := p.DB.First(&existing, i+1)
		if result.RowsAffected == 0 {
			if r := p.DB.Create(&e); r.Error != nil {
				return r.Error
			}
		}
	}

	slog.Info("Database connection established")

	return nil
}

func (p *Postgres[T]) Close() {
	sqlDB, err := p.DB.DB()
	if err != nil {
		slog.Error(err.Error())
	}
	if err := sqlDB.Close(); err != nil {
		slog.Error(err.Error())
	}
}

func (p *Postgres[T]) Get(id string) (*T, error) {
	var record T
	result := p.DB.First(&record, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

func (p *Postgres[T]) GetAll() ([]*T, error) {
	var records []*T
	if err := p.DB.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("error finding records: %w", err)
	}
	return records, nil
}

func (p *Postgres[T]) Insert(_ string, element *T) error {
	// GORM handles primary key auto-increment
	if err := p.DB.Create(element).Error; err != nil {
		return err
	}

	return nil
}

func (p *Postgres[T]) Update(id string, fields map[string]any) error {
	// `db.Model(&Post{}).Where("id = ?", id).Updates(updates)` updates the fields in the database.
	// `updates` contains the fields and values to be updated for the post with the specified ID.
	if err := p.DB.Model(new(T)).Where("id = ?", id).Updates(fields).Error; err != nil {
		return err
	}
	return nil
}

func (p *Postgres[T]) Delete(id string) error {
	if err := p.DB.Delete(new(T), id).Error; err != nil {
		return err
	}
	return nil
}

type Cache[T any] struct {
	Data map[string]*T
	sync.Mutex
}

func (c *Cache[T]) Initialize() error {
	c.Data = map[string]*T{
		"0": new(T),
		"1": new(T),
		"2": new(T),
	}
	return nil
}

func (c *Cache[T]) Close() {
	clear(c.Data)
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

func (c *Cache[T]) GetAll() ([]*T, error) {
	c.Lock()
	defer c.Unlock()
	var records []*T
	for _, v := range c.Data {
		records = append(records, v)
	}
	return records, nil
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

func (c *Cache[T]) Update(id string, fields map[string]any) error {
	c.Lock()
	defer c.Unlock()
	if e := c.Data[id]; e == nil {
		return fmt.Errorf("%s does not exist", id)
	}
	c.Data[id] = new(T) // mock
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
