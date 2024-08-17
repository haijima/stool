package cmd

import (
	"fmt"
	"log/slog"

	"github.com/haijima/stool/internal/endpoint"
	"github.com/haijima/stool/internal/genconf"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewEndpointCmd(v *viper.Viper, fs afero.Fs) *cobra.Command {
	genConfCmd := &cobra.Command{}
	genConfCmd.Use = "endpoint <dir>"
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

	usedFramework, err := genconf.CheckImportedFramework(dir, pattern)
	if err != nil {
		return err
	}
	var ext endpoint.Extractor
	switch usedFramework {
	case genconf.EchoV4:
		slog.Info("Detected Echo: \"github.com/labstack/echo/v4\"")
		ext = &endpoint.EchoExtractor{}
	case genconf.Gin:
		slog.Info("Detected Gin: \"github.com/gin-gonic/gin\"")
		return fmt.Errorf("unsupported framework: %v", usedFramework)
	case genconf.ChiV5:
		slog.Info("Detected go-chi: \"github.com/go-chi/chi/v5\"")
		return fmt.Errorf("unsupported framework: %v", usedFramework)
	case genconf.Iris12:
		slog.Info("Detected Iris: \"github.com/kataras/iris/v12\"")
		return fmt.Errorf("unsupported framework: %v", usedFramework)
	case genconf.Gorilla:
		slog.Info("Detected Gorilla: \"github.com/gorilla/mux\"")
		return fmt.Errorf("unsupported framework: %v", usedFramework)
	case genconf.NetHttp:
		slog.Info("Detected \"net/http\"")
	case genconf.None:
		return fmt.Errorf("not found web framework from %s", dir)
	}

	endpoints, err := endpoint.FindEndpoints(dir, pattern, ext)
	if err != nil {
		return err
	}

	t := table.NewWriter()
	t.SetOutputMirror(cmd.OutOrStdout())
	header := table.Row{"Method", "Path", "Function", "Declared Package", "Declared Position", "Function Package", "Function Position"}
	t.AppendHeader(header)

	aligns := []table.ColumnConfig{{Number: 4, Align: text.AlignRight}, {Number: 5, Align: text.AlignRight}, {Number: 6, Align: text.AlignRight}, {Number: 7, Align: text.AlignRight}}
	t.SetColumnConfigs(aligns)
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

func endpointsToRows(endpoints []*endpoint.Endpoint) []table.Row {
	rows := make([]table.Row, 0, len(endpoints))
	for _, e := range endpoints {
		row := table.Row{e.Method, e.Path, e.FuncName, e.DeclarePos.PackagePath(), e.DeclarePos.FLC(), e.FuncPos.PackagePath(), e.FuncPos.FLC()}
		rows = append(rows, row)
	}
	return rows
}
