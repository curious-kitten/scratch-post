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
	"github.com/curious-kitten/scratch-post/pkg/metadata"
	"github.com/curious-kitten/scratch-post/pkg/scenarios"
	mockScenarios "github.com/curious-kitten/scratch-post/pkg/scenarios/mocks"
)

var (
	identity = metadata.Identity{
		ID:           "aabbccddee",
		Type:         "scenario",
		Version:      1,
		CreatedBy:    "author",
		UpdatedBy:    "author",
		CreationTime: time.Now(),
		UpdateTime:   time.Now(),
	}

	testScenario = &scenarios.Scenario{
		Name:      "test scenario",
		ProjectID: "zzxxxccvv",
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

func TestScenario_AddIdentity(t *testing.T) {
	g := NewWithT(t)
	s := scenarios.Scenario{}
	s.AddIdentity(&identity)
	g.Expect(s.GetIdentity()).To(Equal(&identity))
}

func TestScenario_Validate(t *testing.T) {
	g := NewWithT(t)
	s := &scenarios.Scenario{}
	err := s.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with empty scenario")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "empty scenario error is not a validation error")
	s.Name = "Test Name"
	err = s.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with scenario that only has a name")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "scenario with only name error is not a validation error")
	s.ProjectID = "aabbccdd"
	err = s.Validate()
	g.Expect(err).ShouldNot(HaveOccurred(), "error occurred when minimun requirements have been met")
}

func TestNew_Create(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockScenarios.NewMockIdentityGenerator(ctrl)
	mockGenerator.
		EXPECT().
		AddMeta("tester", "scenario", matchers.OfType(&scenarios.Scenario{})).
		Return(nil).
		Do(func(author string, objType string, identifiable metadata.Identifiable) {
			identifiable.AddIdentity(&identity)
		})
	mockAdder := mockScenarios.NewMockAdder(ctrl)
	mockAdder.
		EXPECT().
		AddOne(ctx, matchers.OfType(&scenarios.Scenario{})).
		Return(nil)

	creator := scenarios.New(mockGenerator, mockAdder, goodGetProject)
	scenario, err := creator(ctx, "tester", transformers.ToReadCloser(testScenario))
	g.Expect(err).ShouldNot(HaveOccurred(), "unexpected error occurred")
	expectedScenario := &scenarios.Scenario{
		Identity:  &identity,
		Name:      testScenario.Name,
		ProjectID: testScenario.ProjectID,
	}
	g.Expect(scenario).To(Equal(expectedScenario), "scenarios did not match")
}

func TestNew_ProjectNotFound(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockScenarios.NewMockIdentityGenerator(ctrl)
	mockAdder := mockScenarios.NewMockAdder(ctrl)
	creator := scenarios.New(mockGenerator, mockAdder, noProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testScenario))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "project not found error is not a validation error")
}

func TestNew_ProjectError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockScenarios.NewMockIdentityGenerator(ctrl)
	mockAdder := mockScenarios.NewMockAdder(ctrl)
	creator := scenarios.New(mockGenerator, mockAdder, errorGetProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testScenario))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeFalse(), "error type was missing")
}

func TestNew_MarshallError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockScenarios.NewMockIdentityGenerator(ctrl)
	mockAdder := mockScenarios.NewMockAdder(ctrl)
	creator := scenarios.New(mockGenerator, mockAdder, errorGetProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(struct{ SomeField string }{SomeField: "test"}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestNew_ValidationError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockScenarios.NewMockIdentityGenerator(ctrl)
	mockAdder := mockScenarios.NewMockAdder(ctrl)
	creator := scenarios.New(mockGenerator, mockAdder, errorGetProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(&scenarios.Scenario{}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestNew_AddMetaError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockScenarios.NewMockIdentityGenerator(ctrl)
	mockGenerator.
		EXPECT().
		AddMeta("tester", "scenario", matchers.OfType(&scenarios.Scenario{})).
		Return(fmt.Errorf("identity error"))
	mockAdder := mockScenarios.NewMockAdder(ctrl)
	creator := scenarios.New(mockGenerator, mockAdder, goodGetProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testScenario))
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestNew_AddToCollectionError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockScenarios.NewMockIdentityGenerator(ctrl)
	mockGenerator.
		EXPECT().
		AddMeta("tester", "scenario", matchers.OfType(&scenarios.Scenario{})).
		Return(nil).
		Do(func(author string, objType string, identifiable metadata.Identifiable) error {
			identifiable.AddIdentity(&identity)
			return nil
		})
	mockAdder := mockScenarios.NewMockAdder(ctrl)
	mockAdder.
		EXPECT().
		AddOne(ctx, matchers.OfType(&scenarios.Scenario{})).
		Return(fmt.Errorf("expected error"))

	creator := scenarios.New(mockGenerator, mockAdder, goodGetProject)
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
		GetAll(ctx, matchers.OfType(&[]scenarios.Scenario{})).
		Return(nil)

	lister := scenarios.List(mockGetter)
	_, err := lister(ctx)
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
		GetAll(ctx, matchers.OfType(&[]scenarios.Scenario{})).
		Return(fmt.Errorf("expected error"))

	lister := scenarios.List(mockGetter)
	_, err := lister(ctx)
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
		Get(ctx, identity.ID, matchers.OfType(&scenarios.Scenario{})).
		Return(nil)
	getter := scenarios.Get(mockGetter)
	_, err := getter(ctx, identity.ID)
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
		Get(ctx, identity.ID, matchers.OfType(&scenarios.Scenario{})).
		Return(fmt.Errorf("expected error"))

	getter := scenarios.Get(mockGetter)
	_, err := getter(ctx, identity.ID)
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
		Delete(ctx, identity.ID).
		Return(nil)
	deleter := scenarios.Delete(mockDeleter)
	err := deleter(ctx, identity.ID)
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
		Delete(ctx, identity.ID).
		Return(fmt.Errorf("returned error"))
	deleter := scenarios.Delete(mockDeleter)
	err := deleter(ctx, identity.ID)
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
		Get(ctx, identity.ID, matchers.OfType(&scenarios.Scenario{})).
		Do(func(ctx context.Context, id string, tp *scenarios.Scenario) {
			tp.Identity = &identity
		})
	mockReaderUpdater.
		EXPECT().
		Update(ctx, identity.ID, matchers.OfType(&scenarios.Scenario{}))
	updater := scenarios.Update(mockReaderUpdater, goodGetProject)
	scenario, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testScenario))
	g.Expect(err).ShouldNot(HaveOccurred(), "unexpected error occurred")
	expectedScenario := &scenarios.Scenario{
		Identity:  &identity,
		Name:      testScenario.Name,
		ProjectID: testScenario.ProjectID,
	}
	g.Expect(scenario).To(Equal(expectedScenario), "scenarios did not match")
}

func TestUpdate_ValidationError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockScenarios.NewMockReaderUpdater(ctrl)
	updater := scenarios.Update(mockReaderUpdater, goodGetProject)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(scenarios.Scenario{}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestUpdate_InvalidProject(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockScenarios.NewMockReaderUpdater(ctrl)
	updater := scenarios.Update(mockReaderUpdater, noProject)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testScenario))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "project not found error is not a validation error")
}

func TestUpdate_ProjectError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockScenarios.NewMockReaderUpdater(ctrl)
	updater := scenarios.Update(mockReaderUpdater, errorGetProject)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testScenario))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
	g.Expect(metadata.IsValidationError(err)).To(BeFalse(), "project not found error is not a validation error")
}

func TestUpdate_GetError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockScenarios.NewMockReaderUpdater(ctrl)
	mockReaderUpdater.
		EXPECT().
		Get(ctx, identity.ID, matchers.OfType(&scenarios.Scenario{})).
		Return(fmt.Errorf("error during get"))
	updater := scenarios.Update(mockReaderUpdater, goodGetProject)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testScenario))
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
		Get(ctx, identity.ID, matchers.OfType(&scenarios.Scenario{})).
		Do(func(ctx context.Context, id string, tp *scenarios.Scenario) {
			tp.Identity = &identity
		})
	mockReaderUpdater.
		EXPECT().
		Update(ctx, identity.ID, matchers.OfType(&scenarios.Scenario{})).
		Return(fmt.Errorf("update error"))
	updater := scenarios.Update(mockReaderUpdater, goodGetProject)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testScenario))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
}
