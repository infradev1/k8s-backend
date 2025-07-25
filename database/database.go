package database

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database[T any] interface {
	Initialize() error
	Close()
	Get(id string) (*T, error)
	GetAll(limit, offset int, filters map[string]string) ([]*T, error)
	Insert(id string, element *T) error
	Update(id string, fields map[string]any) error
	Delete(id string) error
}

type Postgres[T any] struct {
	DB           *gorm.DB
	InitElements []T
	sync.Mutex
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
			if err := p.DB.Create(&e).Error; err != nil {
				return err
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
	p.Lock()
	defer p.Unlock()

	var record T
	if err := p.DB.First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (p *Postgres[T]) GetAll(limit, offset int, filters map[string]string) ([]*T, error) {
	p.Lock()
	defer p.Unlock()

	query := p.DB.Model(new(T))
	for k, v := range filters {
		if n, err := strconv.ParseFloat(v, 32); err == nil {
			query = query.Where(fmt.Sprintf("%s >= ?", k), n)
		} else {
			query = query.Where(fmt.Sprintf("LOWER(%s) ILIKE ?", k), "%"+strings.ToLower(v)+"%")
		}
	}

	var records []*T
	if err := query.Limit(limit).Offset(offset).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("error finding records: %w", err)
	}
	return records, nil
}

func (p *Postgres[T]) Insert(_ string, element *T) error {
	p.Lock()
	defer p.Unlock()

	// GORM handles primary key auto-increment
	if err := p.DB.Create(element).Error; err != nil {
		return err
	}

	return nil
}

func (p *Postgres[T]) Update(id string, fields map[string]any) error {
	p.Lock()
	defer p.Unlock()

	// `db.Model(&Post{}).Where("id = ?", id).Updates(updates)` updates the fields in the database.
	// `updates` contains the fields and values to be updated for the post with the specified ID.
	if err := p.DB.Model(new(T)).Where("id = ?", id).Updates(fields).Error; err != nil {
		return err
	}
	return nil
}

func (p *Postgres[T]) Delete(id string) error {
	p.Lock()
	defer p.Unlock()

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
		return nil, errors.New("record not found")
	}
	return element, nil
}

func (c *Cache[T]) GetAll(limit, offset int, filters map[string]string) ([]*T, error) {
	c.Lock()
	defer c.Unlock()

	var records []*T
	skipped := 0

	for _, v := range c.Data {
		if skipped < offset {
			skipped++
			continue
		}
		if len(records) == limit {
			break
		}
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
