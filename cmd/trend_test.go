package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/haijima/stool/internal"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewTrendCmd(t *testing.T) {
	p := internal.NewTrendProfiler()
	v, fs := createViperAndFs()
	cmd := NewTrendCmd(p, v, fs)

	assert.Equal(t, "trend", cmd.Name(), "NewTrendCmd() should return command named \"trend\". but: \"%s\"", cmd.Name())
}

func TestNewTrendCmd_Flag(t *testing.T) {
	p := internal.NewTrendProfiler()
	v, fs := createViperAndFs()
	cmd := NewTrendCmd(p, v, fs)
	fileFlag := cmd.Flags().Lookup("file")
	matchingGroupsFlag := cmd.Flags().Lookup("matching_groups")
	timeFormatFlag := cmd.Flags().Lookup("time_format")
	logLabelsFlag := cmd.Flags().Lookup("log_labels")
	filterFlag := cmd.Flags().Lookup("filter")
	intervalFlag := cmd.LocalFlags().Lookup("interval")

	assert.NotNil(t, fileFlag, "trend command should have \"file\" flag")
	assert.Equal(t, "f", fileFlag.Shorthand, "\"file\" flag's shorthand is \"f\"")
	assert.Equal(t, "string", fileFlag.Value.Type(), "\"file\" flag is string")
	assert.NotNil(t, matchingGroupsFlag, "trend command should have \"matching_groups\" flag")
	assert.Equal(t, "m", matchingGroupsFlag.Shorthand, "\"matching_groups\" flag's shorthand is \"m\"")
	assert.Equal(t, "stringSlice", matchingGroupsFlag.Value.Type(), "\"matching_groups\" flag is string slice")
	assert.NotNil(t, timeFormatFlag, "trend command should have \"time_format\" flag")
	assert.Equal(t, "string", timeFormatFlag.Value.Type(), "\"time_format\" flag is string")
	assert.NotNil(t, logLabelsFlag, "trend command should have \"log_labels\" flag")
	assert.Equal(t, "stringToString", logLabelsFlag.Value.Type(), "\"log_labels\" flag is stringToString")
	assert.NotNil(t, filterFlag, "trend command should have \"filter\" flag")
	assert.Equal(t, "string", filterFlag.Value.Type(), "\"filter\" flag is string")
	assert.True(t, cmd.HasAvailableLocalFlags(), "trend command should have available local flag")
	assert.NotNil(t, intervalFlag, "trend command should have \"interval\" flag")
	assert.Equal(t, "i", intervalFlag.Shorthand, "\"interval\" flag's shorthand is \"i\"")
	assert.Equal(t, "int", intervalFlag.Value.Type(), "\"interval\" flag is int")
}

func Test_Trend_RunE(t *testing.T) {
	p := internal.NewTrendProfiler()
	v, fs := createViperAndFs()
	cmd := NewTrendCmd(p, v, fs)

	fileName := "./access.log"
	v.Set("file", fileName)
	v.Set("interval", "5")
	v.Set("format", "csv")
	_, _ = fs.Create(fileName)
	_ = afero.WriteFile(fs, fileName, []byte("time:20/Jan/2023:14:39:01 +0900\thost:192.168.0.10\tforwardedfor:-\treq:POST /initialize HTTP/2.0\tstatus:200\tmethod:POST\turi:/initialize\tsize:18\treferer:-\tua:benchmarker-initializer\treqtime:0.268\tcache:-\truntime:-\tapptime:0.268\tvhost:192.168.0.11\tuidset:uid=0B00A8C0F528CA635B26685F02030303\tuidgot:-\tcookie:-\ntime:20/Jan/2023:14:39:06 +0900\thost:192.168.0.10\tforwardedfor:-\treq:GET / HTTP/2.0\tstatus:200\tmethod:GET\turi:/\tsize:528\treferer:-\tua:Mozilla/5.0 (X11; U; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36 Edg/85.0.564.44\treqtime:0.002\tcache:-\truntime:-\tapptime:0.000\tvhost:192.168.0.11\tuidset:uid=0B00A8C0FA28CA635B26685F02040303\tuidgot:-\tcookie:-\ntime:20/Jan/2023:14:39:07 +0900\thost:192.168.0.10\tforwardedfor:-\treq:GET / HTTP/2.0\tstatus:200\tmethod:GET\turi:/\tsize:528\treferer:-\tua:Mozilla/5.0 (X11; U; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36 Edg/85.0.564.44\treqtime:0.002\tcache:-\truntime:-\tapptime:0.000\tvhost:192.168.0.11\tuidset:uid=0B00A8C0FA28CA635B26685F02040303\tuidgot:-\tcookie:-"), 0777)

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	err := cmd.RunE(cmd, []string{})

	assert.NoError(t, err)
	assert.Equal(t, "Method,Uri,0,5\nGET,/,0,2\nPOST,/initialize,1,0\n", stdout.String())
}

func Test_TrendCmd_RunE_Flag_interval_not_positive(t *testing.T) {
	p := internal.NewTrendProfiler()
	v, fs := createViperAndFs()
	cmd := NewTrendCmd(p, v, fs)

	v.Set("interval", "0")

	err := cmd.RunE(cmd, []string{})

	assert.ErrorContains(t, err, "interval flag should be positive. but: 0")
}

func Test_TrendCmd_RunE_file_not_exists(t *testing.T) {
	p := internal.NewTrendProfiler()
	v, fs := createViperAndFs()
	cmd := NewTrendCmd(p, v, fs)

	fileName := "./not_exists.log"
	v.Set("file", fileName)
	v.Set("format", "table")
	v.Set("interval", "5")

	err := cmd.RunE(cmd, []string{})

	assert.ErrorContains(t, err, "not_exists.log")
}

func Test_TrendCmd_RunE_file_profiler_error(t *testing.T) {
	p := internal.NewTrendProfiler()
	v, fs := createViperAndFs()
	cmd := NewTrendCmd(p, v, fs)

	fileName := "./access.log"
	v.Set("file", fileName)
	v.Set("format", "table")
	v.Set("time_format", "invalid")
	v.Set("interval", "5")
	_, _ = fs.Create(fileName)
	_ = afero.WriteFile(fs, fileName, []byte("time:20/Jan/2023:14:39:01 +0900\thost:192.168.0.10\tforwardedfor:-\treq:POST /initialize HTTP/2.0\tstatus:200\tmethod:POST\turi:/initialize\tsize:18\treferer:-\tua:benchmarker-initializer\treqtime:0.268\tcache:-\truntime:-\tapptime:0.268\tvhost:192.168.0.11\tuidset:uid=0B00A8C0F528CA635B26685F02030303\tuidgot:-\tcookie:-\ntime:20/Jan/2023:14:39:06 +0900\thost:192.168.0.10\tforwardedfor:-\treq:GET / HTTP/2.0\tstatus:200\tmethod:GET\turi:/\tsize:528\treferer:-\tua:Mozilla/5.0 (X11; U; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36 Edg/85.0.564.44\treqtime:0.002\tcache:-\truntime:-\tapptime:0.000\tvhost:192.168.0.11\tuidset:uid=0B00A8C0FA28CA635B26685F02040303\tuidgot:-\tcookie:-"), 0777)

	err := cmd.RunE(cmd, []string{})

	assert.ErrorContains(t, err, "parsing time")
}

func Test_printTrendCsv(t *testing.T) {
	data := make(map[string]*internal.TrendData, 2)
	data["GET /"] = &internal.TrendData{Method: "GET", Uri: "/"}
	data["GET /"].AddCount(0, 1)
	data["GET /"].AddCount(1, 2)
	data["GET /"].AddCount(2, 3)
	data["GET /"].AddCount(3, 4)
	data["GET /"].AddCount(4, 0)
	data["POST /"] = &internal.TrendData{Method: "POST", Uri: "/"}
	data["POST /"].AddCount(0, 1)
	data["POST /"].AddCount(1, 0)
	data["POST /"].AddCount(2, 0)
	data["POST /"].AddCount(3, 0)
	data["POST /"].AddCount(4, 0)

	//v := viper.New()
	//fs := afero.NewMemMapFs()
	cmd := &cobra.Command{}
	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	result := internal.NewTrend(data, 5, 5)
	_ = printTrendTable(cmd, result, "csv")

	assert.Equal(t, `Method,Uri,0,5,10,15,20
GET,/,1,2,3,4,0
POST,/,1,0,0,0,0
`, stdout.String())
}

func BenchmarkTrendCommand_RunE(b *testing.B) {
	p := internal.NewTrendProfiler()
	v := viper.New()
	fs := afero.NewOsFs()
	cmd := NewTrendCmd(p, v, fs)

	dir, _ := os.Getwd()
	fileName := dir + "/testdata/access.log"
	v.Set("file", fileName)
	v.Set("interval", "5")

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	for i := 0; i < b.N; i++ {
		_ = cmd.RunE(cmd, []string{})
	}
}

func TestTrendExecute(t *testing.T) {
	v := viper.New()
	fs := afero.NewOsFs()
	cmd := NewRootCmd(v, fs)
	cmd.SetArgs([]string{"trend"})

	assert.NoError(t, cmd.Execute())
}
