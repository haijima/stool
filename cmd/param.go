package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/haijima/cobrax"
	"github.com/haijima/gini"
	"github.com/haijima/stool/internal"
	"github.com/haijima/stool/internal/log"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// NewParamCmd returns the param command
func NewParamCmd(p *internal.ParamProfiler, v *viper.Viper, fs afero.Fs) *cobra.Command {
	paramCmd := &cobra.Command{}
	paramCmd.Use = "param [flags] <matching_group>..."
	paramCmd.Short = "Show the parameter statistics for each endpoint"
	paramCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runParam(cmd, v, fs, p, args)
	}
	paramCmd.Args = cobra.MinimumNArgs(1)

	paramCmd.Flags().StringP("type", "t", "all", "The type of the parameter {path|query|all}")
	paramCmd.Flags().IntP("num", "n", 5, "The number of parameters to show")
	paramCmd.Flags().Bool("stat", false, "Show statistics of the parameters")
	paramCmd.Flags().String("format", "table", "The stat output format {table|md|csv|tsv}")
	paramCmd.Flags().StringP("file", "f", "", "access log file to profile")
	paramCmd.Flags().String("time_format", "02/Jan/2006:15:04:05 -0700", "format to parse time field on log file")
	paramCmd.Flags().StringToString("log_labels", map[string]string{}, "comma-separated list of key=value pairs to override log labels")
	paramCmd.Flags().String("filter", "", "filter log lines by regular expression")
	_ = paramCmd.MarkFlagFilename("file", viper.SupportedExts...)

	return paramCmd
}

func runParam(cmd *cobra.Command, v *viper.Viper, fs afero.Fs, p *internal.ParamProfiler, args []string) error {
	timeFormat := v.GetString("time_format")
	labels := v.GetStringMapString("log_labels")
	filter := v.GetString("filter")
	paramType := v.GetString("type")
	num := v.GetInt("num")
	statFlg := v.GetBool("stat")
	format := v.GetString("format")

	paramType = strings.ToLower(paramType)
	if paramType != "path" && paramType != "query" && paramType != "all" {
		return fmt.Errorf("type flag should be 'path', 'query' or 'all'. but: %s", paramType)
	}
	if format != "table" && format != "md" && format != "csv" && format != "tsv" {
		return fmt.Errorf("format flag should be 'table', 'md', 'csv' or 'tsv'. but: %s", format)
	}

	f, err := cobrax.OpenOrStdIn(v.GetString("file"), fs, cobrax.WithStdin(cmd.InOrStdin()))
	if err != nil {
		return err
	}
	defer f.Close()
	logReader, err := log.NewLTSVReader(f, log.LTSVReadOpt{
		MatchingGroups: args,
		TimeFormat:     timeFormat,
		Labels:         labels,
		Filter:         filter,
	})
	if err != nil {
		return err
	}

	result, err := p.Profile(logReader)
	if err != nil {
		return err
	}

	if statFlg {
		printParamStat(cmd, result, paramType, format)
	} else {
		printParamResult(cmd, result, paramType, num, v.GetBool("quiet"))
	}
	return nil
}

type kv struct {
	Key   string
	Value int
}

func printParamResult(cmd *cobra.Command, result *internal.Param, paramType string, displayNum int, quiet bool) {
	for _, k := range result.Endpoints {
		v := result.Count[k]
		pathParams, hasPathParam := result.Path[k]
		queryParams, hasQuery := result.QueryValue[k]
		if paramType != "all" && !hasPathParam && !hasQuery {
			if !quiet {
				if strings.HasPrefix(k, "GET ") {
					cmd.PrintErrln(color.YellowString(fmt.Sprintf("[Warning] Neither path parameter nor query parameter for %q", k)))
				} else {
					cmd.PrintErrln(color.YellowString(fmt.Sprintf("[Warning] No path parameter for %q", k)))
				}
				cmd.PrintErrln("Use capture group of regular expression to get path parameters. e.g. \"/users/([^/]+)\" or \"/users/(?P<id>[0-9]+)/posts\"")
				cmd.PrintErrln()
			}
			continue // has no param
		} else if paramType == "path" && !hasPathParam {
			if !quiet {
				cmd.PrintErrln(color.YellowString(fmt.Sprintf("[Warning] No path parameter for %q", k)))
				cmd.PrintErrln("Use capture group of regular expression to get path parameters. e.g. \"/users/([^/]+)\" or \"/users/(?P<id>[0-9]+)/posts\"")
				cmd.PrintErrln("When you want to show query parameters, please use \"-t query\" or \"-t all\" option")
				cmd.PrintErrln()
			}
			continue // has no path param
		} else if paramType == "query" && !hasQuery {
			if !quiet && strings.HasPrefix(k, "GET ") {
				cmd.PrintErrln(color.YellowString(fmt.Sprintf("[Warning] No query parameter for %q", k)))
				cmd.PrintErrln("When you want to show path parameters, please use \"-t path\" or \"-t all\" option")
				cmd.PrintErrln()
			}
			continue // has no query param
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%s (Count: %s)\n", color.New(color.FgHiBlue, color.Underline).Sprint(k), emphasisInt(v))

		if hasPathParam && (paramType == "path" || paramType == "all") {
			printPathParamsResult(cmd, pathParams, result.PathName[k], displayNum, v)
		}
		if hasQuery && (paramType == "query" || paramType == "all") {
			printQueryResult(cmd, result, displayNum, queryParams, k, v)
		}
	}
}

func printPathParamsResult(cmd *cobra.Command, pathParams []map[string]int, pathNames []string, displayNum int, v int) {
	for i, vv := range pathParams {
		ks := len(vv)
		var paramName string
		if pathNames[i] != "" {
			paramName = ":" + color.CyanString(pathNames[i])
		} else {
			paramName = color.CyanString(fmt.Sprintf("%s path parameter", humanize.Ordinal(i+1)))
		}
		g, _ := gini.Gini(maps.Values(vv))
		fmt.Fprintf(cmd.OutOrStdout(), "\t%s (Cardinality: %s, Gini: %s)\n", paramName, emphasisInt(ks), printGini(g, true))

		var ss []kv
		for kkk, vvv := range vv {
			ss = append(ss, kv{kkk, vvv})
		}
		sort.Slice(ss, func(i, j int) bool { return ss[i].Value > ss[j].Value })

		if ks > displayNum {
			ss = ss[:displayNum]
		}
		var p int
		for _, s := range ss {
			p += s.Value
			fmt.Fprintf(cmd.OutOrStdout(), "\t\t%s: %s (Cum: %s)\n", color.GreenString(s.Key), emphasisInt(s.Value), emphasisPercentage(p, v))
		}
		if ks > displayNum {
			fmt.Fprintf(cmd.OutOrStdout(), "\t\t... and %s more\n", humanize.Comma(int64(ks-displayNum)))
		}
	}
	fmt.Fprintln(cmd.OutOrStdout())
}

func printQueryResult(cmd *cobra.Command, result *internal.Param, displayNum int, queryParams map[string]map[string]int, k string, v int) {
	//cmd.PrintOutln("\tQuery parameter")
	queryKeys := maps.Keys(queryParams)
	slices.Sort(queryKeys)
	for _, kk := range queryKeys {
		vv := queryParams[kk]
		ks := len(vv)
		g, _ := gini.Gini(maps.Values(vv))
		fmt.Fprintf(cmd.OutOrStdout(), "\t?%s (Count: %s, Rate: %s, Cardinality: %s, Gini: %s)\n", color.MagentaString(kk), emphasisInt(result.QueryKey[k][kk]), emphasisPercentage(result.QueryKey[k][kk], v), emphasisInt(ks), printGini(g, true))
		var ss []kv
		for kkk, vvv := range vv {
			ss = append(ss, kv{kkk, vvv})
		}
		sort.Slice(ss, func(i, j int) bool { return ss[i].Value > ss[j].Value })

		if ks > displayNum {
			ss = ss[:displayNum]
		}
		var p int
		for _, s := range ss {
			p += s.Value
			fmt.Fprintf(cmd.OutOrStdout(), "\t\t%s: %s (Cum: %s)\n", color.GreenString(s.Key), emphasisInt(s.Value), emphasisPercentage(p, result.QueryKey[k][kk]))
		}
		if ks > displayNum {
			fmt.Fprintf(cmd.OutOrStdout(), "\t\t... and %s more\n", humanize.Comma(int64(ks-displayNum)))
		}
	}

	if len(result.QueryKeyCombination[k]) > 1 {
		var qkcSum int
		var ss []kv
		for kk, vv := range result.QueryKeyCombination[k] {
			ss = append(ss, kv{kk, vv})
			qkcSum += vv
		}
		if v-qkcSum > 0 {
			ss = append(ss, kv{"(none)", v - qkcSum})
		}
		sort.Slice(ss, func(i, j int) bool { return ss[i].Value > ss[j].Value })
		ks := len(ss)

		g, _ := gini.Gini(maps.Values(result.QueryKeyCombination[k]))
		fmt.Fprintf(cmd.OutOrStdout(), "\n\tQuery key combination (Cardinality: %s, Gini: %s)\n", emphasisInt(ks), printGini(g, true))

		if ks > displayNum {
			ss = ss[:displayNum]
		}
		var p int
		for _, s := range ss {
			p += s.Value

			qks := strings.Split(s.Key, "&")
			for i, qk := range qks {
				qks[i] = color.MagentaString(qk)
			}
			qkc := strings.Join(qks, "&")
			if s.Key != "(none)" {
				qkc = "?" + qkc
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\t\t%s (Count: %s, Cum: %s)\n", qkc, emphasisInt(s.Value), emphasisPercentage(p, v))
		}
		if ks > displayNum {
			fmt.Fprintf(cmd.OutOrStdout(), "\t\t... and %s more\n", humanize.Comma(int64(ks-displayNum)))
		}
	}
	if len(result.QueryKeyCombination[k]) > 1 {
		var qkcSum int
		var ss []kv
		for kk, vv := range result.QueryValueCombination[k] {
			ss = append(ss, kv{kk, vv})
			qkcSum += vv
		}
		if v-qkcSum > 0 {
			ss = append(ss, kv{"(none)", v - qkcSum})
		}
		sort.Slice(ss, func(i, j int) bool { return ss[i].Value > ss[j].Value })
		ks := len(ss)

		g, _ := gini.Gini(maps.Values(result.QueryValueCombination[k]))
		fmt.Fprintf(cmd.OutOrStdout(), "\n\tQuery key value combination (Cardinality: %s, Gini: %s)\n", emphasisInt(ks), printGini(g, true))

		if ks > displayNum {
			ss = ss[:displayNum]
		}
		var p int
		for _, s := range ss {
			p += s.Value

			qkvs := strings.Split(s.Key, "&")
			for i, qk := range qkvs {
				qkv := strings.Split(qk, "=")
				if len(qkv) == 2 {
					qkvs[i] = color.MagentaString(qkv[0]) + "=" + color.GreenString(qkv[1])
				}
			}
			qkvc := strings.Join(qkvs, "&")
			if s.Key != "(none)" {
				qkvc = "?" + qkvc
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\t\t%s (Count: %s, Cum: %s)\n", qkvc, emphasisInt(s.Value), emphasisPercentage(p, v))
		}
		if ks > displayNum {
			fmt.Fprintf(cmd.OutOrStdout(), "\t\t... and %s more\n", humanize.Comma(int64(ks-displayNum)))
		}
	}
	fmt.Fprintln(cmd.OutOrStdout())
}

func printParamStat(cmd *cobra.Command, result *internal.Param, paramType, format string) {
	rows := make([]table.Row, 0)
	for _, k := range result.Endpoints {
		v := result.Count[k]
		pathParams, hasPathParam := result.Path[k]
		queryParams, hasQuery := result.QueryValue[k]
		if !hasPathParam && !hasQuery {
			continue
		}

		if hasPathParam && (paramType == "path" || paramType == "all") {
			for i, vv := range pathParams {
				var paramName string
				if result.PathName[k][i] != "" {
					paramName = ":" + result.PathName[k][i]
				} else {
					paramName = fmt.Sprintf("Path param(%d)", i+1)
				}
				g, _ := gini.Gini(maps.Values(vv))
				rows = append(rows, table.Row{
					k,
					"path",
					paramName,
					humanize.Comma(int64(v)),
					color.HiBlackString("100.00"),
					humanize.Comma(int64(len(vv))),
					printGini(g, false),
				})
			}
		}

		if hasQuery && (paramType == "query" || paramType == "all") {
			queryKeys := maps.Keys(queryParams)
			slices.Sort(queryKeys)
			for _, kk := range queryKeys {
				vv := queryParams[kk]
				g, _ := gini.Gini(maps.Values(vv))
				rows = append(rows, table.Row{
					k,
					"query",
					fmt.Sprintf("?%s", kk),
					humanize.Comma(int64(result.QueryKey[k][kk])),
					fmt.Sprintf("%.2f", float64(result.QueryKey[k][kk])/float64(v)*100),
					humanize.Comma(int64(len(vv))),
					printGini(g, false),
				})
			}
		}
	}

	t := table.NewWriter()
	t.SetOutputMirror(cmd.OutOrStdout())
	header := table.Row{"Endpoint", "Type", "Parameter", "Count", "Count(%)", "Cardinality", "Gini"}
	t.AppendHeader(header)

	aligns := []table.ColumnConfig{{Number: 4, Align: text.AlignRight}, {Number: 5, Align: text.AlignRight}, {Number: 6, Align: text.AlignRight}, {Number: 7, Align: text.AlignRight}}
	t.SetColumnConfigs(aligns)
	t.AppendRows(rows)

	if format == "csv" {
		t.RenderCSV()
	} else if format == "tsv" {
		t.RenderTSV()
	} else if format == "table" {
		t.Render()
	} else if format == "md" {
		t.RenderMarkdown()
	} else {
		cmd.PrintErrf("invalid format: %s\n", format)
	}
}

func emphasisInt(num int) string {
	return color.New(color.Bold).Sprint(humanize.Comma(int64(num)))
}

func emphasisPercentage(numerator, denominator int) string {
	return color.New(color.Bold).Sprintf("%.2f%%", float64(numerator)/float64(denominator)*100)
}

func printGini(gini float64, bold bool) string {
	c := color.New()
	if gini < 0.25 {
		// Noop
	} else if gini < 0.4 {
		c.Add(color.FgYellow)
	} else {
		c.Add(color.FgRed)
	}
	if bold {
		c.Add(color.Bold)
	}
	return c.Sprint(fmt.Sprintf("%.3f", gini))
}
