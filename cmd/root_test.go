package cmd

import (
	"testing"

	"github.com/haijima/cobrax"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewRootCmd(t *testing.T) {
	fs := afero.NewMemMapFs()
	v := viper.New()
	cmd := NewRootCmd(v, fs)
	_ = cmd.BindFlags()

	assert.Equal(t, "stool", cmd.Name(), "NewRootCommand() should return command named \"stool\". but: \"%s\"", cmd.Name())
	assert.False(t, cmd.HasParent(), "RootCommand should not have parent command.")
	assert.Equal(t, 4, len(cmd.Commands()), "RootCommand should have 1 sub command. but: %d", len(cmd.Commands()))
	assert.False(t, cmd.Runnable(), "RootCommand should not runnable.")
}

func TestNewRootCmd_Flag(t *testing.T) {
	fs := afero.NewMemMapFs()
	v := viper.New()
	cmd := NewRootCmd(v, fs)
	_ = cmd.BindFlags()
	fileFlag := cmd.PersistentFlags().Lookup("file")
	matchingGroupsFlag := cmd.PersistentFlags().Lookup("matching_groups")
	ignorePatternsFlag := cmd.PersistentFlags().Lookup("ignore_patterns")
	timeFormatFlag := cmd.PersistentFlags().Lookup("time_format")
	logLabelsFlag := cmd.PersistentFlags().Lookup("log_labels")

	assert.True(t, cmd.HasAvailablePersistentFlags(), "transition command should have available flag")
	assert.NotNil(t, fileFlag, "transition command should have \"file\" flag")
	assert.Equal(t, "f", fileFlag.Shorthand, "\"file\" flag's shorthand is \"f\"")
	assert.Equal(t, "string", fileFlag.Value.Type(), "\"file\" flag is string")
	assert.NotNil(t, matchingGroupsFlag, "transition command should have \"matching_groups\" flag")
	assert.Equal(t, "m", matchingGroupsFlag.Shorthand, "\"matching_groups\" flag's shorthand is \"m\"")
	assert.Equal(t, "stringSlice", matchingGroupsFlag.Value.Type(), "\"matching_groups\" flag is string slice")
	assert.NotNil(t, ignorePatternsFlag, "transition command should have \"ignore_patterns\" flag")
	assert.Equal(t, "stringSlice", ignorePatternsFlag.Value.Type(), "\"ignore_patterns\" flag is string slice")
	assert.NotNil(t, timeFormatFlag, "transition command should have \"time_format\" flag")
	assert.Equal(t, "string", timeFormatFlag.Value.Type(), "\"time_format\" flag is string")
	assert.NotNil(t, logLabelsFlag, "transition command should have \"log_labels\" flag")
	assert.Equal(t, "stringToString", logLabelsFlag.Value.Type(), "\"log_labels\" flag is stringToString")
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
	fs := afero.NewMemMapFs()
	v := viper.New()
	cmd := NewRootCmd(v, fs)
	_ = cmd.BindFlags()

	err := cmd.Execute()

	assert.Nil(t, err)
}
