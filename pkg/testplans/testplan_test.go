package testplans_test

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
	testplan "github.com/curious-kitten/scratch-post/pkg/api/v1/testplan"
	"github.com/curious-kitten/scratch-post/pkg/errors"
	"github.com/curious-kitten/scratch-post/pkg/testplans"
	mocktestplans "github.com/curious-kitten/scratch-post/pkg/testplans/mocks"
)

var (
	identity = metadata.Identity{
		Id:           "aabbccddee",
		Type:         "testplan",
		Version:      1,
		CreatedBy:    "author",
		UpdatedBy:    "author",
		CreationTime: time.Now().Unix(),
		UpdateTime:   time.Now().Unix(),
	}

	testTestPlan = &testplan.TestPlan{
		Name:      "test testplan",
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

func TestTestPlan_Validate(t *testing.T) {
	g := NewWithT(t)
	s := &testplan.TestPlan{}
	err := s.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with empty testplan")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "empty testplan error is not a validation error")
	s.Name = "Test Name"
	err = s.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with testplan that only has a name")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "testplan with only name error is not a validation error")
	s.ProjectId = "aabbccdd"
	err = s.Validate()
	g.Expect(err).ShouldNot(HaveOccurred(), "error occurred when minimun requirements have been met")
}

func TestNew_Create(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mocktestplans.NewMockMetaHandler(ctrl)
	mockMetaHandler.
		EXPECT().
		NewMeta("tester", "testplan").
		Return(&identity, nil)
	mockAdder := mocktestplans.NewMockAdder(ctrl)
	mockAdder.
		EXPECT().
		AddOne(ctx, matchers.OfType(&testplan.TestPlan{})).
		Return(nil)

	creator := testplans.New(mockMetaHandler, mockAdder, goodGetProject)
	createdTestplan, err := creator(ctx, "tester", transformers.ToReadCloser(testTestPlan))
	g.Expect(err).ShouldNot(HaveOccurred(), "unexpected error occurred")
	expectedTestPlan := &testplan.TestPlan{
		Identity:  &identity,
		Name:      testTestPlan.Name,
		ProjectId: testTestPlan.ProjectId,
	}
	g.Expect(createdTestplan).To(Equal(expectedTestPlan), "testplans did not match")
}

func TestNew_ProjectNotFound(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mocktestplans.NewMockMetaHandler(ctrl)
	mockAdder := mocktestplans.NewMockAdder(ctrl)
	creator := testplans.New(mockMetaHandler, mockAdder, noProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testTestPlan))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "project not found error is not a validation error")
}

func TestNew_ProjectError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mocktestplans.NewMockMetaHandler(ctrl)
	mockAdder := mocktestplans.NewMockAdder(ctrl)
	creator := testplans.New(mockMetaHandler, mockAdder, errorGetProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testTestPlan))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(errors.IsValidationError(err)).To(BeFalse(), "error type was missing")
}

func TestNew_MarshallError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mocktestplans.NewMockMetaHandler(ctrl)
	mockAdder := mocktestplans.NewMockAdder(ctrl)
	creator := testplans.New(mockMetaHandler, mockAdder, errorGetProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(struct{ SomeField string }{SomeField: "test"}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestNew_ValidationError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mocktestplans.NewMockMetaHandler(ctrl)
	mockAdder := mocktestplans.NewMockAdder(ctrl)
	creator := testplans.New(mockMetaHandler, mockAdder, errorGetProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(&testplan.TestPlan{}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestNew_AddMetaError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mocktestplans.NewMockMetaHandler(ctrl)
	mockMetaHandler.
		EXPECT().
		NewMeta("tester", "testplan").
		Return(nil, fmt.Errorf("identity error"))
	mockAdder := mocktestplans.NewMockAdder(ctrl)
	creator := testplans.New(mockMetaHandler, mockAdder, goodGetProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testTestPlan))
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestNew_AddToCollectionError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockMetaHandler := mocktestplans.NewMockMetaHandler(ctrl)
	mockMetaHandler.
		EXPECT().
		NewMeta("tester", "testplan").
		Return(&identity, nil)
	mockAdder := mocktestplans.NewMockAdder(ctrl)
	mockAdder.
		EXPECT().
		AddOne(ctx, matchers.OfType(&testplan.TestPlan{})).
		Return(fmt.Errorf("expected error"))

	creator := testplans.New(mockMetaHandler, mockAdder, goodGetProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testTestPlan))
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestList(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGetter := mocktestplans.NewMockGetter(ctrl)
	mockGetter.
		EXPECT().
		GetAll(ctx, matchers.OfType(&[]testplan.TestPlan{})).
		Return(nil)

	lister := testplans.List(mockGetter)
	_, err := lister(ctx)
	g.Expect(err).ShouldNot(HaveOccurred(), "expected error did not occur")
}

func TestList_RetrieveError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGetter := mocktestplans.NewMockGetter(ctrl)
	mockGetter.
		EXPECT().
		GetAll(ctx, matchers.OfType(&[]testplan.TestPlan{})).
		Return(fmt.Errorf("expected error"))

	lister := testplans.List(mockGetter)
	_, err := lister(ctx)
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestGet(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGetter := mocktestplans.NewMockGetter(ctrl)
	mockGetter.
		EXPECT().
		Get(ctx, identity.Id, matchers.OfType(&testplan.TestPlan{})).
		Return(nil)
	getter := testplans.Get(mockGetter)
	_, err := getter(ctx, identity.Id)
	g.Expect(err).ShouldNot(HaveOccurred(), "expected error did not occur")
}

func TestGet_Error(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGetter := mocktestplans.NewMockGetter(ctrl)
	mockGetter.
		EXPECT().
		Get(ctx, identity.Id, matchers.OfType(&testplan.TestPlan{})).
		Return(fmt.Errorf("expected error"))

	getter := testplans.Get(mockGetter)
	_, err := getter(ctx, identity.Id)
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestDelete(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockDeleter := mocktestplans.NewMockDeleter(ctrl)
	mockDeleter.
		EXPECT().
		Delete(ctx, identity.Id).
		Return(nil)
	deleter := testplans.Delete(mockDeleter)
	err := deleter(ctx, identity.Id)
	g.Expect(err).ShouldNot(HaveOccurred(), "expected error did not occur")
}

func TestDelete_Error(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockDeleter := mocktestplans.NewMockDeleter(ctrl)
	mockDeleter.
		EXPECT().
		Delete(ctx, identity.Id).
		Return(fmt.Errorf("returned error"))
	deleter := testplans.Delete(mockDeleter)
	err := deleter(ctx, identity.Id)
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestUpdate(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mocktestplans.NewMockReaderUpdater(ctrl)
	mockReaderUpdater.
		EXPECT().
		Get(ctx, identity.Id, matchers.OfType(&testplan.TestPlan{})).
		Do(func(ctx context.Context, id string, tp *testplan.TestPlan) {
			tp.Identity = &identity
		})
	mockReaderUpdater.
		EXPECT().
		Update(ctx, identity.Id, matchers.OfType(&testplan.TestPlan{}))
	mockMetaHandler := mocktestplans.NewMockMetaHandler(ctrl)
	mockMetaHandler.EXPECT().UpdateMeta("tester", matchers.OfType(&metadata.Identity{}))
	updater := testplans.Update(mockMetaHandler, mockReaderUpdater, goodGetProject)
	createdTestplan, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testTestPlan))
	g.Expect(err).ShouldNot(HaveOccurred(), "unexpected error occurred")
	expectedTestPlan := &testplan.TestPlan{
		Identity:  &identity,
		Name:      testTestPlan.Name,
		ProjectId: testTestPlan.ProjectId,
	}
	g.Expect(createdTestplan).To(Equal(expectedTestPlan), "testplans did not match")
}

func TestUpdate_ValidationError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mocktestplans.NewMockReaderUpdater(ctrl)
	mockMetaHandler := mocktestplans.NewMockMetaHandler(ctrl)
	updater := testplans.Update(mockMetaHandler, mockReaderUpdater, goodGetProject)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testplan.TestPlan{}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestUpdate_InvalidProject(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mocktestplans.NewMockReaderUpdater(ctrl)
	mockMetaHandler := mocktestplans.NewMockMetaHandler(ctrl)
	updater := testplans.Update(mockMetaHandler, mockReaderUpdater, noProject)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testTestPlan))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
	g.Expect(errors.IsValidationError(err)).To(BeTrue(), "project not found error is not a validation error")
}

func TestUpdate_ProjectError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mocktestplans.NewMockReaderUpdater(ctrl)
	mockMetaHandler := mocktestplans.NewMockMetaHandler(ctrl)
	updater := testplans.Update(mockMetaHandler, mockReaderUpdater, errorGetProject)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testTestPlan))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
	g.Expect(errors.IsValidationError(err)).To(BeFalse(), "project not found error is not a validation error")
}

func TestUpdate_GetError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mocktestplans.NewMockReaderUpdater(ctrl)
	mockReaderUpdater.
		EXPECT().
		Get(ctx, identity.Id, matchers.OfType(&testplan.TestPlan{})).
		Return(fmt.Errorf("error during get"))
	mockMetaHandler := mocktestplans.NewMockMetaHandler(ctrl)
	updater := testplans.Update(mockMetaHandler, mockReaderUpdater, goodGetProject)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testTestPlan))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
}

func TestUpdate_UpdateError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mocktestplans.NewMockReaderUpdater(ctrl)
	mockReaderUpdater.
		EXPECT().
		Get(ctx, identity.Id, matchers.OfType(&testplan.TestPlan{})).
		Do(func(ctx context.Context, id string, tp *testplan.TestPlan) {
			tp.Identity = &identity
		})
	mockReaderUpdater.
		EXPECT().
		Update(ctx, identity.Id, matchers.OfType(&testplan.TestPlan{})).
		Return(fmt.Errorf("update error"))
	mockMetaHandler := mocktestplans.NewMockMetaHandler(ctrl)
	mockMetaHandler.EXPECT().UpdateMeta("tester", matchers.OfType(&metadata.Identity{}))
	updater := testplans.Update(mockMetaHandler, mockReaderUpdater, goodGetProject)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testTestPlan))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
}
