package users_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aminpaks/go-streams/pkg/tools"
)

func Test_UserModels(t *testing.T) {
	t.Parallel()
	runner, it := tools.New(t)

	runner([]interface{}{
		it("should the value be equal to 1", func(t *testing.T) {
			assert := assert.New(t)

			assert.Equal(1, 2)
		}),

		it("should always fail", func(t *testing.T) {
			assert := assert.New(t)

			assert.True(false)
			assert.Len("check", 1)
		}),

		it("should X has Top equal to 2", func(t *testing.T) {
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
	})
}
