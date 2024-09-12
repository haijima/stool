package cmd

import (
	"fmt"

	"github.com/haijima/epf"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewEndpointCmd(v *viper.Viper, fs afero.Fs) *cobra.Command {
	genConfCmd := &cobra.Command{}
	genConfCmd.Use = "endpoint <dir>"
	genConfCmd.Aliases = []string{"endpoints"}
	genConfCmd.DisableFlagsInUseLine = true
	genConfCmd.Short = "Show endpoints"
	genConfCmd.Example = "  stool endpoint ."
	genConfCmd.Args = cobra.ExactArgs(1)
	genConfCmd.RunE = func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		return runEndpoint(cmd, v, fs, dir)
	}

	genConfCmd.Flags().StringP("pattern", "p", "./...", "The pattern to analyze")
	genConfCmd.Flags().String("format", "table", "The output format {table|csv|tsv|md}")

	return genConfCmd
}

func runEndpoint(cmd *cobra.Command, v *viper.Viper, _ afero.Fs, dir string) error {
	pattern := v.GetString("pattern")
	format := v.GetString("format")

	if format != "table" && format != "csv" && format != "tsv" && format != "md" {
		return fmt.Errorf("invalid format: %s", format)
	}

	ext, err := epf.AutoExtractor(dir, pattern)
	if err != nil {
		return err
	}

	endpoints, err := epf.FindEndpoints(dir, pattern, ext)
	if err != nil {
		return err
	}

	t := table.NewWriter()
	t.SetOutputMirror(cmd.OutOrStdout())
	header := table.Row{"#", "Method", "Path", "Function", "Declared Package", "Declared Position"}
	t.AppendHeader(header)

	//aligns := []table.ColumnConfig{{Number: 4, Align: text.AlignRight}, {Number: 5, Align: text.AlignRight}, {Number: 6, Align: text.AlignRight}, {Number: 7, Align: text.AlignRight}}
	//t.SetColumnConfigs(aligns)
	t.AppendRows(endpointsToRows(endpoints))

	switch format {
	case "csv":
		t.RenderCSV()
	case "tsv":
		t.RenderTSV()
	case "table":
		t.Render()
	case "md":
		t.RenderMarkdown()
	}
	return nil
}

func endpointsToRows(endpoints []*epf.Endpoint) []table.Row {
	rows := make([]table.Row, 0, len(endpoints))
	for i, e := range endpoints {
		row := table.Row{i + 1, e.Method, e.Path, e.FuncName, e.DeclarePos.PackagePath(true), e.DeclarePos.PositionString()}
		rows = append(rows, row)
	}
	return rows
}
