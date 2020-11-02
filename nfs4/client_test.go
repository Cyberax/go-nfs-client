package nfs4

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSplitPath(t *testing.T) {
	path := splitPath("/a/b/c//d")
	assert.True(t, assert.ObjectsAreEqual([]string{"a", "b", "c", "d"}, path))

	path = splitPath("a/b/c")
	assert.True(t, assert.ObjectsAreEqual([]string{"a", "b", "c"}, path))
}
