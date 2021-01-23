package store

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
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

type Config struct {
	Address     string      `json:"address"`
	DataBase    string      `json:"database"`
	Collections Collections `json:"collections"`
}

type Collections struct {
	Projects  string `json:"projects"`
	Scenarios string `json:"scenarios"`
	TestPlans string `json:"testplans"`
}

// Validate that the config object is correct
func (c Config) Validate() error {
	errs := &errList{}
	if c.Address == "" {
		errs.add("address field is mandatory")
	}
	if c.DataBase == "" {
		errs.add("dataBase field is mandatory")
	}
	if err := c.Collections.Validate(); err != nil {
		errs.add(err.Error())
	}
	if !errs.isEmpty() {
		return errs
	}
	return nil
}

// Validate if all collections have been passed
func (c Collections) Validate() error {
	errs := &errList{}
	if c.Projects == "" {
		errs.add("projects field is mandatory")
	}
	if c.Scenarios == "" {
		errs.add("scenarios field is mandatory")
	}
	if c.TestPlans == "" {
		errs.add("testplans field is mandatory")
	}
	if !errs.isEmpty() {
		return errs
	}
	return nil
}

// NewConfig returns a the necessary information to connect to the Data Base
func NewConfig(data io.Reader) (Config, error) {
	config := Config{}
	decoder := json.NewDecoder(data)
	if err := decoder.Decode(&config); err != nil {
		return Config{}, err
	}
	if err := config.Validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}
