package cmd

import (
	"bytes"
	"flag"
	"path/filepath"
	"testing"

	"github.com/haijima/cobrax"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/tenntenn/golden"
)

var flagUpdateGolden bool

func init() {
	flag.BoolVar(&flagUpdateGolden, "update", false, "update golden files")
}

func TestNewRootCmd(t *testing.T) {
	fs := afero.NewMemMapFs()
	v := viper.New()
	cmd := NewRootCmd(v, fs)
	_ = cmd.BindFlags()

	assert.Equal(t, "stool", cmd.Name(), "NewRootCommand() should return command named \"stool\". but: \"%s\"", cmd.Name())
	assert.False(t, cmd.HasParent(), "RootCommand should not have parent command.")
	assert.Equal(t, 6, len(cmd.Commands()), "RootCommand should have 1 sub command. but: %d", len(cmd.Commands()))
	assert.False(t, cmd.Runnable(), "RootCommand should not runnable.")
}

func TestNewRootCmd_Flag(t *testing.T) {
	fs := afero.NewMemMapFs()
	v := viper.New()
	cmd := NewRootCmd(v, fs)
	_ = cmd.BindFlags()
	fileFlag := cmd.PersistentFlags().Lookup("file")
	matchingGroupsFlag := cmd.PersistentFlags().Lookup("matching_groups")
	timeFormatFlag := cmd.PersistentFlags().Lookup("time_format")
	logLabelsFlag := cmd.PersistentFlags().Lookup("log_labels")
	filterFlag := cmd.PersistentFlags().Lookup("filter")

	assert.True(t, cmd.HasAvailablePersistentFlags(), "transition command should have available flag")
	assert.NotNil(t, fileFlag, "transition command should have \"file\" flag")
	assert.Equal(t, "f", fileFlag.Shorthand, "\"file\" flag's shorthand is \"f\"")
	assert.Equal(t, "string", fileFlag.Value.Type(), "\"file\" flag is string")
	assert.NotNil(t, matchingGroupsFlag, "transition command should have \"matching_groups\" flag")
	assert.Equal(t, "m", matchingGroupsFlag.Shorthand, "\"matching_groups\" flag's shorthand is \"m\"")
	assert.Equal(t, "stringSlice", matchingGroupsFlag.Value.Type(), "\"matching_groups\" flag is string slice")
	assert.NotNil(t, timeFormatFlag, "transition command should have \"time_format\" flag")
	assert.Equal(t, "string", timeFormatFlag.Value.Type(), "\"time_format\" flag is string")
	assert.NotNil(t, logLabelsFlag, "transition command should have \"log_labels\" flag")
	assert.Equal(t, "stringToString", logLabelsFlag.Value.Type(), "\"log_labels\" flag is stringToString")
	assert.NotNil(t, filterFlag, "transition command should have \"filter\" flag")
	assert.Equal(t, "string", filterFlag.Value.Type(), "\"filter\" flag is string")
}
func TestNewRootCmd_Flag_Verbose(t *testing.T) {
	fs := afero.NewMemMapFs()
	v := viper.New()
	cmd := NewRootCmd(v, fs)
	_ = cmd.BindFlags()

	v.Set("verbose", true)

	cmd.Run = func(cmd *cobrax.Command, args []string) {} // dummy function to make command runnable
	cmd.SetArgs([]string{})                               // dummy not to use os.Args[1:]
	err := cmd.Execute()

	assert.Nil(t, err)
}

func TestNewRootCmd_Flag_Debug(t *testing.T) {
	fs := afero.NewMemMapFs()
	v := viper.New()
	cmd := NewRootCmd(v, fs)
	_ = cmd.BindFlags()

	v.Set("debug", true)

	cmd.Run = func(cmd *cobrax.Command, args []string) {} // dummy function to make command runnable
	cmd.SetArgs([]string{})                               // dummy not to use os.Args[1:]
	err := cmd.Execute()

	assert.Nil(t, err)
}

func TestNewRootCmd_ConfigFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	v := viper.New()
	cmd := NewRootCmd(v, fs)
	_ = cmd.BindFlags()

	_, _ = fs.Create(".stool.yaml")

	cmd.Run = func(cmd *cobrax.Command, args []string) {} // dummy function to make command runnable
	cmd.SetArgs([]string{"--config", ".stool.yaml"})
	err := cmd.Execute()

	assert.Nil(t, err)
}

func TestExecute(t *testing.T) {
	testdata := filepath.Join("testdata", t.Name())
	v := viper.New()
	fs := afero.NewOsFs()
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCmd(v, fs)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"--config", "testdata/.stool.yaml"})
	_ = cmd.BindFlags()

	assert.NoError(t, cmd.Execute())

	c := golden.New(t, flagUpdateGolden, testdata, "usage")
	if diff := c.Check("_stdout", stdout); diff != "" {
		t.Error("stdout\n", diff)
	}
	if diff := c.Check("_stderr", stderr); diff != "" {
		t.Error("stderr\n", diff)
	}
}
