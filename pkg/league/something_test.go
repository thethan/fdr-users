package league

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSomethingIsTrue(t *testing.T) {
	t.Run("get something", func(t *testing.T) {
		var sandwhich bool
		assert.True(t, sandwhich)
	})
}
