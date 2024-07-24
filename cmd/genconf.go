package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/haijima/stool/internal/genconf"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func NewGenConfCmd(v *viper.Viper, fs afero.Fs) *cobra.Command {
	genConfCmd := &cobra.Command{}
	genConfCmd.Use = "genconf [flags] <dir>"
	genConfCmd.DisableFlagsInUseLine = true
	genConfCmd.Short = "Generate configuration file"
	genConfCmd.Long = `Extract the routing information from the source code and generate the "matching_group" configuration.

The web framework used in the source code is automatically detected.
Currently, the following frameworks are supported:
- Echo (https://echo.labstack.com/)`
	genConfCmd.Example = "  stool genconf -f yaml main.go >> .stool.yaml"
	genConfCmd.Args = cobra.ExactArgs(1)
	genConfCmd.RunE = func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		return runGenConf(cmd, v, fs, dir)
	}

	genConfCmd.Flags().StringP("pattern", "p", "./...", "The pattern to analyze")
	genConfCmd.Flags().String("format", "yaml", "The output format {toml|yaml|json|flag}")
	genConfCmd.Flags().Bool("capture-group-name", false, "Add names to captured groups like \"(?P<name>pattern)\"")

	return genConfCmd
}

func runGenConf(cmd *cobra.Command, v *viper.Viper, fs afero.Fs, dir string) error {
	pattern := v.GetString("pattern")
	format := v.GetString("format")
	captureGroupName := v.GetBool("capture-group-name")

	if format != "toml" && format != "yaml" && format != "json" && format != "flag" {
		return fmt.Errorf("invalid format: %s", format)
	}

	usedFramework, err := genconf.CheckImportedFramework(dir, pattern)
	if err != nil {
		return err
	}
	var ext genconf.APIPathPatternExtractor
	switch usedFramework {
	case genconf.EchoV4:
		slog.Info("Detected Echo: \"github.com/labstack/echo/v4\"")
		ext = &genconf.EchoExtractor{}
	case genconf.None:
		return fmt.Errorf("not found web framework from %s", dir)
	}

	matchingGroups, err := genconf.GenMatchingGroup(dir, pattern, ext, captureGroupName)
	if err != nil {
		return err
	}
	conf := MatchingGroupConf{matchingGroups}
	switch format {
	case "toml":
		return printMatchingGroupInToml(cmd, conf)
	case "yaml":
		return printMatchingGroupInYaml(cmd, conf)
	case "json":
		return printMatchingGroupInJson(cmd, conf)
	case "flag":
		return printMatchingGroupAsFlag(cmd, conf)
	}
	return nil
}

type MatchingGroupConf struct {
	MatchingGroups []string `toml:"matching_groups" yaml:"matching_groups" json:"matching_groups"`
}

func printMatchingGroupInToml(cmd *cobra.Command, conf MatchingGroupConf) error {
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf).SetArraysMultiline(true)
	if err := enc.Encode(conf); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), buf.String())
	return nil
}

func printMatchingGroupInYaml(cmd *cobra.Command, conf MatchingGroupConf) error {
	b, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(b))
	return nil
}

func printMatchingGroupInJson(cmd *cobra.Command, conf MatchingGroupConf) error {
	b, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(b))
	return nil
}

func printMatchingGroupAsFlag(cmd *cobra.Command, conf MatchingGroupConf) error {
	fmt.Fprintf(cmd.OutOrStdout(), "-m '%s'\n", strings.Join(conf.MatchingGroups, ","))
	return nil
}
