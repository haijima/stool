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

func TestRunGenConf(t *testing.T) {
	v := viper.New()
	fs := afero.NewOsFs()
	stdout := new(bytes.Buffer)
	cmd := NewRootCmd(v, fs)
	cmd.SetOut(stdout)
	cmd.SetArgs([]string{"genconf", "--format", "yaml", "--dir", "./testdata/src"})
	err := cmd.Execute()

	require.NoError(t, err)
	assert.Equal(t, "config: \"\"\nendpoint:\n    format: table\n    pattern: ./...\nfile: \"\"\nfilter: \"\"\ngenconf:\n    capture-group-name: \"false\"\n    dir: ./testdata/src\n    format: yaml\n    pattern: ./...\nlog-labels: '[]'\nmatching-groups:\n    - ^/api/users/([^/]+)$\n    - ^/api/users$\n    - ^/api/items$\nno-color: \"false\"\nparam:\n    format: table\n    num: \"5\"\n    stat: \"false\"\n    type: all\nquiet: \"false\"\nscenario:\n    format: dot\n    palette: \"false\"\ntime-format: 02/Jan/2006:15:04:05 -0700\ntransition:\n    format: dot\ntrend:\n    format: table\n    interval: \"5\"\n    sort: '[sum:desc]'\nverbose: \"0\"\n", stdout.String())
}
