package scenarios_test

import (
	"testing"
	"time"

	// "github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/curious-kitten/scratch-post/pkg/metadata"
	"github.com/curious-kitten/scratch-post/pkg/scenarios"
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
)

func TestScenario_AddIdentity(t *testing.T) {
	g := NewWithT(t)
	s := scenarios.Scenario{}
	s.AddIdentity(&identity)
	g.Expect(s.GetIdentity()).To(Equal(&identity))
}
