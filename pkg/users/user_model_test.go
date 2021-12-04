package users_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aminpaks/go-streams/pkg/testrun"
)

func Test_UserModels(t *testing.T) {
	t.Parallel()

	test := testrun.New(t)

	test.Run(
		test.It("should the value be equal to 1", func(t *testing.T) {
			assert := assert.New(t)

			assert.Equal(1, 1)
		}),

		test.XIt("should always fail if not skipped", func(t *testing.T) {
			assert := assert.New(t)

			assert.True(false)
		}),

		test.XIt("should X has Top equal to 2", func(t *testing.T) {
			assert := assert.New(t)

			x := func() *struct {
				Check int
				Top   string
			} {
				return nil
			}()

			assert.Equal(2, x.Top)
			assert.Equal(1, 2)
		}),
	)
}
