package internal

import (
	"bytes"
	"sort"
	"testing"

	"github.com/haijima/stool/internal/log"
	"github.com/stretchr/testify/assert"
)

func TestTransitionProfiler_Profile(t *testing.T) {
	p := NewTransitionProfiler()
	stdin := bytes.NewBufferString("time:01/Jan/2023:12:00:00 +0900\thost:192.168.0.10\tforwardedfor:-\treq:POST /initialize HTTP/2.0\tstatus:200\tmethod:POST\turi:/initialize\tsize:18\treferer:-\tua:benchmarker-initializer\treqtime:0.268\tcache:-\truntime:-\tapptime:0.268\tvhost:192.168.0.11\tuidset:uid=0B00A8C0F528CA635B26685F02030303\tuidgot:-\tcookie:-\ntime:01/Jan/2023:12:00:05 +0900\thost:192.168.0.10\tforwardedfor:-\treq:GET / HTTP/2.0\tstatus:200\tmethod:GET\turi:/\tsize:528\treferer:-\tua:Mozilla/5.0 (X11; U; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36 Edg/85.0.564.44\treqtime:0.002\tcache:-\truntime:-\tapptime:0.000\tvhost:192.168.0.11\tuidset:uid=0B00A8C0FA28CA635B26685F02040303\tuidgot:-\tcookie:-\ntime:01/Jan/2023:12:00:06 +0900\thost:192.168.0.10\tforwardedfor:-\treq:GET / HTTP/2.0\tstatus:200\tmethod:GET\turi:/\tsize:528\treferer:-\tua:Mozilla/5.0 (X11; U; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36 Edg/85.0.564.44\treqtime:0.002\tcache:-\truntime:-\tapptime:0.000\tvhost:192.168.0.11\tuidset:-\tuidgot:uid=0B00A8C0FA28CA635B26685F02040303\tcookie:-\ntime:01/Jan/2023:12:00:07 +0900\thost:192.168.0.10\tforwardedfor:-\treq:GET / HTTP/2.0\tstatus:200\tmethod:GET\turi:/\tsize:528\treferer:-\tua:Mozilla/5.0 (X11; U; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36 Edg/85.0.564.44\treqtime:0.002\tcache:-\truntime:-\tapptime:0.000\tvhost:192.168.0.11\tuidset:uid=0B00A8C0FA28CA635C26725F02560303\tuidgot:-\tcookie:-\ntime:01/Jan/2023:12:00:07 +0900\thost:192.168.0.10\tforwardedfor:-\treq:GET / HTTP/2.0\tstatus:200\tmethod:GET\turi:/\tsize:528\treferer:-\tua:Mozilla/5.0 (X11; U; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36 Edg/85.0.564.44\treqtime:0.002\tcache:-\truntime:-\tapptime:0.000\tvhost:192.168.0.11\tuidset:uid=0B00A8C0FA28CA635B26685F02CA0303\tuidgot:-\tcookie:-\n")
	logReader, _ := log.NewLTSVReader(stdin, log.LTSVReadOpt{
		TimeFormat:     "02/Jan/2006:15:04:05 -0700",
		MatchingGroups: []string{"^/api/user/[^\\/]+$", "^/api/group/[^\\/]+$"},
	})

	transition, err := p.Profile(logReader)

	assert.NoError(t, err)
	assert.NotNil(t, transition)
	assert.Equal(t, 3, len(transition.Data))
	assert.Equal(t, 1, transition.Data[""]["POST /initialize"])
	assert.Equal(t, 3, transition.Data[""]["GET /"])
	assert.Equal(t, 1, transition.Data["POST /initialize"][""])
	assert.Equal(t, 1, transition.Data["GET /"]["GET /"])
	assert.Equal(t, 3, transition.Data["GET /"][""])
	assert.Equal(t, 2, len(transition.Sum))
	assert.Equal(t, 1, transition.Sum["POST /initialize"])
	assert.Equal(t, 4, transition.Sum["GET /"])
	sort.Strings(transition.Endpoints)
	assert.Equal(t, 3, len(transition.Endpoints))
	assert.Equal(t, "", transition.Endpoints[0])
	assert.Equal(t, "GET /", transition.Endpoints[1])
	assert.Equal(t, "POST /initialize", transition.Endpoints[2])

}
