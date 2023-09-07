package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/haijima/stool/internal"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/tenntenn/golden"
)

func TestNewScenarioCmd(t *testing.T) {
	p := internal.NewScenarioProfiler()
	v := viper.New()
	fs := afero.NewMemMapFs()
	cmd := NewScenarioCmd(p, v, fs)
	_ = cmd.BindFlags()

	assert.Equal(t, "scenario", cmd.Name(), "NewScenarioCmd() should return command named \"scenario\". but: \"%s\"", cmd.Name())
}

func TestNewScenarioCmd_Flag(t *testing.T) {
	p := internal.NewScenarioProfiler()
	v := viper.New()
	fs := afero.NewMemMapFs()
	cmd := NewScenarioCmd(p, v, fs)
	_ = cmd.BindFlags()
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
	v := viper.New()
	fs := afero.NewMemMapFs()
	cmd := NewScenarioCmd(p, v, fs)
	_ = cmd.BindFlags()

	fileName := "./access.log"
	v.Set("file", fileName)
	_, _ = fs.Create(fileName)
	_ = afero.WriteFile(fs, fileName, []byte("time:01/Jan/2023:12:00:01 +0900\treq:POST /initialize HTTP/2.0\tstatus:200\tuidset:uid=0B00A8C0F528CA635B26685F02030303\tuidgot:-\ntime:01/Jan/2023:12:00:02 +0900\treq:GET / HTTP/2.0\tstatus:200\tuidset:uid=0B00A8C0FA28CA635B26685F02040303\tuidgot:-\ntime:01/Jan/2023:12:00:03 +0900\treq:GET / HTTP/2.0\tstatus:200\tuidset:-\tuidgot:uid=0B00A8C0FA28CA635B26685F02040303\n"), 0777)

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

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
	v := viper.New()
	fs := afero.NewMemMapFs()
	cmd := NewScenarioCmd(p, v, fs)
	_ = cmd.BindFlags()

	fileName := "./access.log"
	v.Set("file", fileName)
	v.Set("format", "csv")
	_, _ = fs.Create(fileName)
	_ = afero.WriteFile(fs, fileName, []byte("time:01/Jan/2023:12:00:01 +0900\treq:POST /initialize HTTP/2.0\tstatus:200\tuidset:uid=0B00A8C0F528CA635B26685F02030303\tuidgot:-\ntime:01/Jan/2023:12:00:02 +0900\treq:GET / HTTP/2.0\tstatus:200\tuidset:uid=0B00A8C0FA28CA635B26685F02040303\tuidgot:-\ntime:01/Jan/2023:12:00:03 +0900\treq:GET / HTTP/2.0\tstatus:200\tuidset:-\tuidgot:uid=0B00A8C0FA28CA635B26685F02040303\n"), 0777)

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	err := cmd.RunE(cmd, []string{})

	assert.NoError(t, err)
	assert.Equal(t, "first call[s],last call[s],count,scenario node\n0,0,1,POST /initialize\n1,2,1,(GET /)*\n", stdout.String())
}

func Test_ScenarioCmd_RunE_palette(t *testing.T) {
	p := internal.NewScenarioProfiler()
	v := viper.New()
	fs := afero.NewMemMapFs()
	cmd := NewScenarioCmd(p, v, fs)
	_ = cmd.BindFlags()

	fileName := "./access.log"
	v.Set("file", fileName)
	v.Set("palette", true)
	_, _ = fs.Create(fileName)
	_ = afero.WriteFile(fs, fileName, []byte("time:01/Jan/2023:12:00:01 +0900\treq:POST /initialize HTTP/2.0\tstatus:200\tuidset:uid=0B00A8C0F528CA635B26685F02030303\tuidgot:-\ntime:01/Jan/2023:12:00:02 +0900\treq:GET / HTTP/2.0\tstatus:200\tuidset:uid=0B00A8C0FA28CA635B26685F02040303\tuidgot:-\ntime:01/Jan/2023:12:00:03 +0900\treq:GET / HTTP/2.0\tstatus:200\tuidset:-\tuidgot:uid=0B00A8C0FA28CA635B26685F02040303\n"), 0777)

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	err := cmd.RunE(cmd, []string{})

	assert.NoError(t, err)
}

func TestScenarioExecute(t *testing.T) {
	testdata := filepath.Join("testdata", t.Name())
	formats := []string{"dot", "mermaid", "csv"}
	for _, tt := range formats {
		tt := tt
		t.Run(tt, func(t *testing.T) {
			v := viper.New()
			fs := afero.NewOsFs()
			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)
			cmd := NewRootCmd(v, fs)
			cmd.SetOut(stdout)
			cmd.SetErr(stderr)
			cmd.SetArgs([]string{"scenario", "--config", "testdata/.stool.yaml"})
			v.Set("format", tt)
			_ = cmd.BindFlags()

			assert.NoError(t, cmd.Execute())

			c := golden.New(t, flagUpdateGolden, testdata, tt)
			if diff := c.Check("_stdout", stdout); diff != "" {
				t.Error("stdout\n", diff)
			}
			if diff := c.Check("_stderr", stderr); diff != "" {
				t.Error("stderr\n", diff)
			}
		})
	}
}

func BenchmarkScenarioCommand_RunE(b *testing.B) {
	p := internal.NewScenarioProfiler()
	v := viper.New()
	fs := afero.NewOsFs()
	cmd := NewScenarioCmd(p, v, fs)
	_ = cmd.BindFlags()

	dir, _ := os.Getwd()
	fileName := dir + "/testdata/access.log"
	v.Set("file", fileName)

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	for i := 0; i < b.N; i++ {
		_ = cmd.RunE(cmd, []string{})
	}
}
