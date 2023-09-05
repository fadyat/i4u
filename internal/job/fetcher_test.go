package job

import (
	"github.com/fadyat/i4u/api"
	"github.com/fadyat/i4u/internal/entity"
	"testing"
)

type fetcherJobTestcase struct {
	name          string
	in            []entity.Message
	pre           func(t *testing.T, c api.Mail, tc fetcherJobTestcase)
	expectedOut   []entity.Message
	expectedError error
}

func TestMessageFetcherJob_Run(t *testing.T) {
	testCases := []fetcherJobTestcase{
		{},
	}

	for _, tt := range testCases {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// ...
		})
	}
}
