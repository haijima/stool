package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/haijima/stool/internal"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewParamCmd(t *testing.T) {
	p := internal.NewParamProfiler()
	v, fs := createViperAndFs()
	cmd := NewParamCmd(p, v, fs)

	assert.Equal(t, "param", cmd.Name(), "NewParamCmd() should return command named \"param\". but: %q", cmd.Name())
}

func TestNewParamCmd_Flag(t *testing.T) {
	p := internal.NewParamProfiler()
	v, fs := createViperAndFs()
	cmd := NewParamCmd(p, v, fs)
	numFlag := cmd.Flags().Lookup("num")
	statFlag := cmd.Flags().Lookup("stat")
	typeFlag := cmd.Flags().Lookup("type")

	assert.NotNil(t, numFlag, "param command should have \"num\" flag")
	assert.Equal(t, "int", numFlag.Value.Type(), "\"num\" flag is int")
	assert.NotNil(t, statFlag, "param command should have \"stat\" flag")
	assert.Equal(t, "bool", statFlag.Value.Type(), "\"stat\" flag is bool")
	assert.True(t, cmd.HasAvailableFlags(), "param command should have available flag")
	assert.NotNil(t, typeFlag, "param command should have \"type\" flag")
	assert.Equal(t, "t", typeFlag.Shorthand, "\"type\" flag's shorthand is \"t\"")
	assert.Equal(t, "string", typeFlag.Value.Type(), "\"type\" flag is string")
}

func TestNewParamCmd_RunE(t *testing.T) {
	p := internal.NewParamProfiler()
	v := viper.New()
	fs := afero.NewOsFs()
	cmd := NewParamCmd(p, v, fs)

	dir, _ := os.Getwd()
	fileName := dir + "/testdata/access.log"
	v.Set("file", fileName)
	v.Set("matching_groups", []string{"^/api/condition/([^/]+)$"})

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	_ = v.BindPFlags(cmd.Flags())
	err := cmd.RunE(cmd, []string{})

	//s := stdout.String()
	//fmt.Println(s)

	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "GET ^/api/condition/([^/]+)$ (Count: 4,489)")
	assert.Contains(t, stdout.String(), "\t?condition_level (Count: 4,441, Rate: 98.93%, Cardinality: 5, Gini: 0.793)\n")
	assert.Contains(t, stdout.String(), "\tQuery key combination (Cardinality: 4, Gini: 0.710)")
	assert.Contains(t, stdout.String(), "\tQuery key value combination (Cardinality: 4,195, Gini: 0.065)")
	assert.Contains(t, stdout.String(), "POST ^/api/condition/([^/]+)$ (Count: 65,993)")
}

func TestNewParamCmdExecute(t *testing.T) {
	v := viper.New()
	fs := afero.NewOsFs()
	cmd := NewRootCmd(v, fs)

	dir, _ := os.Getwd()
	fileName := dir + "/testdata/access.log"
	cmd.SetArgs([]string{"param", "-f", fileName, "--format", "table", "--stat"})

	assert.NoError(t, cmd.Execute())
}
