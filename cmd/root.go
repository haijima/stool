/*
Copyright © 2023 haijima

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"github.com/haijima/stool"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type RootCommand struct {
	IO
}

// NewRootCmd returns the base command used when called without any subcommands
func NewRootCmd() *RootCommand {
	return &RootCommand{
		IO: NewStdIO(),
	}
}

func (c *RootCommand) Cmd() *cobra.Command {
	viper.SetFs(c.Fs)

	rootCmd := &cobra.Command{
		Use:          "stool",
		Short:        "stool is access log profiler",
		SilenceUsage: true, // don't show help content when error occurred
	}

	addLoggingOption(rootCmd)
	useConfig(rootCmd)

	rootCmd.PersistentFlags().StringP("file", "f", "", "access log file to profile")
	rootCmd.PersistentFlags().StringSliceP("matching_groups", "m", []string{}, "comma-separated list of regular expression patterns to group matched URIs")
	rootCmd.PersistentFlags().String("time_format", "02/Jan/2006:15:04:05 -0700", "format to parse time field on log file")
	_ = viper.BindPFlag("file", rootCmd.PersistentFlags().Lookup("file"))
	_ = viper.BindPFlag("matching_groups", rootCmd.PersistentFlags().Lookup("matching_groups"))
	_ = viper.BindPFlag("time_format", rootCmd.PersistentFlags().Lookup("time_format"))

	rootCmd.AddCommand(NewTrendCommand(*stool.NewTrendProfiler()).Cmd())

	return rootCmd
}
