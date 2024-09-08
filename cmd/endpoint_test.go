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

	assert.True(t, cmd.HasAvailableFlags(), "endpoint command should have available flag")
	assert.NotNil(t, formatFlag, "endpoint command should have \"format\" flag")
	assert.Equal(t, "string", formatFlag.Value.Type(), "\"format\" flag is string")
}

func TestRunEndpoint(t *testing.T) {
	v := viper.New()
	fs := afero.NewOsFs()
	stdout := new(bytes.Buffer)
	cmd := NewRootCmd(v, fs)
	cmd.SetOut(stdout)
	cmd.SetArgs([]string{"endpoint", "--format", "csv", "./testdata/src"})
	err := cmd.Execute()

	require.NoError(t, err)
	assert.Equal(t, "#,Method,Path,Function,Declared Package,Declared Position\n1,POST,/api/users,CreateUser,g/h/s/i/testdata,main.go:18:6\n2,GET,/api/users,GetUsers,g/h/s/i/testdata,main.go:22:6\n3,GET,/api/users/:id,GetUser,g/h/s/i/testdata,main.go:26:6\n4,GET,/api/items,GetItems,g/h/s/i/testdata,main.go:30:6\n", stdout.String())
}
