package cmd

import (
	"encoding/csv"
	"fmt"
	"sort"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/haijima/cobrax"
	"github.com/haijima/stool/internal"
	"github.com/haijima/stool/internal/log"
	"github.com/haijima/stool/internal/stat"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// NewParamCommand returns the param command
func NewParamCommand(p *internal.ParamProfiler, v *viper.Viper, fs afero.Fs) *cobrax.Command {
	paramCmd := cobrax.NewCommand(v, fs)
	paramCmd.Use = "param"
	paramCmd.Short = "Show the parameters of each endpoint"
	paramCmd.RunE = func(cmd *cobrax.Command, args []string) error {
		return runParam(cmd, p, args)
	}
	paramCmd.Args = cobra.MinimumNArgs(1)

	paramCmd.Flags().StringP("type", "t", "all", "The type of the parameter {path|query|all}")
	paramCmd.Flags().IntP("num", "n", 5, "The number of parameters to show")
	paramCmd.Flags().Bool("stat", false, "Show statistics of the parameters")
	paramCmd.Flags().String("format", "table", "The stat output format (table, md, csv, tsv)")

	return paramCmd
}

func runParam(cmd *cobrax.Command, p *internal.ParamProfiler, args []string) error {
	timeFormat := cmd.Viper().GetString("time_format")
	labels := cmd.Viper().GetStringMapString("log_labels")
	filter := cmd.Viper().GetString("filter")
	paramType := cmd.Viper().GetString("type")
	num := cmd.Viper().GetInt("num")
	statFlg := cmd.Viper().GetBool("stat")
	format := cmd.Viper().GetString("format")
	noColor := cmd.Viper().GetBool("no_color")
	cmd.V.Printf("%+v", cmd.Viper().AllSettings())

	paramType = strings.ToLower(paramType)
	if paramType != "path" && paramType != "query" && paramType != "all" {
		return fmt.Errorf("type flag should be 'path', 'query' or 'all'. but: %s", paramType)
	}
	if format != "table" && format != "md" && format != "csv" && format != "tsv" {
		return fmt.Errorf("format flag should be 'table', 'md', 'csv' or 'tsv'. but: %s", format)
	}

	color.NoColor = color.NoColor || noColor

	f, err := cmd.OpenOrStdIn(cmd.Viper().GetString("file"))
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
		printParamResult(cmd, result, paramType, num)
	}
	return nil
}

type kv struct {
	Key   string
	Value int
}

func printParamResult(cmd *cobrax.Command, result *internal.Param, paramType string, displayNum int) {
	validKeys := make([]string, 0)
	for _, k := range result.Endpoints {
		v := result.Count[k]
		pathParams, hasPathParam := result.Path[k]
		queryParams, hasQuery := result.QueryValue[k]
		if !hasPathParam && !hasQuery {
			continue // has no param
		} else if paramType == "path" && !hasPathParam {
			if !cmd.Viper().GetBool("quiet") {
				cmd.PrintErrln(color.YellowString(fmt.Sprintf("[Warning] No path parameter for \"%s\"", k)))
				cmd.PrintErrln("Use capture group of regular expression to get path parameters. e.g. \"/users/([^/]+)\" or \"/users/(?P<id>[0-9]+)/posts\"")
				cmd.PrintErrln("When you want to show query parameters, please use \"-t query\" or \"-t all\" option")
				cmd.PrintErrln()
			}
			continue // has no path param
		} else if paramType == "query" && !hasQuery {
			if !cmd.Viper().GetBool("quiet") && strings.HasPrefix(k, "GET ") {
				cmd.PrintErrln(color.YellowString(fmt.Sprintf("[Warning] No query parameter for \"%s\"", k)))
				cmd.PrintErrln("When you want to show path parameters, please use \"-t path\" or \"-t all\" option")
				cmd.PrintErrln()
			}
			continue // has no query param
		}

		validKeys = append(validKeys, strings.Split(k, " ")[1])
		cmd.PrintOutf("%s (Count: %s)\n", color.New(color.FgHiBlue, color.Underline).Sprint(k), emphasisInt(v))

		if hasPathParam && (paramType == "path" || paramType == "all") {
			printPathParamsResult(cmd, pathParams, result.PathName[k], displayNum, v)
		}
		if hasQuery && (paramType == "query" || paramType == "all") {
			printQueryResult(cmd, result, displayNum, queryParams, k, v)
		}
	}

	// Show stool param command to call again
	cmd.PrintOut("stool param")
	for _, k := range validKeys {
		cmd.PrintOutf(" \"%s\"", k)
	}
	cmd.PrintOutf(" -n %d\n", displayNum)
}

func printPathParamsResult(cmd *cobrax.Command, pathParams []map[string]int, pathNames []string, displayNum int, v int) {
	for i, vv := range pathParams {
		ks := len(vv)
		var paramName string
		if pathNames[i] != "" {
			paramName = ":" + color.CyanString(pathNames[i])
		} else {
			paramName = color.CyanString(fmt.Sprintf("%s path parameter", humanize.Ordinal(i+1)))
		}
		gini := stat.Gini(maps.Values(vv), stat.Unsorted)
		cmd.PrintOutf("\t%s (Cardinality: %s, Gini: %s)\n", paramName, emphasisInt(ks), printGini(gini, true))

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
			cmd.PrintOutf("\t\t%s: %s (Cum: %s)\n", color.GreenString(s.Key), emphasisInt(s.Value), emphasisPercentage(p, v))
		}
		if ks > displayNum {
			cmd.PrintOutf("\t\t... and %s more\n", humanize.Comma(int64(ks-displayNum)))
		}
	}
	cmd.PrintOutln()
}

func printQueryResult(cmd *cobrax.Command, result *internal.Param, displayNum int, queryParams map[string]map[string]int, k string, v int) {
	//cmd.PrintOutln("\tQuery parameter")
	queryKeys := maps.Keys(queryParams)
	slices.Sort(queryKeys)
	for _, kk := range queryKeys {
		vv := queryParams[kk]
		ks := len(vv)
		gini := stat.Gini(maps.Values(vv), stat.Unsorted)
		cmd.PrintOutf("\t?%s (Count: %s, Rate: %s, Cardinality: %s, Gini: %s)\n", color.MagentaString(kk), emphasisInt(result.QueryKey[k][kk]), emphasisPercentage(result.QueryKey[k][kk], v), emphasisInt(ks), printGini(gini, true))
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
			cmd.PrintOutf("\t\t%s: %s (Cum: %s)\n", color.GreenString(s.Key), emphasisInt(s.Value), emphasisPercentage(p, result.QueryKey[k][kk]))
		}
		if ks > displayNum {
			cmd.PrintOutf("\t\t... and %s more\n", humanize.Comma(int64(ks-displayNum)))
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

		gini := stat.Gini(maps.Values(result.QueryKeyCombination[k]), stat.Unsorted)
		cmd.PrintOutf("\tQuery key combination (Cardinality: %s, Gini: %s)\n", emphasisInt(ks), printGini(gini, true))

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
			cmd.PrintOutf("\t\t%s (Count: %s, Cum: %s)\n", qkc, emphasisInt(s.Value), emphasisPercentage(p, v))
		}
		if ks > displayNum {
			cmd.PrintOutf("\t\t... and %s more\n", humanize.Comma(int64(ks-displayNum)))
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

		gini := stat.Gini(maps.Values(result.QueryValueCombination[k]), stat.Unsorted)
		cmd.PrintOutf("\tQuery key value combination (Cardinality: %s, Gini: %s)\n", emphasisInt(ks), printGini(gini, true))

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
			cmd.PrintOutf("\t\t%s (Count: %s, Cum: %s)\n", qkvc, emphasisInt(s.Value), emphasisPercentage(p, v))
		}
		if ks > displayNum {
			cmd.PrintOutf("\t\t... and %s more\n", humanize.Comma(int64(ks-displayNum)))
		}
	}
	cmd.PrintOutln()
}

func printParamStat(cmd *cobrax.Command, result *internal.Param, paramType, format string) {
	rows := make([][]string, 0)
	validKeys := make([]string, 0)
	for _, k := range result.Endpoints {
		v := result.Count[k]
		validKeys = append(validKeys, strings.Split(k, " ")[1])
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
				gini := stat.Gini(maps.Values(vv), stat.Unsorted)
				row := make([]string, 0, 6)
				row = append(row, k)
				row = append(row, "path")
				row = append(row, paramName)
				row = append(row, humanize.Comma(int64(v)))
				row = append(row, color.HiBlackString("100.00"))
				row = append(row, humanize.Comma(int64(len(vv))))
				row = append(row, printGini(gini, false))
				rows = append(rows, row)
			}
		}

		if hasQuery && (paramType == "query" || paramType == "all") {
			queryKeys := maps.Keys(queryParams)
			slices.Sort(queryKeys)
			for _, kk := range queryKeys {
				vv := queryParams[kk]
				gini := stat.Gini(maps.Values(vv), stat.Unsorted)
				row := make([]string, 0, 6)
				row = append(row, k)
				row = append(row, "query")
				row = append(row, fmt.Sprintf("?%s", kk))
				row = append(row, humanize.Comma(int64(result.QueryKey[k][kk])))
				row = append(row, fmt.Sprintf("%.2f", float64(result.QueryKey[k][kk])/float64(v)*100))
				row = append(row, humanize.Comma(int64(len(vv))))
				row = append(row, printGini(gini, false))
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
		cmd.PrintErrln("invalid format: %s", format)
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
