package cmd

import (
	"github.com/haijima/cobrax"
	"github.com/haijima/stool/internal"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// NewParamCommand returns the param command
func NewParamCommand(p *internal.ParamProfiler, v *viper.Viper, fs afero.Fs) *cobrax.Command {
	paramCmd := cobrax.NewCommand(v, fs)
	paramCmd.Use = "param"
	paramCmd.Short = "Show the parameters of each endpoint"
	paramCmd.RunE = func(cmd *cobrax.Command, args []string) error {
		return runParam(cmd, p, args)
	}

	return paramCmd
}

func runParam(cmd *cobrax.Command, p *internal.ParamProfiler, args []string) error {
	return nil
}
