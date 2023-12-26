package cmd

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/haijima/stool/internal"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

func init() {

}

// NewRootCmd returns the base command used when called without any subcommands
func NewRootCmd(v *viper.Viper, fs afero.Fs) *cobra.Command {
	rootCmd := &cobra.Command{}
	rootCmd.Use = "stool"
	rootCmd.Short = "stool is access log profiler"
	rootCmd.Args = cobra.NoArgs
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.SilenceUsage = true // don't show help content when error occurred
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if v.GetInt("verbosity") >= 2 {
			log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
			log.SetOutput(cmd.ErrOrStderr())
		} else if v.GetInt("verbosity") >= 1 {
			log.SetOutput(cmd.ErrOrStderr())
		} else {
			log.SetOutput(io.Discard)
		}

		color.NoColor = color.NoColor || v.GetBool("no_color")

		return nil
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $XDG_CONFIG_HOME/.stool.yaml)")
	//rootCmd.PersistentFlags().BoolP("version", "v", false, "Show the version of this command")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "quiet output")
	_ = v.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	rootCmd.PersistentFlags().Bool("no_color", false, "disable colorized output")
	_ = v.BindPFlag("no_color", rootCmd.PersistentFlags().Lookup("no_color"))

	rootCmd.AddCommand(NewTrendCommand(internal.NewTrendProfiler(), v, fs))
	rootCmd.AddCommand(NewTransitionCmd(internal.NewTransitionProfiler(), v, fs))
	rootCmd.AddCommand(NewScenarioCmd(internal.NewScenarioProfiler(), v, fs))
	rootCmd.AddCommand(NewParamCommand(internal.NewParamProfiler(), v, fs))
	rootCmd.AddCommand(NewAaCommand(v, fs))
	rootCmd.AddCommand(NewGenConfCmd(v, fs))

	err := ReadConfigFile(*v)
	cobra.CheckErr(err)

	return rootCmd
}

func OpenOrStdIn(filename string, fs afero.Fs, stdin io.Reader) (io.ReadCloser, error) {
	if filename != "" {
		f, err := fs.Open(filename)
		if err != nil {
			return nil, err
		}
		return f, nil
	} else {
		return io.NopCloser(stdin), nil
	}
}

func ReadConfigFile(v viper.Viper) error {
	//cobra.OnInitialize(func() {
	if cfgFile != "" {
		v.SetConfigFile(cfgFile) // Use config file from the flag.
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		v.AddConfigPath(wd) // adding current working directory as first search path
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			xdgConfig = filepath.Join(home, ".config")
		}
		v.AddConfigPath(filepath.Join(xdgConfig, "stool")) // adding XDG config directory as second search path
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		v.AddConfigPath(home) // adding home directory as third search path
		v.SetConfigName(".stool")
	}

	// If a config file is found, read it in.
	if err := v.ReadInConfig(); err == nil {
		log.Printf("Using config file: %s", v.ConfigFileUsed())
		log.Printf("%+v", v.AllSettings())
	}
	//})
	return nil
}
