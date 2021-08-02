package tags

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetTagsKey(t *testing.T) {
	assert.Equal(t, "tag1=a,tag2=b,", GetTagsKey(map[string]string{"tag1": "a", "tag2": "b"}))
}
