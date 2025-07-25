package database

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"sync"

	m "k8s-backend/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database[T any] interface {
	Initialize() error
	Close()
	Get(id string) (*T, error)
	GetAll(f *m.Filters[T]) ([]*T, error)
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

func (p *Postgres[T]) GetAll(f *m.Filters[T]) ([]*T, error) {
	p.Lock()
	defer p.Unlock()

	t := reflect.TypeOf(*f.Model)
	v := reflect.ValueOf(*f.Model)

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("invalid model: %v", t.Kind().String())
	}

	query := p.DB.Model(new(T))
	query = query.Order(f.SortBy + " " + f.Order)

	for i := range t.NumField() {
		field := t.Field(i)
		value := v.Field(i)

		if !value.CanInterface() {
			continue
		}
		fieldValue := value.Interface()

		switch f := fieldValue.(type) {
		case string:
			if f != "" {
				query = query.Where(fmt.Sprintf("LOWER(%s) ILIKE ?", field.Name), "%"+strings.ToLower(f)+"%")
			}
		case float64:
			if f > 0 {
				query = query.Where(fmt.Sprintf("%s >= ?", field.Name), f)
			}
		case int:
			if column := strings.ToLower(field.Name); column != "id" {
				query = query.Where(fmt.Sprintf("%s >= ?", column), f)
			}
		default:
			return nil, fmt.Errorf("model field data type not supported: %v", f)
		}
	}

	query = query.Debug() // Enable SQL logging

	var records []*T
	if err := query.Limit(f.Limit).Offset(f.Offset).Find(&records).Error; err != nil {
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

func (c *Cache[T]) GetAll(f *m.Filters[T]) ([]*T, error) {
	c.Lock()
	defer c.Unlock()

	var records []*T
	skipped := 0

	for _, v := range c.Data {
		if skipped < f.Offset {
			skipped++
			continue
		}
		if len(records) == f.Limit {
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
