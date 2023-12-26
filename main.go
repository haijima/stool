package main

import (
	"os"
	"regexp"
	"strconv"

	"github.com/haijima/stool/cmd"
	"github.com/mattn/go-colorable"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	v := viper.New()
	fs := afero.NewOsFs()
	v.SetFs(fs)
	rootCmd := cmd.NewRootCmd(v, fs)
	rootCmd.SetOut(colorable.NewColorableStdout())
	rootCmd.SetErr(colorable.NewColorableStderr())
	apply(rootCmd)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func apply(c *cobra.Command) {
	level, others := extract(os.Args[1:])
	c.SetArgs(others)
	cobra.OnInitialize(func() {
		verbosity, err := c.Flags().GetInt("verbosity")
		if err == nil && level > verbosity {
			err := c.Flags().Set("verbosity", strconv.Itoa(level))
			if err != nil {
				panic(err)
			}
		}
	})
}

var verbosityRegex = regexp.MustCompile(`^-v+$`)

func extract(xs []string) (int, []string) {
	var level int
	others := make([]string, 0, len(xs))
	for _, x := range xs {
		if verbosityRegex.MatchString(x) && (len(x)-1 > level) {
			level = len(x) - 1
		} else {
			others = append(others, x)
		}
	}
	return level, others
}
