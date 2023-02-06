package internal

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_key(t *testing.T) {
	mg := "^/api/user/[^\\/]+$"
	pattern, err := regexp.Compile(mg)
	assert.Nil(t, err)

	k := key("GET /api/user/xyz?key=v&foo=bar", []*regexp.Regexp{pattern})

	assert.Equal(t, "GET ^/api/user/[^\\/]+$", k)
}
