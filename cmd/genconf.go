package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/types"
	"log/slog"
	"strings"

	"github.com/fatih/color"
	"github.com/haijima/stool/internal/genconf"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func NewGenConfCmd(v *viper.Viper, fs afero.Fs) *cobra.Command {
	genConfCmd := &cobra.Command{}
	genConfCmd.Use = "genconf [-f format] <filename>"
	genConfCmd.DisableFlagsInUseLine = true
	genConfCmd.Short = "Generate configuration file"
	genConfCmd.Long = `Extract the routing information from the source code and generate the "matching_group" configuration.

The web framework used in the source code is automatically detected.
Currently, the following frameworks are supported:
- Echo (https://echo.labstack.com/)

The output format is specified by the -f option.
toml, yaml, and json formats are supported.`
	genConfCmd.Example = "  stool genconf -f yaml main.go >> .stool.yaml"
	genConfCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		return v.BindPFlags(cmd.Flags())
	}
	genConfCmd.RunE = func(cmd *cobra.Command, args []string) error {
		fileName := args[0]
		return runGenConf(cmd, v, fs, fileName)
	}

	genConfCmd.Flags().String("format", "yaml", "The output format {toml|yaml|json|flag}")
	genConfCmd.Flags().Bool("capture-group-name", false, "Add names to captured groups like \"(?P<name>pattern)\"")

	return genConfCmd
}

func runGenConf(cmd *cobra.Command, v *viper.Viper, fs afero.Fs, fileName string) error {
	format := v.GetString("format")
	captureGroupName := v.GetBool("capture-group-name")

	slog.Debug(fmt.Sprintf("%+v", v.AllSettings()))

	if format != "toml" && format != "yaml" && format != "json" && format != "flag" {
		return fmt.Errorf("invalid format: %s", format)
	}

	usedFramework, err := genconf.CheckImportedFramework(fileName)
	if err != nil {
		return err
	}

	var matchingGroups []string
	switch usedFramework.Kind {
	case genconf.EchoV4:
		cmd.PrintErrln("Detected Echo: \"github.com/labstack/echo/v4\"")

		var anblErr *genconf.ArgNotBasicLitError
		matchingGroups, err = genconf.GenMatchingGroupFromEchoV4(fileName, usedFramework.PkgName, captureGroupName)
		if err != nil {
			if errors.As(err, &anblErr) {
				printArgNotBasicLitError(cmd, anblErr)
			} else {
				return err
			}
		}
	case genconf.None:
		return fmt.Errorf("not found web framework from %s", fileName)
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
	fmt.Fprintln(cmd.OutOrStdout(), fmt.Sprintf("-m '%s'", strings.Join(conf.MatchingGroups, ",")))
	return nil
}

func printArgNotBasicLitError(cmd *cobra.Command, err *genconf.ArgNotBasicLitError) {
	for _, x := range err.Info {
		tag := color.YellowString("[Warning]")
		pos := x.ArgPos.String()
		msg := fmt.Sprintf("Unable to parse %T", x.Call.Args[x.ArgIndex])
		args := make([]string, 0, len(x.Call.Args))
		for _, a := range x.Call.Args {
			args = append(args, types.ExprString(a))
		}
		args[x.ArgIndex] = color.New(color.Underline).Sprint(args[x.ArgIndex])
		expr := fmt.Sprintf("%s(%s)", types.ExprString(x.Call.Fun), strings.Join(args, ", "))
		cmd.PrintErrf("%s %s %s:\t%s\n", tag, pos, msg, expr)
	}
}
