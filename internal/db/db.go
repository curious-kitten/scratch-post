package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

type errList struct {
	b strings.Builder
}

func (e *errList) add(err string) {
	e.b.WriteString(fmt.Sprintf("%s\n", err))
}

func (e *errList) isEmpty() bool {
	return e.b.Len() == 0
}

func (e *errList) Error() string {
	return e.b.String()
}

// Config represents the DB coonection information
type Config struct {
	Address     string      `json:"address"`
	Connections Connections `json:"connections"`
}

// Validate that the config object is correct
func (c Config) Validate() error {
	errs := &errList{}
	if c.Address == "" {
		errs.add("address field is mandatory")
	}
	if !errs.isEmpty() {
		return errs
	}
	return nil
}

// Connections stores the DB connection information
type Connections struct {
	MaxLifetime int `json:"maxLifetime"`
	MaxOpen     int `json:"maxOpen"`
	MaxIdle     int `json:"maxIdle"`
}

// New instantiates a new DB instance based on the configuration
func New(config Config) (*sql.DB, error) {
	pgURL, err := pq.ParseURL(config.Address)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("postgres", pgURL)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Duration(config.Connections.MaxLifetime) * time.Second)
	db.SetMaxOpenConns(config.Connections.MaxOpen)
	db.SetMaxIdleConns(config.Connections.MaxIdle)
	return db, err
}
