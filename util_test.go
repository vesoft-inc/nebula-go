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

func TestUtil_parseTTL(t *testing.T) {
	s := "CREATE TAG `user` (\n\t\t`name` string NOT NULL,\n\t\t`created_at` int64 NULL\n\t) ttl_duration = 5, ttl_col = \"created_at\""
	col, duration, err := parseTTL(s)
	assert.Nil(t, err)
	assert.Equal(t, col, "created_at")
	assert.Equal(t, duration, uint(5))

	s = "CREATE TAG `user` (\n\t\t`name` string NOT NULL,\n\t\t`created_at` int64 NULL\n\t) ttl_duration = 0, ttl_col = \"\""
	col, duration, err = parseTTL(s)
	assert.Nil(t, err)
	assert.Equal(t, col, "")
	assert.Equal(t, duration, uint(0))
}
