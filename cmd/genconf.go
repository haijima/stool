package cmd

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/haijima/cobrax"
	"github.com/haijima/epf"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func NewGenConfCmd(v *viper.Viper, fs afero.Fs) *cobra.Command {
	genConfCmd := &cobra.Command{}
	genConfCmd.Use = "genconf"
	genConfCmd.Short = "Generate configuration file"
	genConfCmd.Example = "  stool genconf --format yaml > .stool.yaml"
	genConfCmd.Args = cobra.NoArgs
	genConfCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runGenConf(cmd, v, fs)
	}

	genConfCmd.Flags().StringP("dir", "d", ".", "The directory to analyze")
	genConfCmd.Flags().StringP("pattern", "p", "./...", "The pattern to analyze")
	genConfCmd.Flags().String("format", "yaml", "The output format {toml|yaml|json|flag}")
	genConfCmd.Flags().Bool("capture-group-name", false, "Add names to captured groups like \"(?P<name>pattern)\"")

	return genConfCmd
}

func runGenConf(cmd *cobra.Command, v *viper.Viper, _ afero.Fs) error {
	dir := v.GetString("dir")
	pattern := v.GetString("pattern")
	format := v.GetString("format")
	if format != "toml" && format != "yaml" && format != "json" {
		return fmt.Errorf("invalid format: %s", format)
	}

	flags := cobrax.GetFlags(cmd.Root())
	matchingGroups, err := getMatchingGroups(dir, pattern)
	if err != nil {
		return err
	}
	flags["matching-groups"] = matchingGroups

	switch format {
	case "toml":
		return toml.NewEncoder(cmd.OutOrStdout()).Encode(flags)
	case "yaml":
		return yaml.NewEncoder(cmd.OutOrStdout()).Encode(flags)
	case "json":
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(flags)
	}
	return nil
}

func getMatchingGroups(dir, pattern string) ([]string, error) {
	ext, err := epf.AutoExtractor(dir, pattern)
	if err != nil {
		return nil, err
	}
	endpoints, err := epf.FindEndpoints(dir, pattern, ext)
	if err != nil {
		return nil, err
	}
	matchingGroups := make([]string, 0, len(endpoints))
	for _, endpoint := range endpoints {
		matchingGroups = append(matchingGroups, endpoint.PathRegexpPattern)
	}
	slices.Sort(matchingGroups)
	slices.Reverse(matchingGroups)
	matchingGroups = slices.Compact(matchingGroups)
	return matchingGroups, nil
}
