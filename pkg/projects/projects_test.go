package projects_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/curious-kitten/scratch-post/internal/decoder"
	"github.com/curious-kitten/scratch-post/internal/test/matchers"
	"github.com/curious-kitten/scratch-post/internal/test/transformers"
	metadata "github.com/curious-kitten/scratch-post/pkg/api/v1/metadata"
	project "github.com/curious-kitten/scratch-post/pkg/api/v1/project"
	"github.com/curious-kitten/scratch-post/pkg/projects"
	mockProjects "github.com/curious-kitten/scratch-post/pkg/projects/mocks"
)

var (
	sortBy            = ""
	reverse           = false
	count             = 1000
	previousLastValue = ""
	identity          = metadata.Identity{
		Id:           "aabbccddee",
		Type:         "project",
		Version:      1,
		CreatedBy:    "author",
		UpdatedBy:    "author",
		CreationTime: time.Now().Unix(),
		UpdateTime:   time.Now().Unix(),
	}

	testProject = &project.Project{
		Name: "test project",
	}
)

func TestScenario_Validate(t *testing.T) {
	g := NewWithT(t)
	p := &project.Project{}
	err := p.Validate()
	g.Expect(err).Should(HaveOccurred(), "No error with empty scenario")
	g.Expect(decoder.IsValidationError(err)).To(BeTrue(), "empty scenario error is not a validation error")
	p.Name = "Test Name"
	err = p.Validate()
	g.Expect(err).ShouldNot(HaveOccurred(), "error occurred when minimun requirements have been met")
}

func TestNew_Create(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockProjects.NewMockMetaHandler(ctrl)
	mockGenerator.
		EXPECT().
		NewMeta("tester", "project").
		Return(&identity, nil)
	mockAdder := mockProjects.NewMockAdder(ctrl)
	mockAdder.
		EXPECT().
		AddOne(ctx, matchers.OfType(&project.Project{})).
		Return(nil)

	creator := projects.New(mockGenerator, mockAdder)
	scenario, err := creator(ctx, "tester", transformers.ToReadCloser(testProject))
	g.Expect(err).ShouldNot(HaveOccurred(), "unexpected error occurred")
	expectedProject := &project.Project{
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
	mockGenerator := mockProjects.NewMockMetaHandler(ctrl)
	mockAdder := mockProjects.NewMockAdder(ctrl)
	creator := projects.New(mockGenerator, mockAdder)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(struct{ SomeField string }{SomeField: "test"}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(decoder.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestNew_ValidationError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockProjects.NewMockMetaHandler(ctrl)
	mockAdder := mockProjects.NewMockAdder(ctrl)
	creator := projects.New(mockGenerator, mockAdder)
	_, err := creator(ctx, "tester", transformers.ToReadCloser(&project.Project{}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(decoder.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestNew_AddMetaError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockGenerator := mockProjects.NewMockMetaHandler(ctrl)
	mockGenerator.
		EXPECT().
		NewMeta("tester", "project").
		Return(nil, fmt.Errorf("identity error"))
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
	mockGenerator := mockProjects.NewMockMetaHandler(ctrl)
	mockGenerator.
		EXPECT().
		NewMeta("tester", "project").
		Return(&identity, nil)
	mockAdder := mockProjects.NewMockAdder(ctrl)
	mockAdder.
		EXPECT().
		AddOne(ctx, matchers.OfType(&project.Project{})).
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
		GetAll(ctx, matchers.OfType(&[]project.Project{}), map[string][]string{}, sortBy, reverse, count, previousLastValue).
		Return(nil)

	lister := projects.List(mockGetter)
	_, err := lister(ctx, map[string][]string{}, sortBy, reverse, count, previousLastValue)
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
		GetAll(ctx, matchers.OfType(&[]project.Project{}), map[string][]string{}, sortBy, reverse, count, previousLastValue).
		Return(fmt.Errorf("expected error"))

	lister := projects.List(mockGetter)
	_, err := lister(ctx, map[string][]string{}, sortBy, reverse, count, previousLastValue)
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
		Get(ctx, identity.Id, matchers.OfType(&project.Project{})).
		Return(nil)
	getter := projects.Get(mockGetter)
	_, err := getter(ctx, identity.Id)
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
		Get(ctx, identity.Id, matchers.OfType(&project.Project{})).
		Return(fmt.Errorf("expected error"))

	getter := projects.Get(mockGetter)
	_, err := getter(ctx, identity.Id)
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
		Delete(ctx, identity.Id).
		Return(nil)
	deleter := projects.Delete(mockDeleter)
	err := deleter(ctx, identity.Id)
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
		Delete(ctx, identity.Id).
		Return(fmt.Errorf("returned error"))
	deleter := projects.Delete(mockDeleter)
	err := deleter(ctx, identity.Id)
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
		Get(ctx, identity.Id, matchers.OfType(&project.Project{})).
		Do(func(ctx context.Context, id string, tp *project.Project) {
			tp.Identity = &identity
		})
	mockReaderUpdater.
		EXPECT().
		Update(ctx, identity.Id, matchers.OfType(&project.Project{}))
	mockMetaHandler := mockProjects.NewMockMetaHandler(ctrl)
	mockMetaHandler.EXPECT().UpdateMeta("tester", matchers.OfType(&metadata.Identity{}))
	updater := projects.Update(mockMetaHandler, mockReaderUpdater)
	createdProject, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testProject))
	g.Expect(err).ShouldNot(HaveOccurred(), "unexpected error occurred")
	expectedProject := &project.Project{
		Identity: &identity,
		Name:     testProject.Name,
	}
	g.Expect(createdProject).To(Equal(expectedProject), "projects did not match")
}

func TestUpdate_ValidationError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockProjects.NewMockReaderUpdater(ctrl)
	mockMetaHandler := mockProjects.NewMockMetaHandler(ctrl)
	updater := projects.Update(mockMetaHandler, mockReaderUpdater)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(project.Project{}))
	g.Expect(err).Should(HaveOccurred(), "error did not occur")
	g.Expect(decoder.IsValidationError(err)).To(BeTrue(), "invalid item passed does not return a validation error")
}

func TestUpdate_GetError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	mockReaderUpdater := mockProjects.NewMockReaderUpdater(ctrl)
	mockReaderUpdater.
		EXPECT().
		Get(ctx, identity.Id, matchers.OfType(&project.Project{})).
		Return(fmt.Errorf("error during get"))
	mockMetaHandler := mockProjects.NewMockMetaHandler(ctrl)
	updater := projects.Update(mockMetaHandler, mockReaderUpdater)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testProject))
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
		Get(ctx, identity.Id, matchers.OfType(&project.Project{})).
		Do(func(ctx context.Context, id string, tp *project.Project) {
			tp.Identity = &identity
		})
	mockReaderUpdater.
		EXPECT().
		Update(ctx, identity.Id, matchers.OfType(&project.Project{})).
		Return(fmt.Errorf("update error"))
	mockMetaHandler := mockProjects.NewMockMetaHandler(ctrl)
	mockMetaHandler.EXPECT().UpdateMeta("tester", matchers.OfType(&metadata.Identity{}))
	updater := projects.Update(mockMetaHandler, mockReaderUpdater)
	_, err := updater(ctx, "tester", identity.Id, transformers.ToReadCloser(testProject))
	g.Expect(err).Should(HaveOccurred(), "unexpected error occurred")
}
