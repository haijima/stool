package cmd

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/haijima/gini"
	"github.com/haijima/stool/internal"
	"github.com/haijima/stool/internal/log"
	"github.com/olekukonko/tablewriter"
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
	paramCmd.Flags().StringSliceP("matching_groups", "m", []string{}, "comma-separated list of regular expression patterns to group matched URIs")
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

	slog.Debug(fmt.Sprintf("%+v", v.AllSettings()))

	paramType = strings.ToLower(paramType)
	if paramType != "path" && paramType != "query" && paramType != "all" {
		return fmt.Errorf("type flag should be 'path', 'query' or 'all'. but: %s", paramType)
	}
	if format != "table" && format != "md" && format != "csv" && format != "tsv" {
		return fmt.Errorf("format flag should be 'table', 'md', 'csv' or 'tsv'. but: %s", format)
	}

	f, err := OpenOrStdIn(v.GetString("file"), fs, cmd.InOrStdin())
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
	validKeys := make([]string, 0)
	for _, k := range result.Endpoints {
		v := result.Count[k]
		pathParams, hasPathParam := result.Path[k]
		queryParams, hasQuery := result.QueryValue[k]
		if paramType != "all" && !hasPathParam && !hasQuery {
			if !quiet {
				if strings.HasPrefix(k, "GET ") {
					cmd.PrintErrln(color.YellowString(fmt.Sprintf("[Warning] Neither path parameter nor query parameter for \"%s\"", k)))
				} else {
					cmd.PrintErrln(color.YellowString(fmt.Sprintf("[Warning] No path parameter for \"%s\"", k)))
				}
				cmd.PrintErrln("Use capture group of regular expression to get path parameters. e.g. \"/users/([^/]+)\" or \"/users/(?P<id>[0-9]+)/posts\"")
				cmd.PrintErrln()
			}
			continue // has no param
		} else if paramType == "path" && !hasPathParam {
			if !quiet {
				cmd.PrintErrln(color.YellowString(fmt.Sprintf("[Warning] No path parameter for \"%s\"", k)))
				cmd.PrintErrln("Use capture group of regular expression to get path parameters. e.g. \"/users/([^/]+)\" or \"/users/(?P<id>[0-9]+)/posts\"")
				cmd.PrintErrln("When you want to show query parameters, please use \"-t query\" or \"-t all\" option")
				cmd.PrintErrln()
			}
			continue // has no path param
		} else if paramType == "query" && !hasQuery {
			if !quiet && strings.HasPrefix(k, "GET ") {
				cmd.PrintErrln(color.YellowString(fmt.Sprintf("[Warning] No query parameter for \"%s\"", k)))
				cmd.PrintErrln("When you want to show path parameters, please use \"-t path\" or \"-t all\" option")
				cmd.PrintErrln()
			}
			continue // has no query param
		}

		validKeys = append(validKeys, strings.Split(k, " ")[1])
		fmt.Fprintf(cmd.OutOrStdout(), "%s (Count: %s)\n", color.New(color.FgHiBlue, color.Underline).Sprint(k), emphasisInt(v))

		if hasPathParam && (paramType == "path" || paramType == "all") {
			printPathParamsResult(cmd, pathParams, result.PathName[k], displayNum, v)
		}
		if hasQuery && (paramType == "query" || paramType == "all") {
			printQueryResult(cmd, result, displayNum, queryParams, k, v)
		}
	}

	// Show stool param command to call again
	fmt.Fprint(cmd.OutOrStdout(), "stool param")
	for _, k := range validKeys {
		fmt.Fprintf(cmd.OutOrStdout(), " \"%s\"\n", k)
	}
	fmt.Fprintf(cmd.OutOrStdout(), " -n %d\n", displayNum)
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
		fmt.Fprintf(cmd.OutOrStdout(), "\tQuery key combination (Cardinality: %s, Gini: %s)\n", emphasisInt(ks), printGini(g, true))

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
		fmt.Fprintf(cmd.OutOrStdout(), "\tQuery key value combination (Cardinality: %s, Gini: %s)\n", emphasisInt(ks), printGini(g, true))

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
	rows := make([][]string, 0)
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
				row := make([]string, 0, 6)
				row = append(row, k)
				row = append(row, "path")
				row = append(row, paramName)
				row = append(row, humanize.Comma(int64(v)))
				row = append(row, color.HiBlackString("100.00"))
				row = append(row, humanize.Comma(int64(len(vv))))
				row = append(row, printGini(g, false))
				rows = append(rows, row)
			}
		}

		if hasQuery && (paramType == "query" || paramType == "all") {
			queryKeys := maps.Keys(queryParams)
			slices.Sort(queryKeys)
			for _, kk := range queryKeys {
				vv := queryParams[kk]
				g, _ := gini.Gini(maps.Values(vv))
				row := make([]string, 0, 6)
				row = append(row, k)
				row = append(row, "query")
				row = append(row, fmt.Sprintf("?%s", kk))
				row = append(row, humanize.Comma(int64(result.QueryKey[k][kk])))
				row = append(row, fmt.Sprintf("%.2f", float64(result.QueryKey[k][kk])/float64(v)*100))
				row = append(row, humanize.Comma(int64(len(vv))))
				row = append(row, printGini(g, false))
				rows = append(rows, row)
			}
		}
	}

	header := []string{"Endpoint", "Type", "Parameter", "Count", "Count(%)", "Cardinality", "Gini"}
	aligns := []int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT}
	if format == "csv" {
		csvWriter := csv.NewWriter(cmd.OutOrStdout())
		_ = csvWriter.Write(header)
		_ = csvWriter.WriteAll(rows)
	} else if format == "tsv" {
		csvWriter := csv.NewWriter(cmd.OutOrStdout())
		csvWriter.Comma = '\t'
		_ = csvWriter.Write(header)
		_ = csvWriter.WriteAll(rows)
	} else if format == "table" {
		tableWriter := tablewriter.NewWriter(cmd.OutOrStdout())
		tableWriter.SetAutoWrapText(false)
		tableWriter.SetHeader(header)
		tableWriter.SetColumnAlignment(aligns)
		tableWriter.AppendBulk(rows)
		tableWriter.Render()
	} else if format == "md" {
		mdWriter := tablewriter.NewWriter(cmd.OutOrStdout())
		mdWriter.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		mdWriter.SetCenterSeparator("|")
		mdWriter.SetAutoWrapText(false)
		mdWriter.SetHeader(header)
		mdWriter.SetColumnAlignment(aligns)
		mdWriter.AppendBulk(rows)
		mdWriter.Render()
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
