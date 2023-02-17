package log

import (
	"github.com/spf13/afero"
	"os"
	"testing"
)

func BenchmarkLTSVReader(b *testing.B) {
	fs := afero.NewOsFs()
	dir, _ := os.Getwd()
	fileName := dir + "/../../cmd/testdata/access.log"
	matchingGroups := []string{
		"^/$",
		"^/initialize$",
		"^/api/auth$",
		"^/api/condition/[^/]+$",
		"^/api/isu$",
		"^/api/isu/[^/]+$",
		"^/api/isu/[^/]+/graph$",
		"^/api/isu/[^/]+/icon$",
		"^/api/signout$",
		"^/api/trend$",
		"^/api/user/me$",
		"^/isu/[^/]+$",
		"^/isu/[^/]+/condition$",
		"^/isu/[^/]+/graph$",
		"^register$",
		"/assets/*",
	}

	for i := 0; i < b.N; i++ {
		f, _ := fs.Open(fileName)
		logReader, _ := NewLTSVReader(f, LTSVReadOpt{MatchingGroups: matchingGroups})
		for logReader.Read() {
			_, _ = logReader.Parse()
		}
		_ = f.Close()
	}
}
