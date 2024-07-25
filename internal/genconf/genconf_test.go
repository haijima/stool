package genconf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckImportedFramework_echo(t *testing.T) {
	fw, err := CheckImportedFramework("testdata/src/echo_simple", "./...")

	assert.NoError(t, err)
	assert.Equal(t, fw, EchoV4)
}

func TestCheckImportedFramework_echo_complex(t *testing.T) {
	fw, err := CheckImportedFramework("testdata/src/echo_complex", "./...")

	assert.NoError(t, err)
	assert.Equal(t, fw, EchoV4)
}

func TestCheckImportedFramework_net_http(t *testing.T) {
	fw, err := CheckImportedFramework("testdata/src/net_http", "./...")

	assert.NoError(t, err)
	assert.Equal(t, fw, NetHttp)
}
