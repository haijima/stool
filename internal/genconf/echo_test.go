package genconf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenMatchingGroupFromEchoV4(t *testing.T) {
	ext := &EchoExtractor{}
	mgs, err := GenMatchingGroup("testdata/src/echo_simple", "./...", ext, false)

	require.NoError(t, err)
	require.Equal(t, 3, len(mgs))
	assert.Equal(t, "^/api/users/([^/]+)$", mgs[0])
	assert.Equal(t, "^/api/users$", mgs[1])
	assert.Equal(t, "^/api/items$", mgs[2])
}

func TestGenMatchingGroupFromEchoV4_complex(t *testing.T) {
	ext := &EchoExtractor{}
	mgs, err := GenMatchingGroup("testdata/src/echo_complex", "./...", ext, false)

	require.NoError(t, err)
	require.Equal(t, 8, len(mgs))
	assert.Equal(t, "^/view/screen1$", mgs[0])
	assert.Equal(t, "^/index$", mgs[1])
	assert.Equal(t, "^/auth/logout$", mgs[2])
	assert.Equal(t, "^/auth/login$", mgs[3])
	assert.Equal(t, "^/auth/admin/login$", mgs[4])
	assert.Equal(t, "^/api/groups/([^/]+)/users/([^/]+)/tasks$", mgs[5])
	assert.Equal(t, "^/api/groups/([^/]+)/users$", mgs[6])
	assert.Equal(t, "^/api/groups$", mgs[7])
}
