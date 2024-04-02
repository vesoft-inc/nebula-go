package nebula_go

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtil_indexOf(t *testing.T) {
	collection := []string{"a", "b", "c"}
	assert.Equal(t, indexOf(collection, "a"), 0)
	assert.Equal(t, indexOf(collection, "b"), 1)
	assert.Equal(t, indexOf(collection, "c"), 2)
	assert.Equal(t, indexOf(collection, "d"), -1)
}
