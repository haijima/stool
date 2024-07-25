package genconf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenMatchingGroupFromNetHttp(t *testing.T) {
	ext := &NetHttpExtractor{}
	mgs, err := GenMatchingGroup("testdata/src/net_http", "./...", ext, false)

	require.NoError(t, err)
	require.Equal(t, 7, len(mgs))
	assert.Equal(t, "^/foo/bar$", mgs[0])
	assert.Equal(t, "^/foo/([^/]+)/bar/([^/]+)$", mgs[1])
	assert.Equal(t, "^/foo/([^/]+)$", mgs[2])
	assert.Equal(t, "^/foo/(.+)$", mgs[3])
	assert.Equal(t, "^/foo/(.*)$", mgs[4])
	assert.Equal(t, "^/foo/$", mgs[5])
	assert.Equal(t, "^/foo$", mgs[6])
}
