package cmd

import (
	"github.com/haijima/cobrax"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// NewAaCommand returns the aa command
func NewAaCommand(v *viper.Viper, fs afero.Fs) *cobrax.Command {
	var aaCmd = cobrax.NewCommand(v, fs)
	aaCmd.Use = "aa"
	aaCmd.Short = "Show an ASCII art of a \"stool\""
	aaCmd.Hidden = true
	aaCmd.Run = func(cmd *cobrax.Command, args []string) {
		if cmd.Viper().GetBool("big") {
			cmd.PrintOut(aaBig)
		} else if cmd.Viper().GetBool("text") {
			cmd.PrintOut(aaText)
		} else {
			cmd.PrintOut(aa)
		}
	}

	aaCmd.Flags().Bool("big", false, "if true, shows a big ASCII art")
	aaCmd.Flags().Bool("text", false, "if true, shows a ASCII art of text")
	aaCmd.MarkFlagsMutuallyExclusive("big", "text")

	return aaCmd
}

const aa = `
   .JZZZZZZZZL.
 .&ZZZZZZZZZZZZ&.
 .&ZZZZZZZZZZZZ&.
  'YXZZZZZZZZXY'
   MN/ \NN/ \NM
  JMY   MM   YML
  MM'   MM   'MM
 JMY    MM    YML
 MM'    MM    'MM
        MM
`

const aaBig = `
        ,............,
   ..JZZZZZZZZZZZZZZZZZZL..
 .&ZZZZZZZZZZZZZZZZZZZZZZZZ&.
.JZZZZZZZZZZZZZZZZZZZZZZZZZZL.
!gZZZZZZZZZZZZZZZZZZZZZZZZZZg!
 ^YgggmmhhhQQQQQQQQhhhmmgggY^
    !NNoggggggggggggggoNN!
    .NMM/   \NNNN/   \MMN.
    [MNM     MMMM     MMM]
    JMMY     MMMM     YMML
   .MMM'     MMMM     'MMM.
   JMMY      MMMM      YMML
  .MMM       MMMM       MMM.
  NMM,       MMMM       ,MMN
 JMMM        MMMM        MMML
 MMMY        MMMM        YMMM
|MMM         MMMM         MMM|
@MM)         MMMM         (MM@
'""'         MMMM         '""'
             \MM/
`

const aaText = `
    _              _
   | |            | |
___| |_ ___   ___ | |
/ __| __/ _ \ / _ \| |
\__ \ || (_) | (_) | |
|___/\__\___/ \___/|_|
`
