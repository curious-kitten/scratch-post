package execution

import (
	"github.com/curious-kitten/scratch-post/internal/decoder"
	scenariov1 "github.com/curious-kitten/scratch-post/pkg/api/v1/scenario"
)

// Validate is used to check the integrity of the execution object
func (e *Execution) Validate() error {
	if e.TestPlanId == "" {
		return decoder.NewValidationError("testplanID is a mandatory parameter")
	}
	if e.ScenarioId == "" {
		return decoder.NewValidationError("scenarioID is a mandatory parameter")
	}
	if e.ProjectId == "" {
		return decoder.NewValidationError("projectId is a mandatory parameter")
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
