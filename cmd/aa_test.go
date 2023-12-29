package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAaCommand(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewAaCmd(v, fs)

	assert.Equal(t, "aa", cmd.Name())
}

func TestNewAaCommand_Run(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewAaCmd(v, fs)
	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	cmd.Run(cmd, []string{})

	assert.Equal(t, aa, stdout.String())
}

func TestNewAaCommand_Flag_Big(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewAaCmd(v, fs)
	v.Set("big", true)
	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	cmd.Run(cmd, []string{})

	assert.Equal(t, aaBig, stdout.String())
}

func TestNewAaCommand_Flag_Text(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewAaCmd(v, fs)
	v.Set("text", true)
	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	cmd.Run(cmd, []string{})

	assert.Equal(t, aaText, stdout.String())
}
