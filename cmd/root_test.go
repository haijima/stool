package cmd

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewRootCmd(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewRootCmd(v, fs)

	assert.Equal(t, "stool", cmd.Name(), "NewRootCommand() should return command named \"stool\". but: %q", cmd.Name())
	assert.False(t, cmd.HasParent(), "RootCommand should not have parent command.")
	assert.Equal(t, 7, len(cmd.Commands()), "RootCommand should have 1 sub command. but: %d", len(cmd.Commands()))
	assert.False(t, cmd.Runnable(), "RootCommand should not runnable.")
}

func TestNewRootCmd_Flag(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewRootCmd(v, fs)
	versionFlag := cmd.PersistentFlags().Lookup("version")
	configFlag := cmd.PersistentFlags().Lookup("config")
	noColorFlag := cmd.PersistentFlags().Lookup("no-color")
	verboseFlag := cmd.PersistentFlags().Lookup("verbose")
	quietFlag := cmd.PersistentFlags().Lookup("quiet")

	assert.True(t, cmd.HasAvailablePersistentFlags(), "root command should have available flag")

	assert.NotNil(t, versionFlag, "root command should have \"version\" flag")
	assert.Equal(t, "V", versionFlag.Shorthand, "\"version\" flag doesn't have shorthand")
	assert.Equal(t, "bool", versionFlag.Value.Type(), "\"version\" flag is bool")
	assert.NotNil(t, configFlag, "root command should have \"config\" flag")
	assert.Equal(t, "", configFlag.Shorthand, "\"config\" flag doesn't have shorthand")
	assert.Equal(t, "string", configFlag.Value.Type(), "\"config\" flag is string")
	assert.NotNil(t, noColorFlag, "root command should have \"no-color\" flag")
	assert.Equal(t, "", noColorFlag.Shorthand, "\"no-color\" flag doesn't have shorthand")
	assert.Equal(t, "bool", noColorFlag.Value.Type(), "\"no-color\" flag is bool")
	assert.NotNil(t, verboseFlag, "root command should have \"verbose\" flag")
	assert.Equal(t, "v", verboseFlag.Shorthand, "\"verbose\" flag doesn't have shorthand")
	assert.Equal(t, "count", verboseFlag.Value.Type(), "\"verbose\" flag is count")
	assert.NotNil(t, quietFlag, "root command should have \"quiet\" flag")
	assert.Equal(t, "q", quietFlag.Shorthand, "\"quiet\" flag doesn't have shorthand")
	assert.Equal(t, "bool", quietFlag.Value.Type(), "\"quiet\" flag is bool")
}
func TestNewRootCmd_Flag_Verbose(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewRootCmd(v, fs)

	v.Set("verbose", true)

	cmd.Run = func(cmd *cobra.Command, args []string) {} // dummy function to make command runnable
	cmd.SetArgs([]string{})                              // dummy not to use os.Args[1:]
	err := cmd.Execute()

	assert.NoError(t, err)
}

func TestNewRootCmd_Flag_Debug(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewRootCmd(v, fs)

	v.Set("debug", true)

	cmd.Run = func(cmd *cobra.Command, args []string) {} // dummy function to make command runnable
	cmd.SetArgs([]string{})                              // dummy not to use os.Args[1:]
	err := cmd.Execute()

	assert.NoError(t, err)
}

func TestNewRootCmd_ConfigFile(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewRootCmd(v, fs)

	_, _ = fs.Create(".stool.yaml")

	cmd.Run = func(cmd *cobra.Command, args []string) {} // dummy function to make command runnable
	cmd.SetArgs([]string{"--config", ".stool.yaml"})
	err := cmd.Execute()

	assert.NoError(t, err)
}

func TestExecute(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewRootCmd(v, fs)

	err := cmd.Execute()

	assert.NoError(t, err)
}

func createViperAndFs() (v *viper.Viper, fs afero.Fs) {
	v = viper.New()
	fs = afero.NewMemMapFs()
	v.SetFs(fs)
	return v, fs
}
