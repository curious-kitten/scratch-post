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
	"github.com/curious-kitten/scratch-post/pkg/executions"
	mockExecutions "github.com/curious-kitten/scratch-post/pkg/executions/mocks"
	"github.com/curious-kitten/scratch-post/pkg/metadata"
	"github.com/curious-kitten/scratch-post/pkg/scenarios"
)

var (
	identity = metadata.Identity{
		ID:           "aabbccddee",
		Type:         "execution",
		Version:      1,
		CreatedBy:    "author",
		UpdatedBy:    "author",
		CreationTime: time.Now(),
		UpdateTime:   time.Now(),
	}

	testExecution = &executions.Execution{
		ProjectID:  "zzxxxccvv",
		ScenarioID: "qwertyuiop",
		TestPlanID: "zxcvbnm",
		Steps: []*executions.Step{
			{
				Step: scenarios.Step{
					Position: 1,
					Name:     "test",
				},
				Status: executions.Pending,
			},
			{
				Step: scenarios.Step{
					Position: 2,
					Name:     "test",
				},
				Status: executions.Pending,
			},
		},
	}
)

func goodGetItem(ctx context.Context, id string) (interface{}, error) {
	return nil, nil
}

func getScenario(ctx context.Context, id string) (interface{}, error) {
	return &scenarios.Scenario{
		Name:      "test scenario",
		ProjectID: "zzxxxccvv",
		Steps: []scenarios.Step{
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

func TestExecution_AddIdentity(t *testing.T) {
	g := NewWithT(t)
	s := executions.Execution{}
	s.AddIdentity(&identity)
	g.Expect(s.GetIdentity()).To(Equal(&identity))
}

func TestExecution_Validate(t *testing.T) {
	g := NewWithT(t)
	s := &executions.Execution{}
	err := s.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with empty execution")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "empty execution error is not a validation error")

	s = &executions.Execution{
		ProjectID:  "zzxxxccvv",
		ScenarioID: "qwertyuio",
	}
	err = s.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with execution that does not have a project ID")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "execution without a scenario ID is not a validation error")

	s = &executions.Execution{
		TestPlanID: "zzxxxccvv",
		ScenarioID: "qwertyuio",
	}
	err = s.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with execution that does not have a project ID")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "execution without a test plan ID is not a validation error")

	s = &executions.Execution{
		TestPlanID: "zzxxxccvv",
		ProjectID:  "qwertyuio",
	}
	err = s.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with execution that does not have a scenario ID")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "execution without a test plan ID is not a validation error")

	err = testExecution.Validate()
	g.Expect(err).ShouldNot(HaveOccurred(), "error occurred when minimun requirements have been met")
}

func TestNew_Create(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutions.NewMockIdentityGenerator(ctrl)
	mockGenerator.
		EXPECT().
		AddMeta("tester", "execution", matchers.OfType(&executions.Execution{})).
		Return(nil).
		Do(func(author string, objType string, identifiable metadata.Identifiable) {
			identifiable.AddIdentity(&identity)
		})
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	mockAdder.
		EXPECT().
		AddOne(ctx, matchers.OfType(&executions.Execution{})).
		Return(nil)

	creator := executions.New(mockGenerator, mockAdder, goodGetItem, getScenario, goodGetItem)
	execution, err := creator(ctx, "tester", transformers.ToReadCloser(testExecution))
	g.Expect(err).ShouldNot(HaveOccurred(), "unexpected error occurred")
	expectedExecution := &executions.Execution{
		Identity:   &identity,
		ProjectID:  testExecution.ProjectID,
		ScenarioID: testExecution.ScenarioID,
		TestPlanID: testExecution.TestPlanID,
		Status:     executions.Pending,
		Steps:      testExecution.Steps,
	}
	g.Expect(execution).To(Equal(expectedExecution), "executions did not match")
}

func TestNew_ProjectNotFound(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutions.NewMockIdentityGenerator(ctrl)
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	creator := executions.New(mockGenerator, mockAdder, noItem, goodGetItem, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "project not found error is not a validation error")
}

func TestNew_ExecutorNotFound(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutions.NewMockIdentityGenerator(ctrl)
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	creator := executions.New(mockGenerator, mockAdder, goodGetItem, goodGetItem, noItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "executor not found error is not a validation error")
}

func TestNew_ScenarioNotFound(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutions.NewMockIdentityGenerator(ctrl)
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	creator := executions.New(mockGenerator, mockAdder, goodGetItem, noItem, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "scenario not found error is not a validation error")
}

func TestNew_ProjectError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutions.NewMockIdentityGenerator(ctrl)
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	creator := executions.New(mockGenerator, mockAdder, errorGetItem, goodGetItem, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeFalse(), "error type was missing")
}

func TestNew_ExecutorError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutions.NewMockIdentityGenerator(ctrl)
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	creator := executions.New(mockGenerator, mockAdder, goodGetItem, goodGetItem, errorGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeFalse(), "error type was missing")
}

func TestNew_ScenarioError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutions.NewMockIdentityGenerator(ctrl)
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	creator := executions.New(mockGenerator, mockAdder, goodGetItem, errorGetItem, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeFalse(), "error type was missing")
}

func TestNew_MarshallError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutions.NewMockIdentityGenerator(ctrl)
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	creator := executions.New(mockGenerator, mockAdder, errorGetItem, goodGetItem, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(struct{ SomeField string }{SomeField: "test"}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestNew_ValidationError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutions.NewMockIdentityGenerator(ctrl)
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	creator := executions.New(mockGenerator, mockAdder, goodGetItem, goodGetItem, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(&executions.Execution{}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestNew_AddMetaError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutions.NewMockIdentityGenerator(ctrl)
	mockGenerator.
		EXPECT().
		AddMeta("tester", "execution", matchers.OfType(&executions.Execution{})).
		Return(fmt.Errorf("identity error"))
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	creator := executions.New(mockGenerator, mockAdder, goodGetItem, getScenario, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestNew_AddToCollectionError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutions.NewMockIdentityGenerator(ctrl)
	mockGenerator.
		EXPECT().
		AddMeta("tester", "execution", matchers.OfType(&executions.Execution{})).
		Return(nil).
		Do(func(author string, objType string, identifiable metadata.Identifiable) error {
			identifiable.AddIdentity(&identity)
			return nil
		})
	mockAdder := mockExecutions.NewMockAdder(ctrl)
	mockAdder.
		EXPECT().
		AddOne(ctx, matchers.OfType(&executions.Execution{})).
		Return(fmt.Errorf("expected error"))

	creator := executions.New(mockGenerator, mockAdder, goodGetItem, getScenario, goodGetItem)
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
		GetAll(ctx, matchers.OfType(&[]executions.Execution{})).
		Return(nil)

	lister := executions.List(mockGetter)
	_, err := lister(ctx)
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
		GetAll(ctx, matchers.OfType(&[]executions.Execution{})).
		Return(fmt.Errorf("expected error"))

	lister := executions.List(mockGetter)
	_, err := lister(ctx)
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
		Get(ctx, identity.ID, matchers.OfType(&executions.Execution{})).
		Return(nil)
	getter := executions.Get(mockGetter)
	_, err := getter(ctx, identity.ID)
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
		Get(ctx, identity.ID, matchers.OfType(&executions.Execution{})).
		Return(fmt.Errorf("expected error"))

	getter := executions.Get(mockGetter)
	_, err := getter(ctx, identity.ID)
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
		Get(ctx, identity.ID, matchers.OfType(&executions.Execution{})).
		Do(func(ctx context.Context, id string, e *executions.Execution) {
			*e = *testExecution
			e.Identity = &identity
		})
	mockReaderUpdater.
		EXPECT().
		Update(ctx, identity.ID, matchers.OfType(&executions.Execution{}))
	updater := executions.Update(mockReaderUpdater, goodGetItem, goodGetItem, goodGetItem)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testExecution))
	g.Expect(err).ShouldNot(HaveOccurred(), "unexpected error occurred")
}

func TestUpdate_ValidationError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockExecutions.NewMockReaderUpdater(ctrl)
	updater := executions.Update(mockReaderUpdater, goodGetItem, goodGetItem, goodGetItem)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(executions.Execution{}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestUpdate_InvalidProject(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockExecutions.NewMockReaderUpdater(ctrl)
	updater := executions.Update(mockReaderUpdater, noItem, goodGetItem, goodGetItem)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "project not found error is not a validation error")
}

func TestUpdate_InvalidScenario(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockExecutions.NewMockReaderUpdater(ctrl)
	updater := executions.Update(mockReaderUpdater, goodGetItem, noItem, goodGetItem)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "project not found error is not a validation error")
}

func TestUpdate_InvalidTestPlant(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockExecutions.NewMockReaderUpdater(ctrl)
	updater := executions.Update(mockReaderUpdater, goodGetItem, goodGetItem, noItem)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "project not found error is not a validation error")
}

func TestUpdate_ProjectError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockExecutions.NewMockReaderUpdater(ctrl)
	updater := executions.Update(mockReaderUpdater, errorGetItem, goodGetItem, goodGetItem)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
	g.Expect(metadata.IsValidationError(err)).To(BeFalse(), "project not found error is not a validation error")
}

func TestUpdate_ScenarioError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockExecutions.NewMockReaderUpdater(ctrl)
	updater := executions.Update(mockReaderUpdater, goodGetItem, errorGetItem, goodGetItem)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
	g.Expect(metadata.IsValidationError(err)).To(BeFalse(), "project not found error is not a validation error")
}

func TestUpdate_TestPlanError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockExecutions.NewMockReaderUpdater(ctrl)
	updater := executions.Update(mockReaderUpdater, goodGetItem, goodGetItem, errorGetItem)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
	g.Expect(metadata.IsValidationError(err)).To(BeFalse(), "project not found error is not a validation error")
}

func TestUpdate_GetError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockExecutions.NewMockReaderUpdater(ctrl)
	mockReaderUpdater.
		EXPECT().
		Get(ctx, identity.ID, matchers.OfType(&executions.Execution{})).
		Return(fmt.Errorf("error during get"))
	updater := executions.Update(mockReaderUpdater, goodGetItem, goodGetItem, goodGetItem)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testExecution))
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
		Get(ctx, identity.ID, matchers.OfType(&executions.Execution{})).
		Do(func(ctx context.Context, id string, e *executions.Execution) {
			*e = *testExecution
			e.Identity = &identity
		})
	mockReaderUpdater.
		EXPECT().
		Update(ctx, identity.ID, matchers.OfType(&executions.Execution{})).
		Return(fmt.Errorf("update error"))
	updater := executions.Update(mockReaderUpdater, goodGetItem, goodGetItem, goodGetItem)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testExecution))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
}
