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
	"github.com/curious-kitten/scratch-post/pkg/metadata"
	"github.com/curious-kitten/scratch-post/pkg/testplans"
	mocktestplans "github.com/curious-kitten/scratch-post/pkg/testplans/mocks"
)

var (
	identity = metadata.Identity{
		ID:           "aabbccddee",
		Type:         "testplan",
		Version:      1,
		CreatedBy:    "author",
		UpdatedBy:    "author",
		CreationTime: time.Now(),
		UpdateTime:   time.Now(),
	}

	testTestPlan = &testplans.TestPlan{
		Name:      "test testplan",
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

func TestTestPlan_AddIdentity(t *testing.T) {
	g := NewWithT(t)
	s := testplans.TestPlan{}
	s.AddIdentity(&identity)
	g.Expect(s.GetIdentity()).To(Equal(&identity))
}

func TestTestPlan_Validate(t *testing.T) {
	g := NewWithT(t)
	s := &testplans.TestPlan{}
	err := s.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with empty testplan")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "empty testplan error is not a validation error")
	s.Name = "Test Name"
	err = s.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with testplan that only has a name")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "testplan with only name error is not a validation error")
	s.ProjectID = "aabbccdd"
	err = s.Validate()
	g.Expect(err).ShouldNot(HaveOccurred(), "error occurred when minimun requirements have been met")
}

func TestNew_Create(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mocktestplans.NewMockIdentityGenerator(ctrl)
	mockGenerator.
		EXPECT().
		AddMeta("tester", "testplan", matchers.OfType(&testplans.TestPlan{})).
		Return(nil).
		Do(func(author string, objType string, identifiable metadata.Identifiable) {
			identifiable.AddIdentity(&identity)
		})
	mockAdder := mocktestplans.NewMockAdder(ctrl)
	mockAdder.
		EXPECT().
		AddOne(ctx, matchers.OfType(&testplans.TestPlan{})).
		Return(nil)

	creator := testplans.New(mockGenerator, mockAdder, goodGetProject)
	testplan, err := creator(ctx, "tester", transformers.ToReadCloser(testTestPlan))
	g.Expect(err).ShouldNot(HaveOccurred(), "unexpected error occurred")
	expectedTestPlan := &testplans.TestPlan{
		Identity:  &identity,
		Name:      testTestPlan.Name,
		ProjectID: testTestPlan.ProjectID,
	}
	g.Expect(testplan).To(Equal(expectedTestPlan), "testplans did not match")
}

func TestNew_ProjectNotFound(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mocktestplans.NewMockIdentityGenerator(ctrl)
	mockAdder := mocktestplans.NewMockAdder(ctrl)
	creator := testplans.New(mockGenerator, mockAdder, noProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testTestPlan))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "project not found error is not a validation error")
}

func TestNew_ProjectError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mocktestplans.NewMockIdentityGenerator(ctrl)
	mockAdder := mocktestplans.NewMockAdder(ctrl)
	creator := testplans.New(mockGenerator, mockAdder, errorGetProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testTestPlan))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeFalse(), "error type was missing")
}

func TestNew_MarshallError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mocktestplans.NewMockIdentityGenerator(ctrl)
	mockAdder := mocktestplans.NewMockAdder(ctrl)
	creator := testplans.New(mockGenerator, mockAdder, errorGetProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(struct{ SomeField string }{SomeField: "test"}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestNew_ValidationError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mocktestplans.NewMockIdentityGenerator(ctrl)
	mockAdder := mocktestplans.NewMockAdder(ctrl)
	creator := testplans.New(mockGenerator, mockAdder, errorGetProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(&testplans.TestPlan{}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestNew_AddMetaError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mocktestplans.NewMockIdentityGenerator(ctrl)
	mockGenerator.
		EXPECT().
		AddMeta("tester", "testplan", matchers.OfType(&testplans.TestPlan{})).
		Return(fmt.Errorf("identity error"))
	mockAdder := mocktestplans.NewMockAdder(ctrl)
	creator := testplans.New(mockGenerator, mockAdder, goodGetProject)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testTestPlan))
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestNew_AddToCollectionError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mocktestplans.NewMockIdentityGenerator(ctrl)
	mockGenerator.
		EXPECT().
		AddMeta("tester", "testplan", matchers.OfType(&testplans.TestPlan{})).
		Return(nil).
		Do(func(author string, objType string, identifiable metadata.Identifiable) error {
			identifiable.AddIdentity(&identity)
			return nil
		})
	mockAdder := mocktestplans.NewMockAdder(ctrl)
	mockAdder.
		EXPECT().
		AddOne(ctx, matchers.OfType(&testplans.TestPlan{})).
		Return(fmt.Errorf("expected error"))

	creator := testplans.New(mockGenerator, mockAdder, goodGetProject)
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
		GetAll(ctx, matchers.OfType(&[]testplans.TestPlan{})).
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
		GetAll(ctx, matchers.OfType(&[]testplans.TestPlan{})).
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
		Get(ctx, identity.ID, matchers.OfType(&testplans.TestPlan{})).
		Return(nil)
	getter := testplans.Get(mockGetter)
	_, err := getter(ctx, identity.ID)
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
		Get(ctx, identity.ID, matchers.OfType(&testplans.TestPlan{})).
		Return(fmt.Errorf("expected error"))

	getter := testplans.Get(mockGetter)
	_, err := getter(ctx, identity.ID)
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
		Delete(ctx, identity.ID).
		Return(nil)
	deleter := testplans.Delete(mockDeleter)
	err := deleter(ctx, identity.ID)
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
		Delete(ctx, identity.ID).
		Return(fmt.Errorf("returned error"))
	deleter := testplans.Delete(mockDeleter)
	err := deleter(ctx, identity.ID)
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
		Get(ctx, identity.ID, matchers.OfType(&testplans.TestPlan{})).
		Do(func (ctx context.Context, id string, tp *testplans.TestPlan){
			tp.Identity = &identity
		})
	mockReaderUpdater.
		EXPECT().
		Update(ctx, identity.ID, matchers.OfType(&testplans.TestPlan{}))
	updater := testplans.Update(mockReaderUpdater, goodGetProject)
	testplan, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testTestPlan))
	g.Expect(err).ShouldNot(HaveOccurred(), "unexpected error occurred")
	expectedTestPlan := &testplans.TestPlan{
		Identity:  &identity,
		Name:      testTestPlan.Name,
		ProjectID: testTestPlan.ProjectID,
	}
	g.Expect(testplan).To(Equal(expectedTestPlan), "testplans did not match")
}

func TestUpdate_ValidationError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mocktestplans.NewMockReaderUpdater(ctrl)
	updater := testplans.Update(mockReaderUpdater, goodGetProject)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testplans.TestPlan{}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestUpdate_InvalidProject(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mocktestplans.NewMockReaderUpdater(ctrl)
	updater := testplans.Update(mockReaderUpdater, noProject)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testTestPlan))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "project not found error is not a validation error")
	
}

func TestUpdate_ProjectError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mocktestplans.NewMockReaderUpdater(ctrl)
	updater := testplans.Update(mockReaderUpdater, errorGetProject)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testTestPlan))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
	g.Expect(metadata.IsValidationError(err)).To(BeFalse(), "project not found error is not a validation error")
	
}

func TestUpdate_GetError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mocktestplans.NewMockReaderUpdater(ctrl)
	mockReaderUpdater.
		EXPECT().
		Get(ctx, identity.ID, matchers.OfType(&testplans.TestPlan{})).
		Return(fmt.Errorf("error during get"))
	updater := testplans.Update(mockReaderUpdater, goodGetProject)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testTestPlan))
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
		Get(ctx, identity.ID, matchers.OfType(&testplans.TestPlan{})).
		Do(func (ctx context.Context, id string, tp *testplans.TestPlan){
			tp.Identity = &identity
		})
	mockReaderUpdater.
		EXPECT().
		Update(ctx, identity.ID, matchers.OfType(&testplans.TestPlan{})).
		Return(fmt.Errorf("update error"))
	updater := testplans.Update(mockReaderUpdater, goodGetProject)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testTestPlan))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
}
