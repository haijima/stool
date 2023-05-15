package genconf

import (
	"fmt"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenMatchingGroupFromEchoV4(t *testing.T) {
	filename := "testdata/echo_simple.go"
	mgs, err := GenMatchingGroupFromEchoV4(filename, "echo")

	assert.NoError(t, err)
	assert.Equal(t, len(mgs), 3)
	assert.Equal(t, mgs[0], "^/api/items$")
	assert.Equal(t, mgs[1], "^/api/users$")
	assert.Equal(t, mgs[2], "^/api/users/(?P<id>[^/]+)$")
}

func TestGenMatchingGroupFromEchoV4_complex(t *testing.T) {
	filename := "testdata/echo_complex.go"
	mgs, err := GenMatchingGroupFromEchoV4(filename, "echov4")

	var anblErr *ArgNotBasicLitError
	assert.ErrorAs(t, err, &anblErr)
	assert.Equal(t, len(anblErr.Info), 4)
	assert.Equal(t, fmt.Sprintf("%T", anblErr.Info[0].Call.Args[anblErr.Info[0].ArgIndex]), "*ast.BinaryExpr")
	assert.Equal(t, types.ExprString(anblErr.Info[0].Call.Fun), "e4.POST")
	assert.Equal(t, types.ExprString(anblErr.Info[0].Call.Args[0]), "api + \"/groups\"")
	assert.Equal(t, fmt.Sprintf("%T", anblErr.Info[1].Call.Args[anblErr.Info[1].ArgIndex]), "*ast.BinaryExpr")
	assert.Equal(t, types.ExprString(anblErr.Info[1].Call.Fun), "e4.GET")
	assert.Equal(t, types.ExprString(anblErr.Info[1].Call.Args[anblErr.Info[1].ArgIndex]), "\"/api/groups/:group_id\" + \"/users\"")
	assert.Equal(t, fmt.Sprintf("%T", anblErr.Info[2].Call.Args[anblErr.Info[2].ArgIndex]), "*ast.Ident")
	assert.Equal(t, types.ExprString(anblErr.Info[2].Call.Fun), "e4.GET")
	assert.Equal(t, types.ExprString(anblErr.Info[2].Call.Args[anblErr.Info[2].ArgIndex]), "index")
	assert.Equal(t, fmt.Sprintf("%T", anblErr.Info[3].Call.Args[anblErr.Info[3].ArgIndex]), "*ast.CallExpr")
	assert.Equal(t, types.ExprString(anblErr.Info[3].Call.Fun), "e4.GET")
	assert.Equal(t, types.ExprString(anblErr.Info[3].Call.Args[anblErr.Info[3].ArgIndex]), "fmt.Sprintf(\"/view/screen%d\", 1)")

	assert.Equal(t, len(mgs), 4)
	assert.Equal(t, mgs[0], "^/api/groups$")
	assert.Equal(t, mgs[1], "^/api/groups/(?P<group_id>[^/]+)/users/(?P<user_id>[^/]+)/tasks$")
	assert.Equal(t, mgs[2], "^/auth/login$")
	assert.Equal(t, mgs[3], "^/auth/logout$")
}
