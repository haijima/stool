package cmd

import (
	"bytes"
	"go/ast"
	"go/token"
	"testing"

	"github.com/haijima/stool/internal/genconf"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewGenConfCmd(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewGenConfCmd(v, fs)

	assert.Equal(t, "genconf", cmd.Name(), "NewGenConfCmd() should return command named \"genconf\". but: \"%s\"", cmd.Name())
}

func TestNewGenConfCmd_Flag(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewGenConfCmd(v, fs)
	formatFlag := cmd.Flags().Lookup("format")

	assert.True(t, cmd.HasAvailableFlags(), "genconf command should have available flag")
	assert.NotNil(t, formatFlag, "genconf command should have \"format\" flag")
	assert.Equal(t, "string", formatFlag.Value.Type(), "\"format\" flag is string")
}

func Test_printMatchingGroupInToml(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewGenConfCmd(v, fs)

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	err := printMatchingGroupInToml(cmd, MatchingGroupConf{MatchingGroups: []string{"foo", "bar"}})

	assert.NoError(t, err)
	assert.Equal(t, "matching_groups = [\n  'foo',\n  'bar'\n]\n\n", stdout.String())
}

func Test_printMatchingGroupInYaml(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewGenConfCmd(v, fs)

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	err := printMatchingGroupInYaml(cmd, MatchingGroupConf{MatchingGroups: []string{"foo", "bar"}})

	assert.NoError(t, err)
	assert.Equal(t, "matching_groups:\n    - foo\n    - bar\n\n", stdout.String())
}

func Test_printMatchingGroupInJson(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewGenConfCmd(v, fs)

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	err := printMatchingGroupInJson(cmd, MatchingGroupConf{MatchingGroups: []string{"foo", "bar"}})

	assert.NoError(t, err)
	assert.Equal(t, "{\n  \"matching_groups\": [\n    \"foo\",\n    \"bar\"\n  ]\n}\n", stdout.String())
}

func Test_printArgNotBasicLitError(t *testing.T) {
	v, fs := createViperAndFs()
	cmd := NewGenConfCmd(v, fs)

	stderr := new(bytes.Buffer)
	cmd.SetErr(stderr)

	printArgNotBasicLitError(
		cmd,
		&genconf.ArgNotBasicLitError{
			Info: []*genconf.ArgAstInfo{{
				Call: &ast.CallExpr{
					Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "foo"}, Sel: &ast.Ident{Name: "bar"}},
					Args: []ast.Expr{&ast.Ident{Name: "baz"}, &ast.Ident{Name: "qux"}},
				},
				CallPos:  token.Position{Filename: "test.go", Line: 1, Column: 1},
				ArgPos:   token.Position{Filename: "test.go", Line: 1, Column: 9},
				ArgIndex: 0,
			}},
		},
	)

	assert.Equal(t, "[Warning] test.go:1:9 Unable to parse *ast.Ident:\tfoo.bar(baz, qux)\n", stderr.String())
}

func TestRunGenConf(t *testing.T) {
	v := viper.New()
	fs := afero.NewOsFs()
	stdout := new(bytes.Buffer)
	cmd := NewRootCmd(v, fs)
	cmd.SetOut(stdout)
	cmd.SetArgs([]string{"genconf", "--format", "yaml", "../internal/genconf/testdata/echo_simple.go"})
	err := cmd.Execute()

	assert.NoError(t, err)
	assert.Equal(t, "matching_groups:\n    - ^/api/items$\n    - ^/api/users$\n    - ^/api/users/([^/]+)$\n\n", stdout.String())
}
