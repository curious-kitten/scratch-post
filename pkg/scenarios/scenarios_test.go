package scenarios_test

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
	metadata "github.com/curious-kitten/scratch-post/pkg/api/v1/metadata"
	scenario "github.com/curious-kitten/scratch-post/pkg/api/v1/scenario"
	"github.com/curious-kitten/scratch-post/pkg/errors"
	"github.com/curious-kitten/scratch-post/pkg/scenarios"
	mockScenarios "github.com/curious-kitten/scratch-post/pkg/scenarios/mocks"
)

var (
	sortBy            = ""
	reverse           = false
	count             = 1000
	previousLastValue = ""
	identity          = metadata.Identity{
		Id:           "aabbccddee",
		Type:         "scenario",
		Version:      1,
		CreatedBy:    "author",
		UpdatedBy:    "author",
		CreationTime: time.Now().Unix(),
		UpdateTime:   time.Now().Unix(),
	}

	testScenario = &scenario.Scenario{
		Name:      "test scenario",
		ProjectId: "zzxxxccvv",
	}
)

func goodGetProject(ctx context.Context, id string) (interface{}, error) {
	return nil, nil
}

func errorGetProject(ctx context.Context, id string) (interface{}, error) {
	return nil, fmt.Errorf("an error")
}

func noProject(ctx context.Context, id string) (interface{}, error) {
	return nil, mongo.ErrNoDocuments
}

func TestScenario_Validate(t *testing.T) {
	g := NewWithT(t)
	s := &scenario.Scenario{}
	err := s.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with empty scenario")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "empty scenario error is not a validation error")
	s.Name = "Test Name"
	err = s.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with scenario that only has a name")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "scenario with only name error is not a validation error")
	s.ProjectId = "aabbccdd"
	err = s.Validate()
	g.Expect(err).ShouldNot(HaveOccurred(), "error occurred when minimun requirements have been met")
}

func TestNew_Create(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mockScenarios.NewMockMetaHandler(ctrl)
	mockMetaHandler.
		EXPECT().
		NewMeta("tester", "scenario").
		Return(&identity, nil)

	mockAdder := mockScenarios.NewMockAdder(ctrl)
	mockAdder.
		EXPECT().
		AddOne(ctx, matchers.OfType(&scenario.Scenario{})).
		Return(nil)

	creator := scenarios.New(mockMetaHandler, mockAdder, goodGetProject)
	createdScenario, err := creator(ctx, "tester", transformers.ToReadCloser(testScenario))
	g.Expect(err).ShouldNot(HaveOccurred(), "unexpected error occurred")
	expectedScenario := &scenario.Scenario{
		Identity:  &identity,
		Name:      testScenario.Name,
		ProjectId: testScenario.ProjectId,
	}
	g.Expect(createdScenario).To(Equal(expectedScenario), "scenarios did not match")
}

func TestNew_ProjectNotFound(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mockScenarios.NewMockMetaHandler(ctrl)
	mockAdder := mockScenarios.NewMockAdder(ctrl)
	creator := scenarios.New(mockMetaHandler, mockAdder, noProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testScenario))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "project not found error is not a validation error")
}

func TestNew_ProjectError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mockScenarios.NewMockMetaHandler(ctrl)
	mockAdder := mockScenarios.NewMockAdder(ctrl)
	creator := scenarios.New(mockMetaHandler, mockAdder, errorGetProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testScenario))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(errors.IsValidationError(err)).To(BeFalse(), "error type was missing")
}

func TestNew_MarshallError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mockScenarios.NewMockMetaHandler(ctrl)
	mockAdder := mockScenarios.NewMockAdder(ctrl)
	creator := scenarios.New(mockMetaHandler, mockAdder, errorGetProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(struct{ SomeField string }{SomeField: "test"}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestNew_ValidationError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mockScenarios.NewMockMetaHandler(ctrl)
	mockAdder := mockScenarios.NewMockAdder(ctrl)
	creator := scenarios.New(mockMetaHandler, mockAdder, errorGetProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(&scenario.Scenario{}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestNew_NewMetaError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mockScenarios.NewMockMetaHandler(ctrl)
	mockMetaHandler.
		EXPECT().
		NewMeta("tester", "scenario").
		Return(nil, fmt.Errorf("identity error"))
	mockAdder := mockScenarios.NewMockAdder(ctrl)
	creator := scenarios.New(mockMetaHandler, mockAdder, goodGetProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testScenario))
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestNew_AddToCollectionError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mockScenarios.NewMockMetaHandler(ctrl)
	mockMetaHandler.
		EXPECT().
		NewMeta("tester", "scenario").
		Return(&identity, nil)
	mockAdder := mockScenarios.NewMockAdder(ctrl)
	mockAdder.
		EXPECT().
		AddOne(ctx, matchers.OfType(&scenario.Scenario{})).
		Return(fmt.Errorf("expected error"))

	creator := scenarios.New(mockMetaHandler, mockAdder, goodGetProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testScenario))
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestList(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGetter := mockScenarios.NewMockGetter(ctrl)
	mockGetter.
		EXPECT().
		GetAll(ctx, matchers.OfType(&[]scenario.Scenario{}), map[string][]string{}, sortBy, reverse, count, previousLastValue).
		Return(nil)

	lister := scenarios.List(mockGetter)
	_, err := lister(ctx, map[string][]string{}, sortBy, reverse, count, previousLastValue)
	g.Expect(err).ShouldNot(HaveOccurred(), "expected error did not occur")
}

func TestList_RetrieveError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGetter := mockScenarios.NewMockGetter(ctrl)
	mockGetter.
		EXPECT().
		GetAll(ctx, matchers.OfType(&[]scenario.Scenario{}), map[string][]string{}, sortBy, reverse, count, previousLastValue).
		Return(fmt.Errorf("expected error"))

	lister := scenarios.List(mockGetter)
	_, err := lister(ctx, map[string][]string{}, sortBy, reverse, count, previousLastValue)
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestGet(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGetter := mockScenarios.NewMockGetter(ctrl)
	mockGetter.
		EXPECT().
		Get(ctx, identity.Id, matchers.OfType(&scenario.Scenario{})).
		Return(nil)
	getter := scenarios.Get(mockGetter)
	_, err := getter(ctx, identity.Id)
	g.Expect(err).ShouldNot(HaveOccurred(), "expected error did not occur")
}

func TestGet_Error(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGetter := mockScenarios.NewMockGetter(ctrl)
	mockGetter.
		EXPECT().
		Get(ctx, identity.Id, matchers.OfType(&scenario.Scenario{})).
		Return(fmt.Errorf("expected error"))

	getter := scenarios.Get(mockGetter)
	_, err := getter(ctx, identity.Id)
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestDelete(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockDeleter := mockScenarios.NewMockDeleter(ctrl)
	mockDeleter.
		EXPECT().
		Delete(ctx, identity.Id).
		Return(nil)
	deleter := scenarios.Delete(mockDeleter)
	err := deleter(ctx, identity.Id)
	g.Expect(err).ShouldNot(HaveOccurred(), "expected error did not occur")
}

func TestDelete_Error(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockDeleter := mockScenarios.NewMockDeleter(ctrl)
	mockDeleter.
		EXPECT().
		Delete(ctx, identity.Id).
		Return(fmt.Errorf("returned error"))
	deleter := scenarios.Delete(mockDeleter)
	err := deleter(ctx, identity.Id)
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestUpdate(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockScenarios.NewMockReaderUpdater(ctrl)
	mockReaderUpdater.
		EXPECT().
		Get(ctx, identity.Id, matchers.OfType(&scenario.Scenario{})).
		Do(func(ctx context.Context, id string, tp *scenario.Scenario) {
			tp.Identity = &identity
		})
	mockReaderUpdater.
		EXPECT().
		Update(ctx, identity.Id, matchers.OfType(&scenario.Scenario{}))
	mockMetaHandler := mockScenarios.NewMockMetaHandler(ctrl)
	mockMetaHandler.EXPECT().UpdateMeta("tester", matchers.OfType(&metadata.Identity{}))
	updater := scenarios.Update(mockMetaHandler, mockReaderUpdater, goodGetProject)
	createdScenario, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testScenario))
	g.Expect(err).ShouldNot(HaveOccurred(), "unexpected error occurred")
	expectedScenario := &scenario.Scenario{
		Identity:  &identity,
		Name:      testScenario.Name,
		ProjectId: testScenario.ProjectId,
	}
	g.Expect(createdScenario).To(Equal(expectedScenario), "scenarios did not match")
}

func TestUpdate_ValidationError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockScenarios.NewMockReaderUpdater(ctrl)
	mockMetaHandler := mockScenarios.NewMockMetaHandler(ctrl)
	updater := scenarios.Update(mockMetaHandler, mockReaderUpdater, goodGetProject)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(scenario.Scenario{}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestUpdate_InvalidProject(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockScenarios.NewMockReaderUpdater(ctrl)
	mockMetaHandler := mockScenarios.NewMockMetaHandler(ctrl)
	updater := scenarios.Update(mockMetaHandler, mockReaderUpdater, noProject)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testScenario))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "project not found error is not a validation error")
}

func TestUpdate_ProjectError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockScenarios.NewMockReaderUpdater(ctrl)
	mockMetaHandler := mockScenarios.NewMockMetaHandler(ctrl)
	updater := scenarios.Update(mockMetaHandler, mockReaderUpdater, errorGetProject)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testScenario))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
	g.Expect(errors.IsValidationError(err)).To(BeFalse(), "project not found error is not a validation error")
}

func TestUpdate_GetError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockScenarios.NewMockReaderUpdater(ctrl)
	mockReaderUpdater.
		EXPECT().
		Get(ctx, identity.Id, matchers.OfType(&scenario.Scenario{})).
		Return(fmt.Errorf("error during get"))
	mockMetaHandler := mockScenarios.NewMockMetaHandler(ctrl)
	updater := scenarios.Update(mockMetaHandler, mockReaderUpdater, goodGetProject)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testScenario))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
}

func TestUpdate_UpdateError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockScenarios.NewMockReaderUpdater(ctrl)
	mockReaderUpdater.
		EXPECT().
		Get(ctx, identity.Id, matchers.OfType(&scenario.Scenario{})).
		Do(func(ctx context.Context, id string, tp *scenario.Scenario) {
			tp.Identity = &identity
		})
	mockReaderUpdater.
		EXPECT().
		Update(ctx, identity.Id, matchers.OfType(&scenario.Scenario{})).
		Return(fmt.Errorf("update error"))
	mockMetaHandler := mockScenarios.NewMockMetaHandler(ctrl)
	mockMetaHandler.EXPECT().UpdateMeta("tester", matchers.OfType(&metadata.Identity{}))
	updater := scenarios.Update(mockMetaHandler, mockReaderUpdater, goodGetProject)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testScenario))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
}
