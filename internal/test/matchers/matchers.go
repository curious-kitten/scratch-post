package matchers

import (
	"fmt"

	"github.com/golang/mock/gomock"
)

type ofType struct {
	t string
}

func (o *ofType) Matches(x interface{}) bool {
	return fmt.Sprintf("%T", x) == o.t
}

func (o *ofType) String() string {
	return o.t
}

// OfType returns a type matcher to be used in test mocking
func OfType(item interface{}) gomock.Matcher {
	return &ofType{
		t: fmt.Sprintf("%T", item),
	}
}
