package api

import (
	"fmt"
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

// Config represents the API information
type Config struct {
	RootPrefix string    `json:"rootPrefix"`
	Port       string    `json:"port"`
	Endpoints  Endpoints `json:"endpoints"`
}

// Endpoints represent the endpoints that are exposed by the server
type Endpoints struct {
	Probes     string `json:"probes"`
	Projects   string `json:"projects"`
	Scenarios  string `json:"scenarios"`
	TestPlans  string `json:"testplans"`
	Executions string `json:"executions"`
}

// Validate that the config object is correct
func (c Config) Validate() error {
	errs := &errList{}
	if c.RootPrefix == "" {
		errs.add("address field is mandatory")
	}
	if c.Port == "" {
		errs.add("dataBase field is mandatory")
	}
	if err := c.Endpoints.Validate(); err != nil {
		errs.add(err.Error())
	}
	if !errs.isEmpty() {
		return errs
	}
	return nil
}

// Validate if all collections have been passed
func (c Endpoints) Validate() error {
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
	if c.Probes == "" {
		errs.add("probes field is mandatory")
	}
	if c.Executions == "" {
		errs.add("executions field is mandatory")
	}
	if !errs.isEmpty() {
		return errs
	}
	return nil
}
