package genconf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckImportedFramework_echo(t *testing.T) {
	filename := "testdata/echo_simple.go"
	fw, err := CheckImportedFramework(filename)

	assert.NoError(t, err)
	assert.Equal(t, fw.Kind, EchoV4)
	assert.Equal(t, fw.PkgName, "echo")
}

func TestCheckImportedFramework_echo_complex(t *testing.T) {
	filename := "testdata/echo_complex.go"
	fw, err := CheckImportedFramework(filename)

	assert.NoError(t, err)
	assert.Equal(t, fw.Kind, EchoV4)
	assert.Equal(t, fw.PkgName, "echov4")
}
