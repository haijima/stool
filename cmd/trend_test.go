package cmd

import (
	"bytes"
	"testing"

	"github.com/haijima/stool"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewTrendCmd(t *testing.T) {
	p := stool.NewTrendProfiler()
	v := viper.GetViper()
	fs := afero.NewOsFs()
	cmd := NewTrendCommand(p, v, fs)

	assert.Equal(t, "trend", cmd.Name(), "NewTrendCommand() should return command named \"trend\". but: \"%s\"", cmd.Name())
}

func TestNewTrendCmd_Flag(t *testing.T) {
	p := stool.NewTrendProfiler()
	v := viper.GetViper()
	fs := afero.NewOsFs()
	cmd := NewTrendCommand(p, v, fs)
	intervalFlag := cmd.LocalFlags().Lookup("interval")

	assert.True(t, cmd.HasAvailableLocalFlags(), "trend command should have available local flag")
	assert.NotNil(t, intervalFlag, "trend command should have \"interval\" flag")
	assert.Equal(t, "i", intervalFlag.Shorthand, "\"interval\" flag's shorthand is \"i\"")
	assert.Equal(t, "int", intervalFlag.Value.Type(), "\"interval\" flag is int")
}

func Test_RunE(t *testing.T) {
	p := stool.NewTrendProfiler()
	v := viper.GetViper()
	fs := afero.NewMemMapFs()
	cmd := NewTrendCommand(p, v, fs)

	fileName := "./access.log"
	v.Set("file", fileName)
	v.Set("matching_groups", "")
	v.Set("time_format", "02/Jan/2006:15:04:05 -0700")
	v.Set("interval", "5")
	_, _ = fs.Create(fileName)
	_ = afero.WriteFile(fs, fileName, []byte("time:20/Jan/2023:14:39:01 +0900\thost:192.168.0.10\tforwardedfor:-\treq:POST /initialize HTTP/2.0\tstatus:200\tmethod:POST\turi:/initialize\tsize:18\treferer:-\tua:benchmarker-initializer\treqtime:0.268\tcache:-\truntime:-\tapptime:0.268\tvhost:192.168.0.11\tuidset:uid=0B00A8C0F528CA635B26685F02030303\tuidgot:-\tcookie:-\ntime:20/Jan/2023:14:39:06 +0900\thost:192.168.0.10\tforwardedfor:-\treq:GET / HTTP/2.0\tstatus:200\tmethod:GET\turi:/\tsize:528\treferer:-\tua:Mozilla/5.0 (X11; U; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36 Edg/85.0.564.44\treqtime:0.002\tcache:-\truntime:-\tapptime:0.000\tvhost:192.168.0.11\tuidset:uid=0B00A8C0FA28CA635B26685F02040303\tuidgot:-\tcookie:-"), 0777)

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	err := cmd.RunE(cmd, []string{})

	assert.NoError(t, err)
	assert.Equal(t, "Method,Uri,0,5\nGET,/,0,1\nPOST,/initialize,1,0\n", stdout.String())
}

func Test_RunE_Flag_interval_not_positive(t *testing.T) {
	p := stool.NewTrendProfiler()
	v := viper.GetViper()
	fs := afero.NewMemMapFs()
	cmd := NewTrendCommand(p, v, fs)

	v.Set("interval", "0")

	err := cmd.RunE(cmd, []string{})

	assert.ErrorContains(t, err, "interval flag should be positive. but: 0")
}

func Test_RunE_file_not_exists(t *testing.T) {
	p := stool.NewTrendProfiler()
	v := viper.GetViper()
	fs := afero.NewMemMapFs()
	cmd := NewTrendCommand(p, v, fs)

	fileName := "./not_exists.log"
	v.Set("file", fileName)
	v.Set("interval", "5")

	err := cmd.RunE(cmd, []string{})

	assert.ErrorContains(t, err, "not_exists.log")
}

func Test_RunE_file_profiler_error(t *testing.T) {
	p := stool.NewTrendProfiler()
	v := viper.GetViper()
	fs := afero.NewMemMapFs()
	cmd := NewTrendCommand(p, v, fs)

	fileName := "./access.log"
	v.Set("file", fileName)
	v.Set("time_format", "invalid")
	_, _ = fs.Create(fileName)
	_ = afero.WriteFile(fs, fileName, []byte("time:20/Jan/2023:14:39:01 +0900\thost:192.168.0.10\tforwardedfor:-\treq:POST /initialize HTTP/2.0\tstatus:200\tmethod:POST\turi:/initialize\tsize:18\treferer:-\tua:benchmarker-initializer\treqtime:0.268\tcache:-\truntime:-\tapptime:0.268\tvhost:192.168.0.11\tuidset:uid=0B00A8C0F528CA635B26685F02030303\tuidgot:-\tcookie:-\ntime:20/Jan/2023:14:39:06 +0900\thost:192.168.0.10\tforwardedfor:-\treq:GET / HTTP/2.0\tstatus:200\tmethod:GET\turi:/\tsize:528\treferer:-\tua:Mozilla/5.0 (X11; U; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36 Edg/85.0.564.44\treqtime:0.002\tcache:-\truntime:-\tapptime:0.000\tvhost:192.168.0.11\tuidset:uid=0B00A8C0FA28CA635B26685F02040303\tuidgot:-\tcookie:-"), 0777)

	err := cmd.RunE(cmd, []string{})

	assert.ErrorContains(t, err, "parsing time")
}

func Test_printCsv(t *testing.T) {
	data := make(map[string]map[int]int, 2)
	data["GET /"] = make(map[int]int, 4)
	data["GET /"][0] = 1
	data["GET /"][1] = 2
	data["GET /"][2] = 3
	data["GET /"][3] = 4
	data["POST /"] = make(map[int]int, 1)
	data["POST /"][0] = 1

	p := stool.NewTrendProfiler()
	v := viper.GetViper()
	fs := afero.NewMemMapFs()
	cmd := NewTrendCommand(p, v, fs)

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	result := stool.NewTrend(data, 5, 5)
	printCsv(cmd, result)

	assert.Equal(t, `Method,Uri,0,5,10,15,20
GET,/,1,2,3,4,0
POST,/,1,0,0,0,0
`, stdout.String())
}
