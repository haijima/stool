package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/haijima/stool/internal"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

func BenchmarkScenarioCommand_RunE(b *testing.B) {
	p := internal.NewScenarioProfiler()
	v := viper.New()
	fs := afero.NewOsFs()
	cmd := NewScenarioCmd(p, v, fs)

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
