package projects_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/curious-kitten/scratch-post/internal/test/matchers"
	"github.com/curious-kitten/scratch-post/internal/test/transformers"
	"github.com/curious-kitten/scratch-post/pkg/metadata"
	"github.com/curious-kitten/scratch-post/pkg/projects"
	mockProjects "github.com/curious-kitten/scratch-post/pkg/projects/mocks"
)

var (
	identity = metadata.Identity{
		ID:           "aabbccddee",
		Type:         "project",
		Version:      1,
		CreatedBy:    "author",
		UpdatedBy:    "author",
		CreationTime: time.Now(),
		UpdateTime:   time.Now(),
	}

	testProject = &projects.Project{
		Name: "test project",
	}
)

func TestProject_AddIdentity(t *testing.T) {
	g := NewWithT(t)
	s := projects.Project{}
	s.AddIdentity(&identity)
	g.Expect(s.GetIdentity()).To(Equal(&identity))
}

func TestScenario_Validate(t *testing.T) {
	g := NewWithT(t)
	p := &projects.Project{}
	err := p.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with empty scenario")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "empty scenario error is not a validation error")
	p.Name = "Test Name"
	err = p.Validate()
	g.Expect(err).ShouldNot(HaveOccurred(), "error occurred when minimun requirements have been met")
}

func TestNew_Create(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockProjects.NewMockIdentityGenerator(ctrl)
	mockGenerator.
		EXPECT().
		AddMeta("tester", "project", matchers.OfType(&projects.Project{})).
		Return(nil).
		Do(func(author string, objType string, identifiable metadata.Identifiable) {
			identifiable.AddIdentity(&identity)
		})
	mockAdder := mockProjects.NewMockAdder(ctrl)
	mockAdder.
		EXPECT().
		AddOne(ctx, matchers.OfType(&projects.Project{})).
		Return(nil)

	creator := projects.New(mockGenerator, mockAdder)
	scenario, err := creator(ctx, "tester", transformers.ToReadCloser(testProject))
	g.Expect(err).ShouldNot(HaveOccurred(), "unexpected error occurred")
	expectedProject := &projects.Project{
		Identity: &identity,
		Name:     testProject.Name,
	}
	g.Expect(scenario).To(Equal(expectedProject), "projects did not match")
}

func TestNew_MarshallError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockProjects.NewMockIdentityGenerator(ctrl)
	mockAdder := mockProjects.NewMockAdder(ctrl)
	creator := projects.New(mockGenerator, mockAdder)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(struct{ SomeField string }{SomeField: "test"}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestNew_ValidationError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockProjects.NewMockIdentityGenerator(ctrl)
	mockAdder := mockProjects.NewMockAdder(ctrl)
	creator := projects.New(mockGenerator, mockAdder)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(&projects.Project{}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestNew_AddMetaError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockProjects.NewMockIdentityGenerator(ctrl)
	mockGenerator.
		EXPECT().
		AddMeta("tester", "project", matchers.OfType(&projects.Project{})).
		Return(fmt.Errorf("identity error"))
	mockAdder := mockProjects.NewMockAdder(ctrl)
	creator := projects.New(mockGenerator, mockAdder)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testProject))
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestNew_AddToCollectionError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockProjects.NewMockIdentityGenerator(ctrl)
	mockGenerator.
		EXPECT().
		AddMeta("tester", "project", matchers.OfType(&projects.Project{})).
		Return(nil).
		Do(func(author string, objType string, identifiable metadata.Identifiable) error {
			identifiable.AddIdentity(&identity)
			return nil
		})
	mockAdder := mockProjects.NewMockAdder(ctrl)
	mockAdder.
		EXPECT().
		AddOne(ctx, matchers.OfType(&projects.Project{})).
		Return(fmt.Errorf("expected error"))

	creator := projects.New(mockGenerator, mockAdder)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(testProject))
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestList(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGetter := mockProjects.NewMockGetter(ctrl)
	mockGetter.
		EXPECT().
		GetAll(ctx, matchers.OfType(&[]projects.Project{})).
		Return(nil)

	lister := projects.List(mockGetter)
	_, err := lister(ctx)
	g.Expect(err).ShouldNot(HaveOccurred(), "expected error did not occur")
}

func TestList_RetrieveError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGetter := mockProjects.NewMockGetter(ctrl)
	mockGetter.
		EXPECT().
		GetAll(ctx, matchers.OfType(&[]projects.Project{})).
		Return(fmt.Errorf("expected error"))

	lister := projects.List(mockGetter)
	_, err := lister(ctx)
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestGet(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGetter := mockProjects.NewMockGetter(ctrl)
	mockGetter.
		EXPECT().
		Get(ctx, identity.ID, matchers.OfType(&projects.Project{})).
		Return(nil)
	getter := projects.Get(mockGetter)
	_, err := getter(ctx, identity.ID)
	g.Expect(err).ShouldNot(HaveOccurred(), "expected error did not occur")
}

func TestGet_Error(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGetter := mockProjects.NewMockGetter(ctrl)
	mockGetter.
		EXPECT().
		Get(ctx, identity.ID, matchers.OfType(&projects.Project{})).
		Return(fmt.Errorf("expected error"))

	getter := projects.Get(mockGetter)
	_, err := getter(ctx, identity.ID)
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestDelete(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockDeleter := mockProjects.NewMockDeleter(ctrl)
	mockDeleter.
		EXPECT().
		Delete(ctx, identity.ID).
		Return(nil)
	deleter := projects.Delete(mockDeleter)
	err := deleter(ctx, identity.ID)
	g.Expect(err).ShouldNot(HaveOccurred(), "expected error did not occur")
}

func TestDelete_Error(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockDeleter := mockProjects.NewMockDeleter(ctrl)
	mockDeleter.
		EXPECT().
		Delete(ctx, identity.ID).
		Return(fmt.Errorf("returned error"))
	deleter := projects.Delete(mockDeleter)
	err := deleter(ctx, identity.ID)
	g.Expect(err).Should(HaveOccurred(), "expected error did not occur")
}

func TestUpdate(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockProjects.NewMockReaderUpdater(ctrl)
	mockReaderUpdater.
		EXPECT().
		Get(ctx, identity.ID, matchers.OfType(&projects.Project{})).
		Do(func (ctx context.Context, id string, tp *projects.Project){
			tp.Identity = &identity
		})
	mockReaderUpdater.
		EXPECT().
		Update(ctx, identity.ID, matchers.OfType(&projects.Project{}))
	updater := projects.Update(mockReaderUpdater)
	project, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testProject))
	g.Expect(err).ShouldNot(HaveOccurred(), "unexpected error occurred")
	expectedProject := &projects.Project{
		Identity:  &identity,
		Name:      testProject.Name,
	}
	g.Expect(project).To(Equal(expectedProject), "projects did not match")
}

func TestUpdate_ValidationError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockProjects.NewMockReaderUpdater(ctrl)
	updater := projects.Update(mockReaderUpdater)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(projects.Project{}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(metadata.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestUpdate_GetError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockProjects.NewMockReaderUpdater(ctrl)
	mockReaderUpdater.
		EXPECT().
		Get(ctx, identity.ID, matchers.OfType(&projects.Project{})).
		Return(fmt.Errorf("error during get"))
	updater := projects.Update(mockReaderUpdater)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testProject))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
}

func TestUpdate_UpdateError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockProjects.NewMockReaderUpdater(ctrl)
	mockReaderUpdater.
		EXPECT().
		Get(ctx, identity.ID, matchers.OfType(&projects.Project{})).
		Do(func (ctx context.Context, id string, tp *projects.Project){
			tp.Identity = &identity
		})
	mockReaderUpdater.
		EXPECT().
		Update(ctx, identity.ID, matchers.OfType(&projects.Project{})).
		Return(fmt.Errorf("update error"))
	updater := projects.Update(mockReaderUpdater)
	_, err := updater(ctx, "tester", identity.ID, transformers.ToReadCloser(testProject))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
}
