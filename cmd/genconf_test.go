package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenConfCmd(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewGenConfCmd(v, fs)

	assert.Equal(t, "genconf", cmd.Name(), "NewGenConfCmd() should return command named \"genconf\". but: %q", cmd.Name())
}

func TestNewGenConfCmd_Flag(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewGenConfCmd(v, fs)
	formatFlag := cmd.Flags().Lookup("format")

	assert.True(t, cmd.HasAvailableFlags(), "genconf command should have available flag")
	assert.NotNil(t, formatFlag, "genconf command should have \"format\" flag")
	assert.Equal(t, "string", formatFlag.Value.Type(), "\"format\" flag is string")
}

func Test_printMatchingGroupInToml(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewGenConfCmd(v, fs)

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	err := printMatchingGroupInToml(cmd, MatchingGroupConf{MatchingGroups: []string{"foo", "bar"}})

	assert.NoError(t, err)
	assert.Equal(t, "matching_groups = [\n  'foo',\n  'bar'\n]\n\n", stdout.String())
}

func Test_printMatchingGroupInYaml(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewGenConfCmd(v, fs)

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	err := printMatchingGroupInYaml(cmd, MatchingGroupConf{MatchingGroups: []string{"foo", "bar"}})

	assert.NoError(t, err)
	assert.Equal(t, "matching_groups:\n    - foo\n    - bar\n\n", stdout.String())
}

func Test_printMatchingGroupInJson(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewGenConfCmd(v, fs)

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	err := printMatchingGroupInJson(cmd, MatchingGroupConf{MatchingGroups: []string{"foo", "bar"}})

	assert.NoError(t, err)
	assert.Equal(t, "{\n  \"matching_groups\": [\n    \"foo\",\n    \"bar\"\n  ]\n}\n", stdout.String())
}

func TestRunGenConf(t *testing.T) {
	v := viper.New()
	fs := afero.NewOsFs()
	stdout := new(bytes.Buffer)
	cmd := NewRootCmd(v, fs)
	cmd.SetOut(stdout)
	cmd.SetArgs([]string{"genconf", "--format", "yaml", "../internal/genconf/testdata/src/echo_simple"})
	err := cmd.Execute()

	require.NoError(t, err)
	assert.Equal(t, "matching_groups:\n    - ^/api/users/([^/]+)$\n    - ^/api/users$\n    - ^/api/items$\n\n", stdout.String())
}
