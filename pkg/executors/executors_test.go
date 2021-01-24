package executors_test

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
	"github.com/curious-kitten/scratch-post/pkg/executors"
	mockExecutors "github.com/curious-kitten/scratch-post/pkg/executors/mocks"
)

var (
	identity = metadata.Identity{
		ID:           "aabbccddee",
		Type:         "executor",
		Version:      1,
		CreatedBy:    "author",
		UpdatedBy:    "author",
		CreationTime: time.Now(),
		UpdateTime:   time.Now(),
	}

	testExecutor = &executors.Executor{
		ProjectID: "zzxxxccvv",
		TestPlanID: "aabbccddeeff",
		ScenarioID: "qwertyuio",
	}
)

func goodGetItem(ctx context.Context, id string) (interface{}, error) {
	return nil, nil
}

func errorGetItem(ctx context.Context, id string) (interface{}, error) {
	return nil, fmt.Errorf("an error")
}

func noItem(ctx context.Context, id string) (interface{}, error) {
	return nil, mongo.ErrNoDocuments
}

func TestExecutor_AddIdentity(t *testing.T) {
	g := NewWithT(t)
	s := executors.Executor{}
	s.AddIdentity(&identity)
	g.Expect(s.GetIdentity()).To(Equal(&identity))
}

func TestExecutor_Validate(t *testing.T) {
	g := NewWithT(t)
	s := &executors.Executor{}
	err := s.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with empty executor")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "empty executor error is not a validation error")

	s = &executors.Executor{
		ProjectID: "zzxxxccvv",
		TestPlanID: "aabbccddeeff",
	}
	err = s.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with executor that does not have a scenario ID")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "executor without a scenario ID is not a validation error")

	s = &executors.Executor{
		ProjectID: "zzxxxccvv",
		ScenarioID: "qwertyuio",
	}
	err = s.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with executor that does not have a test plan ID")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "executor without a test plan ID is not a validation error")

	s = &executors.Executor{
		TestPlanID: "aabbccddeeff",
		ScenarioID: "qwertyuio",
	}
	err = s.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with executor that does not have a project ID")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "executor without a project ID is not a validation error")


	err = testExecutor.Validate()
	g.Expect(err).ShouldNot(HaveOccurred(), "error occurred when minimun requirements have been met")
}

func TestNew_Create(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutors.NewMockIdentityGenerator(ctrl)
	mockGenerator.
		EXPECT().
		AddMeta("tester", "executor", matchers.OfType(&executors.Executor{})).
		Return(nil).
		Do(func(author string, objType string, identifiable metadata.Identifiable) {
			identifiable.AddIdentity(&identity)
		})
	mockAdder := mockExecutors.NewMockAdder(ctrl)
	mockAdder.
		EXPECT().
		AddOne(ctx, matchers.OfType(&executors.Executor{})).
		Return(nil)

	creator := executors.New(mockGenerator, mockAdder, goodGetItem, goodGetItem, goodGetItem)
	executor, err := creator(ctx, "tester", transformers.ToReadCloser(testExecutor))
	g.Expect(err).ShouldNot(HaveOccurred(), "unexpected error occurred")
	expectedExecutor := &executors.Executor{
		Identity:  &identity,
		ProjectID: testExecutor.ProjectID,
		ScenarioID: testExecutor.ScenarioID,
		TestPlanID: testExecutor.TestPlanID,
	}
	g.Expect(executor).To(Equal(expectedExecutor), "executors did not match")
}

func TestNew_ProjectNotFound(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutors.NewMockIdentityGenerator(ctrl)
	mockAdder := mockExecutors.NewMockAdder(ctrl)
	creator := executors.New(mockGenerator, mockAdder, noItem, goodGetItem, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecutor))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "project not found error is not a validation error")
}

func TestNew_TestPlanNotFound(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutors.NewMockIdentityGenerator(ctrl)
	mockAdder := mockExecutors.NewMockAdder(ctrl)
	creator := executors.New(mockGenerator, mockAdder, goodGetItem, goodGetItem, noItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecutor))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "project not found error is not a validation error")
}

func TestNew_ScenarioNotFound(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutors.NewMockIdentityGenerator(ctrl)
	mockAdder := mockExecutors.NewMockAdder(ctrl)
	creator := executors.New(mockGenerator, mockAdder, goodGetItem, noItem, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecutor))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "project not found error is not a validation error")
}

func TestNew_ProjectError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutors.NewMockIdentityGenerator(ctrl)
	mockAdder := mockExecutors.NewMockAdder(ctrl)
	creator := executors.New(mockGenerator, mockAdder, errorGetItem, goodGetItem, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecutor))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeFalse(), "error type was missing")
}

func TestNew_TestPlanError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutors.NewMockIdentityGenerator(ctrl)
	mockAdder := mockExecutors.NewMockAdder(ctrl)
	creator := executors.New(mockGenerator, mockAdder, goodGetItem, goodGetItem, errorGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecutor))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeFalse(), "error type was missing")
}

func TestNew_ScenarioError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutors.NewMockIdentityGenerator(ctrl)
	mockAdder := mockExecutors.NewMockAdder(ctrl)
	creator := executors.New(mockGenerator, mockAdder, goodGetItem, errorGetItem, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecutor))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeFalse(), "error type was missing")
}

func TestNew_MarshallError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutors.NewMockIdentityGenerator(ctrl)
	mockAdder := mockExecutors.NewMockAdder(ctrl)
	creator := executors.New(mockGenerator, mockAdder, errorGetItem, goodGetItem, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(struct{ SomeField string }{SomeField: "test"}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestNew_ValidationError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutors.NewMockIdentityGenerator(ctrl)
	mockAdder := mockExecutors.NewMockAdder(ctrl)
	creator := executors.New(mockGenerator, mockAdder, goodGetItem, goodGetItem, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(&executors.Executor{}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestNew_AddMetaError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutors.NewMockIdentityGenerator(ctrl)
	mockGenerator.
		EXPECT().
		AddMeta("tester", "executor", matchers.OfType(&executors.Executor{})).
		Return(fmt.Errorf("identity error"))
	mockAdder := mockExecutors.NewMockAdder(ctrl)
	creator := executors.New(mockGenerator, mockAdder, goodGetItem, goodGetItem, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecutor))
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestNew_AddToCollectionError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockExecutors.NewMockIdentityGenerator(ctrl)
	mockGenerator.
		EXPECT().
		AddMeta("tester", "executor", matchers.OfType(&executors.Executor{})).
		Return(nil).
		Do(func(author string, objType string, identifiable metadata.Identifiable) error {
			identifiable.AddIdentity(&identity)
			return nil
		})
	mockAdder := mockExecutors.NewMockAdder(ctrl)
	mockAdder.
		EXPECT().
		AddOne(ctx, matchers.OfType(&executors.Executor{})).
		Return(fmt.Errorf("expected error"))

	creator := executors.New(mockGenerator, mockAdder, goodGetItem, goodGetItem, goodGetItem)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testExecutor))
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestList(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGetter := mockExecutors.NewMockGetter(ctrl)
	mockGetter.
		EXPECT().
		GetAll(ctx, matchers.OfType(&[]executors.Executor{})).
		Return(nil)

	lister := executors.List(mockGetter)
	_, err := lister(ctx)
	g.Expect(err).ShouldNot(HaveOccurred(), "expected error did not occur")
}

func TestList_RetrieveError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGetter := mockExecutors.NewMockGetter(ctrl)
	mockGetter.
		EXPECT().
		GetAll(ctx, matchers.OfType(&[]executors.Executor{})).
		Return(fmt.Errorf("expected error"))

	lister := executors.List(mockGetter)
	_, err := lister(ctx)
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestGet(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGetter := mockExecutors.NewMockGetter(ctrl)
	mockGetter.
		EXPECT().
		Get(ctx, identity.ID, matchers.OfType(&executors.Executor{})).
		Return(nil)
	getter := executors.Get(mockGetter)
	_, err := getter(ctx, identity.ID)
	g.Expect(err).ShouldNot(HaveOccurred(), "expected error did not occur")
}

func TestGet_Error(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGetter := mockExecutors.NewMockGetter(ctrl)
	mockGetter.
		EXPECT().
		Get(ctx, identity.ID, matchers.OfType(&executors.Executor{})).
		Return(fmt.Errorf("expected error"))

	getter := executors.Get(mockGetter)
	_, err := getter(ctx, identity.ID)
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

