package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEndpointCmd(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewEndpointCmd(v, fs)

	assert.Equal(t, "endpoint", cmd.Name(), "NewEndpointCmd() should return command named \"endpoint\". but: %q", cmd.Name())
}

func TestNewEndpointCmd_Flag(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewEndpointCmd(v, fs)
	formatFlag := cmd.Flags().Lookup("format")

	assert.True(t, cmd.HasAvailableFlags(), "genconf command should have available flag")
	assert.NotNil(t, formatFlag, "genconf command should have \"format\" flag")
	assert.Equal(t, "string", formatFlag.Value.Type(), "\"format\" flag is string")
}

func TestRunEndpoint(t *testing.T) {
	v := viper.New()
	fs := afero.NewOsFs()
	stdout := new(bytes.Buffer)
	cmd := NewRootCmd(v, fs)
	cmd.SetOut(stdout)
	cmd.SetArgs([]string{"endpoint", "--format", "table", "../internal/genconf/testdata/src/echo_simple"})
	err := cmd.Execute()

	require.NoError(t, err)
	assert.Equal(t, "matching_groups:\n    - ^/api/users/([^/]+)$\n    - ^/api/users$\n    - ^/api/items$\n\n", stdout.String())
}
