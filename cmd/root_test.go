package cmd

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewRootCmd(t *testing.T) {
	fs := afero.NewOsFs()
	v := viper.New()
	cmd := NewRootCmd(v, fs)

	assert.Equal(t, "stool", cmd.Name(), "NewRootCommand() should return command named \"stool\". but: \"%s\"", cmd.Name())
	assert.False(t, cmd.HasParent(), "RootCommand should not have parent command.")
	assert.Equal(t, 1, len(cmd.Commands()), "RootCommand should have 1 sub command. but: %d", len(cmd.Commands()))
	assert.False(t, cmd.Runnable(), "RootCommand should not runnable.")
}

func TestNewRootCmd_Flag_Verbose(t *testing.T) {
	fs := afero.NewOsFs()
	v := viper.New()
	cmd := NewRootCmd(v, fs)

	v.Set("verbose", true)

	cmd.Run = func(cmd *cobra.Command, args []string) {} // dummy function to make command runnable
	cmd.SetArgs([]string{})                              // dummy not to use os.Args[1:]
	cmd.Execute()
}

func TestNewRootCmd_Flag_Debug(t *testing.T) {
	fs := afero.NewOsFs()
	v := viper.New()
	cmd := NewRootCmd(v, fs)

	v.Set("debug", true)

	cmd.Run = func(cmd *cobra.Command, args []string) {} // dummy function to make command runnable
	cmd.SetArgs([]string{})                              // dummy not to use os.Args[1:]
	cmd.Execute()
}

func TestNewRootCmd_ConfigFile(t *testing.T) {
	fs := afero.NewOsFs()
	v := viper.New()
	cmd := NewRootCmd(v, fs)

	fs.Create(".stool.yaml")

	cmd.Run = func(cmd *cobra.Command, args []string) {} // dummy function to make command runnable
	cmd.SetArgs([]string{"--config", ".stool.yaml"})
	cmd.Execute()
}

func TestExecute(t *testing.T) {
	fs := afero.NewOsFs()
	v := viper.New()
	cmd := NewRootCmd(v, fs)

	err := cmd.Execute()

	assert.Nil(t, err)
}