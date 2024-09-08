package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/haijima/stool/internal"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewScenarioCmd(t *testing.T) {
	p := internal.NewScenarioProfiler()
	v, fs := createViperAndFs()
	cmd := NewScenarioCmd(p, v, fs)

	assert.Equal(t, "scenario", cmd.Name(), "NewScenarioCmd() should return command named \"scenario\". but: %q", cmd.Name())
}

func TestNewScenarioCmd_Flag(t *testing.T) {
	p := internal.NewScenarioProfiler()
	v, fs := createViperAndFs()
	cmd := NewScenarioCmd(p, v, fs)
	formatFlag := cmd.Flags().Lookup("format")
	paletteFlag := cmd.Flags().Lookup("palette")

	assert.True(t, cmd.HasAvailableFlags(), "scenario command should have available flag")
	assert.NotNil(t, formatFlag, "scenario command should have \"format\" flag")
	assert.Equal(t, "string", formatFlag.Value.Type(), "\"format\" flag is string")
	assert.NotNil(t, paletteFlag, "scenario command should have \"palette\" flag")
	assert.Equal(t, "bool", paletteFlag.Value.Type(), "\"palette\" flag is boolean")
}

func Test_ScenarioCmd_RunE(t *testing.T) {
	p := internal.NewScenarioProfiler()
	v, fs := createViperAndFs()
	cmd := NewScenarioCmd(p, v, fs)

	fileName := "./access.log"
	v.Set("file", fileName)
	_, _ = fs.Create(fileName)
	_ = afero.WriteFile(fs, fileName, []byte("time:01/Jan/2023:12:00:01 +0900\treq:POST /initialize HTTP/2.0\tstatus:200\tuidset:uid=0B00A8C0F528CA635B26685F02030303\tuidgot:-\ntime:01/Jan/2023:12:00:02 +0900\treq:GET / HTTP/2.0\tstatus:200\tuidset:uid=0B00A8C0FA28CA635B26685F02040303\tuidgot:-\ntime:01/Jan/2023:12:00:03 +0900\treq:GET / HTTP/2.0\tstatus:200\tuidset:-\tuidgot:uid=0B00A8C0FA28CA635B26685F02040303\n"), 0777)

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	_ = v.BindPFlags(cmd.Flags())
	err := cmd.RunE(cmd, []string{})

	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "digraph")
	assert.Contains(t, stdout.String(), "start")
	assert.Contains(t, stdout.String(), "end")
	assert.Contains(t, stdout.String(), "\"POST /initialize\"")
	assert.Contains(t, stdout.String(), "\"GET /\"")
}

func Test_ScenarioCmd_RunE_format_csv(t *testing.T) {
	p := internal.NewScenarioProfiler()
	v, fs := createViperAndFs()
	cmd := NewScenarioCmd(p, v, fs)

	fileName := "./access.log"
	v.Set("file", fileName)
	v.Set("format", "csv")
	_, _ = fs.Create(fileName)
	_ = afero.WriteFile(fs, fileName, []byte("time:01/Jan/2023:12:00:01 +0900\treq:POST /initialize HTTP/2.0\tstatus:200\tuidset:uid=0B00A8C0F528CA635B26685F02030303\tuidgot:-\ntime:01/Jan/2023:12:00:02 +0900\treq:GET / HTTP/2.0\tstatus:200\tuidset:uid=0B00A8C0FA28CA635B26685F02040303\tuidgot:-\ntime:01/Jan/2023:12:00:03 +0900\treq:GET / HTTP/2.0\tstatus:200\tuidset:-\tuidgot:uid=0B00A8C0FA28CA635B26685F02040303\n"), 0777)

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	_ = v.BindPFlags(cmd.Flags())
	err := cmd.RunE(cmd, []string{})

	assert.NoError(t, err)
	assert.Equal(t, "first call[s],last call[s],count,scenario node\n0,0,1,POST /initialize\n1,2,1,(GET /)*\n", stdout.String())
}

func Test_ScenarioCmd_RunE_palette(t *testing.T) {
	p := internal.NewScenarioProfiler()
	v, fs := createViperAndFs()
	cmd := NewScenarioCmd(p, v, fs)

	fileName := "./access.log"
	v.Set("file", fileName)
	v.Set("palette", true)
	_, _ = fs.Create(fileName)
	_ = afero.WriteFile(fs, fileName, []byte("time:01/Jan/2023:12:00:01 +0900\treq:POST /initialize HTTP/2.0\tstatus:200\tuidset:uid=0B00A8C0F528CA635B26685F02030303\tuidgot:-\ntime:01/Jan/2023:12:00:02 +0900\treq:GET / HTTP/2.0\tstatus:200\tuidset:uid=0B00A8C0FA28CA635B26685F02040303\tuidgot:-\ntime:01/Jan/2023:12:00:03 +0900\treq:GET / HTTP/2.0\tstatus:200\tuidset:-\tuidgot:uid=0B00A8C0FA28CA635B26685F02040303\n"), 0777)

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	_ = v.BindPFlags(cmd.Flags())
	err := cmd.RunE(cmd, []string{})

	assert.NoError(t, err)
}

func BenchmarkScenarioCommand_RunE(b *testing.B) {
	p := internal.NewScenarioProfiler()
	v := viper.New()
	fs := afero.NewOsFs()
	cmd := NewScenarioCmd(p, v, fs)

	dir, _ := os.Getwd()
	fileName := dir + "/testdata/access.log"
	v.Set("file", fileName)

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	for i := 0; i < b.N; i++ {
		_ = cmd.RunE(cmd, []string{})
	}
}
