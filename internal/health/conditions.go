package health

import (
	"fmt"
	"sync"

	"github.com/curious-kitten/scratch-post/internal/info"
)

// ConditionCheck defines the API for functions that check the condition of the system
type ConditionCheck func() Condition

// Condition contains the information regarding the status of a piece of the System
type Condition struct {
	Ready   bool   `json:"ready"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name"`
}

func alwaysGood() Condition {
	return Condition{
		Ready:   true,
		Message: "",
		Name:    "Default",
	}
}

// NewConditions creates a Conditions object, for which to register health and readiness conditions
func NewConditions(app info.App, instance info.Instance) *Conditions {
	return &Conditions{
		app:              app,
		instance:         instance,
		healthConditions: []ConditionCheck{alwaysGood},
		readyConditions:  []ConditionCheck{alwaysGood},
	}
}

// Status models the status of the app and the meta information of the instance
type Status struct {
	App      info.App      `json:"app"`
	Instance info.Instance `json:"instance"`
	Status   []Condition   `json:"status"`
}

// Conditions holds the health and readiness conditions and verifies if they have been met
type Conditions struct {
	app              info.App
	instance         info.Instance
	healthConditions []ConditionCheck
	readyConditions  []ConditionCheck
}

// RegisterHealthCondition adds a condition that must pass in order for the health probe to respond with 200
func (c *Conditions) RegisterHealthCondition(condition ConditionCheck) {
	c.healthConditions = append(c.healthConditions, condition)
}

// RegisterReadynessCondition adds a condition that must pass in order for the readiness probe to respond with 200
func (c *Conditions) RegisterReadynessCondition(condition ConditionCheck) {
	c.readyConditions = append(c.readyConditions, condition)
}

// IsReady checks whether the ready conditions are met
func (c *Conditions) IsReady() (bool, interface{}) {
	return c.check(c.readyConditions)
}

// IsAlive checks whether the health conditions are met
func (c *Conditions) IsAlive() (bool, interface{}) {
	return c.check(c.healthConditions)
}

// Check verfies if the conditions have been met and returns the status of the conditions
func (c *Conditions) check(conditionCheck []ConditionCheck) (bool, interface{}) {
	ready := true
	status := Status{
		App:      c.app,
		Status:   []Condition{},
		Instance: c.instance,
	}
	results := make(chan Condition)
	wg := &sync.WaitGroup{}
	wg.Add(len(conditionCheck))
	fmt.Println(len(conditionCheck))
	for _, v := range conditionCheck {
		go func(check ConditionCheck) {
			defer wg.Done()
			results <- check()
		}(v)
	}
	wg2 := &sync.WaitGroup{}
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		for cond := range results {
			status.Status = append(status.Status, cond)
			if !cond.Ready {
				ready = false
			}
		}
	}()
	wg.Wait()
	close(results)
	wg2.Wait()

	return ready, status
}
