package nebula_go

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtil_IndexOf(t *testing.T) {
	collection := []string{"a", "b", "c"}
	assert.Equal(t, IndexOf(collection, "a"), 0)
	assert.Equal(t, IndexOf(collection, "b"), 1)
	assert.Equal(t, IndexOf(collection, "c"), 2)
	assert.Equal(t, IndexOf(collection, "d"), -1)
}
