package execution

import (
	scenariov1 "github.com/curious-kitten/scratch-post/pkg/api/v1/scenario"
	"github.com/curious-kitten/scratch-post/pkg/errors"
)

// Validate is used to check the integrity of the execution object
func (e *Execution) Validate() error {
	if e.TestPlanId == "" {
		return errors.NewValidationError("testplanID is a mandatory parameter")
	}
	if e.ScenarioId == "" {
		return errors.NewValidationError("scenarioID is a mandatory parameter")
	}
	if e.ProjectId == "" {
		return errors.NewValidationError("projectId is a mandatory parameter")
	}
	return nil
}

// PopulateSteps the Execution stepts given scenario steps
func (e *Execution) PopulateSteps(s []*scenariov1.Step) {
	e.Steps = make([]*StepExecution, len(s))
	for i, v := range s {
		e.Steps[i] = &StepExecution{Definition: v, Status: Status_Pending}
	}
}
