package executions_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/curious-kitten/scratch-post/internal/test/matchers"
	"github.com/curious-kitten/scratch-post/internal/test/transformers"
	execution "github.com/curious-kitten/scratch-post/pkg/api/v1/execution"
	metadata "github.com/curious-kitten/scratch-post/pkg/api/v1/metadata"
	scenario "github.com/curious-kitten/scratch-post/pkg/api/v1/scenario"
	"github.com/curious-kitten/scratch-post/pkg/errors"
	"github.com/curious-kitten/scratch-post/pkg/executions"
	mockExecutions "github.com/curious-kitten/scratch-post/pkg/executions/mocks"
)

var (
	sortBy            = ""
	reverse           = false
	count             = 1000
	previousLastValue = ""
	identity          = metadata.Identity{
		Id:           "aabbccddee",
		Type:         "execution",
		Version:      1,
		CreatedBy:    "author",
		UpdatedBy:    "author",
		CreationTime: time.Now().Unix(),
		UpdateTime:   time.Now().Unix(),
	}

	testExecution = &execution.Execution{
		ProjectId:  "zzxxxccvv",
		ScenarioId: "qwertyuiop",
		TestPlanId: "zxcvbnm",
		Steps: []*execution.StepExecution{
			{
				Definition: &scenario.Step{
					Position: 1,
					Name:     "test",
				},
				Status: execution.Status_Pending,
			},
			{
				Definition: &scenario.Step{
					Position: 2,
					Name:     "test",
				},
				Status: execution.Status_Pending,
			},
		},
	}
)

func goodGetItem(ctx context.Context, id string) (interface{}, error) {
	return nil, nil
}

func getScenario(ctx context.Context, id string) (interface{}, error) {
	return &scenario.Scenario{
		Name:      "test scenario",
		ProjectId: "zzxxxccvv",
		Steps: []*scenario.Step{
			{
				Position: 1,
				Name:     "test",
			},
			{
				Position: 2,
				Name:     "test",
			},
		},
	}, nil
}

func errorGetItem(ctx context.Context, id string) (interface{}, error) {
	return nil, fmt.Errorf("an error")
}

func noItem(ctx context.Context, id string) (interface{}, error) {
	return nil, mongo.ErrNoDocuments
}

func TestExecution_Validate(t *testing.T) {
	g := NewWithT(t)
	s := &execution.Execution{}
	err := s.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with empty execution")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "empty execution error is not a validation error")

	s = &execution.Execution{
		ProjectId:  "zzxxxccvv",
		ScenarioId: "qwertyuio",
	}
	err = s.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with execution that does not have a project ID")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "execution without a scenario ID is not a validation error")

	s = &execution.Execution{
		TestPlanId: "zzxxxccvv",
		ScenarioId: "qwertyuio",
	}
	err = s.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with execution that does not have a project ID")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "execution without a test plan ID is not a validation error")

	s = &execution.Execution{
		TestPlanId: "zzxxxccvv",
		ProjectId:  "qwertyuio",
	}
	err = s.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with execution that does not have a scenario ID")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "execution without a test plan ID is not a validation error")

	err = testExecution.Validate()
	g.Expect(err).ShouldNot(HaveOccurred(), "error occurred when minimun requirements have been met")
}

func TestNew_Create(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mockExecutions.NewMockMetaHandler(ctrl)
	mockMetaHandler.
		EXPECT().
		NewMeta("tester", "execution").
		Return(&identity, nil)
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	mockAdder.
		EXPECT().
		AddOne(ctx, matchers.OfType(&execution.Execution{})).
		Return(nil)

	creator := executions.New(mockMetaHandler, mockAdder, goodGetItem, getScenario, goodGetItem)
	createdExecution, err := creator(ctx, "tester", transformers.ToReadCloser(testExecution))
	g.Expect(err).ShouldNot(HaveOccurred(), "unexpected error occurred")
	expectedExecution := &execution.Execution{
		Identity:   &identity,
		ProjectId:  testExecution.ProjectId,
		ScenarioId: testExecution.ScenarioId,
		TestPlanId: testExecution.TestPlanId,
		Status:     execution.Status_Pending,
		Steps:      testExecution.Steps,
	}
	g.Expect(createdExecution).To(Equal(expectedExecution), "executions did not match")
}

func TestNew_ProjectNotFound(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mockExecutions.NewMockMetaHandler(ctrl)
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	creator := executions.New(mockMetaHandler, mockAdder, noItem, goodGetItem, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "project not found error is not a validation error")
}

func TestNew_ExecutorNotFound(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mockExecutions.NewMockMetaHandler(ctrl)
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	creator := executions.New(mockMetaHandler, mockAdder, goodGetItem, goodGetItem, noItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "executor not found error is not a validation error")
}

func TestNew_ScenarioNotFound(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mockExecutions.NewMockMetaHandler(ctrl)
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	creator := executions.New(mockMetaHandler, mockAdder, goodGetItem, noItem, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "scenario not found error is not a validation error")
}

func TestNew_ProjectError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mockExecutions.NewMockMetaHandler(ctrl)
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	creator := executions.New(mockMetaHandler, mockAdder, errorGetItem, goodGetItem, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(errors.IsValidationError(err)).To(BeFalse(), "error type was missing")
}

func TestNew_ExecutorError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mockExecutions.NewMockMetaHandler(ctrl)
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	creator := executions.New(mockMetaHandler, mockAdder, goodGetItem, goodGetItem, errorGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(errors.IsValidationError(err)).To(BeFalse(), "error type was missing")
}

func TestNew_ScenarioError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mockExecutions.NewMockMetaHandler(ctrl)
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	creator := executions.New(mockMetaHandler, mockAdder, goodGetItem, errorGetItem, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(errors.IsValidationError(err)).To(BeFalse(), "error type was missing")
}

func TestNew_MarshallError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mockExecutions.NewMockMetaHandler(ctrl)
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	creator := executions.New(mockMetaHandler, mockAdder, errorGetItem, goodGetItem, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(struct{ SomeField string }{SomeField: "test"}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestNew_ValidationError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mockExecutions.NewMockMetaHandler(ctrl)
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	creator := executions.New(mockMetaHandler, mockAdder, goodGetItem, goodGetItem, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(&execution.Execution{}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestNew_AddMetaError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mockExecutions.NewMockMetaHandler(ctrl)
	mockMetaHandler.
		EXPECT().
		NewMeta("tester", "execution").
		Return(nil, fmt.Errorf("identity error"))
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	creator := executions.New(mockMetaHandler, mockAdder, goodGetItem, getScenario, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestNew_AddToCollectionError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mockExecutions.NewMockMetaHandler(ctrl)
	mockMetaHandler.
		EXPECT().
		NewMeta("tester", "execution").
		Return(&identity, nil)
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	mockAdder.
		EXPECT().
		AddOne(ctx, matchers.OfType(&execution.Execution{})).
		Return(fmt.Errorf("expected error"))

	creator := executions.New(mockMetaHandler, mockAdder, goodGetItem, getScenario, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestList(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGetter := mockExecutions.NewMockGetter(ctrl)
	mockGetter.
		EXPECT().
		GetAll(ctx, matchers.OfType(&[]execution.Execution{}), map[string][]string{}, sortBy, reverse, count, previousLastValue).
		Return(nil)

	lister := executions.List(mockGetter)
	_, err := lister(ctx, map[string][]string{}, sortBy, reverse, count, previousLastValue)
	g.Expect(err).ShouldNot(HaveOccurred(), "expected error did not occur")
}

func TestList_RetrieveError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGetter := mockExecutions.NewMockGetter(ctrl)
	mockGetter.
		EXPECT().
		GetAll(ctx, matchers.OfType(&[]execution.Execution{}), map[string][]string{}, sortBy, reverse, count, previousLastValue).
		Return(fmt.Errorf("expected error"))

	lister := executions.List(mockGetter)
	_, err := lister(ctx, map[string][]string{}, sortBy, reverse, count, previousLastValue)
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestGet(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGetter := mockExecutions.NewMockGetter(ctrl)
	mockGetter.
		EXPECT().
		Get(ctx, identity.Id, matchers.OfType(&execution.Execution{})).
		Return(nil)
	getter := executions.Get(mockGetter)
	_, err := getter(ctx, identity.Id)
	g.Expect(err).ShouldNot(HaveOccurred(), "expected error did not occur")
}

func TestGet_Error(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGetter := mockExecutions.NewMockGetter(ctrl)
	mockGetter.
		EXPECT().
		Get(ctx, identity.Id, matchers.OfType(&execution.Execution{})).
		Return(fmt.Errorf("expected error"))

	getter := executions.Get(mockGetter)
	_, err := getter(ctx, identity.Id)
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestUpdate(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockExecutions.NewMockReaderUpdater(ctrl)
	mockReaderUpdater.
		EXPECT().
		Get(ctx, identity.Id, matchers.OfType(&execution.Execution{})).
		Do(func(ctx context.Context, id string, e *execution.Execution) {
			e.Steps = testExecution.Steps
			e.Identity = &identity
		})
	mockReaderUpdater.
		EXPECT().
		Update(ctx, identity.Id, matchers.OfType(&execution.Execution{}))
	mockMetaHandler := mockExecutions.NewMockMetaHandler(ctrl)
	mockMetaHandler.EXPECT().UpdateMeta("tester", matchers.OfType(&metadata.Identity{}))
	updater := executions.Update(mockMetaHandler, mockReaderUpdater, goodGetItem, goodGetItem, goodGetItem)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testExecution))
	g.Expect(err).ShouldNot(HaveOccurred(), "unexpected error occurred")
}

func TestUpdate_ValidationError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockExecutions.NewMockReaderUpdater(ctrl)
	mockMetaHandler := mockExecutions.NewMockMetaHandler(ctrl)
	updater := executions.Update(mockMetaHandler, mockReaderUpdater, goodGetItem, goodGetItem, goodGetItem)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(execution.Execution{}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestUpdate_InvalidProject(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockExecutions.NewMockReaderUpdater(ctrl)
	mockMetaHandler := mockExecutions.NewMockMetaHandler(ctrl)
	updater := executions.Update(mockMetaHandler, mockReaderUpdater, noItem, goodGetItem, goodGetItem)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "project not found error is not a validation error")
}

func TestUpdate_InvalidScenario(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockExecutions.NewMockReaderUpdater(ctrl)
	mockMetaHandler := mockExecutions.NewMockMetaHandler(ctrl)
	updater := executions.Update(mockMetaHandler, mockReaderUpdater, goodGetItem, noItem, goodGetItem)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "project not found error is not a validation error")
}

func TestUpdate_InvalidTestPlan(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockExecutions.NewMockReaderUpdater(ctrl)
	mockMetaHandler := mockExecutions.NewMockMetaHandler(ctrl)
	updater := executions.Update(mockMetaHandler, mockReaderUpdater, goodGetItem, noItem, goodGetItem)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "project not found error is not a validation error")
}

func TestUpdate_ProjectError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockExecutions.NewMockReaderUpdater(ctrl)
	mockMetaHandler := mockExecutions.NewMockMetaHandler(ctrl)
	updater := executions.Update(mockMetaHandler, mockReaderUpdater, errorGetItem, goodGetItem, goodGetItem)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
	g.Expect(errors.IsValidationError(err)).To(BeFalse(), "project not found error is not a validation error")
}

func TestUpdate_ScenarioError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockExecutions.NewMockReaderUpdater(ctrl)
	mockMetaHandler := mockExecutions.NewMockMetaHandler(ctrl)
	updater := executions.Update(mockMetaHandler, mockReaderUpdater, goodGetItem, errorGetItem, goodGetItem)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
	g.Expect(errors.IsValidationError(err)).To(BeFalse(), "project not found error is not a validation error")
}

func TestUpdate_TestPlanError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockExecutions.NewMockReaderUpdater(ctrl)
	mockMetaHandler := mockExecutions.NewMockMetaHandler(ctrl)
	updater := executions.Update(mockMetaHandler, mockReaderUpdater, goodGetItem, goodGetItem, errorGetItem)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
	g.Expect(errors.IsValidationError(err)).To(BeFalse(), "project not found error is not a validation error")
}

func TestUpdate_GetError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockExecutions.NewMockReaderUpdater(ctrl)
	mockReaderUpdater.
		EXPECT().
		Get(ctx, identity.Id, matchers.OfType(&execution.Execution{})).
		Return(fmt.Errorf("error during get"))
	mockMetaHandler := mockExecutions.NewMockMetaHandler(ctrl)
	updater := executions.Update(mockMetaHandler, mockReaderUpdater, goodGetItem, goodGetItem, goodGetItem)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
}

func TestUpdate_UpdateError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockExecutions.NewMockReaderUpdater(ctrl)
	mockReaderUpdater.
		EXPECT().
		Get(ctx, identity.Id, matchers.OfType(&execution.Execution{})).
		Do(func(ctx context.Context, id string, e *execution.Execution) {
			e.Steps = testExecution.Steps
			e.Identity = &identity
		})
	mockReaderUpdater.
		EXPECT().
		Update(ctx, identity.Id, matchers.OfType(&execution.Execution{})).
		Return(fmt.Errorf("update error"))
	mockMetaHandler := mockExecutions.NewMockMetaHandler(ctrl)
	mockMetaHandler.EXPECT().UpdateMeta("tester", matchers.OfType(&metadata.Identity{}))
	updater := executions.Update(mockMetaHandler, mockReaderUpdater, goodGetItem, goodGetItem, goodGetItem)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
}
