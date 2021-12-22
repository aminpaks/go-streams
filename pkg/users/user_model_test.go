package users_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aminpaks/go-streams/pkg/testrun"
)

func Test_UserModels(t *testing.T) {
	t.Parallel()

	r := testrun.New(t)

	r.Run(
		r.It("should the value be equal to 1", func(t *testing.T) {
			assert := assert.New(t)

			assert.Equal(1, 1)
		}),

		r.XIt("should always fail if not skipped", func(t *testing.T) {
			assert := assert.New(t)

			assert.True(false)
		}),

		r.XIt("should X has Top equal to 2", func(t *testing.T) {
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
