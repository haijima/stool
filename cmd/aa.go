package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewAaCommand returns the aa command
func NewAaCommand(v *viper.Viper, fs afero.Fs) *cobra.Command {
	// aaCmd represents the aa command
	var aaCmd = &cobra.Command{
		Use:   "aa",
		Short: "Show an ASCII art of a \"stool\"",
		Run: func(cmd *cobra.Command, args []string) {
			if v.GetBool("big") {
				cmd.Print(aaBig)
			} else if v.GetBool("text") {
				cmd.Print(aaText)
			} else {
				cmd.Print(aa)
			}
		},
		Hidden: true,
	}

	aaCmd.Flags().Bool("big", false, "if true, shows a big ASCII art")
	aaCmd.Flags().Bool("text", false, "if true, shows a ASCII art of text")
	aaCmd.MarkFlagsMutuallyExclusive("big", "text")

	_ = v.BindPFlags(aaCmd.Flags())
	v.SetFs(fs)

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
