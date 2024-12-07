package internal

import (
	"bytes"
	"testing"

	"github.com/haijima/stool/internal/log"
	"github.com/stretchr/testify/assert"
)

func TestScenarioProfiler_Profile(t *testing.T) {
	p := NewScenarioProfiler()
	stdin := bytes.NewBufferString("time:01/Jan/2023:12:00:00 +0900\thost:192.168.0.10\tforwardedfor:-\treq:POST /initialize HTTP/2.0\tstatus:200\tmethod:POST\turi:/initialize\tsize:18\treferer:-\tua:benchmarker-initializer\treqtime:0.268\tcache:-\truntime:-\tapptime:0.268\tvhost:192.168.0.11\tuidset:uid=0B00A8C0F528CA635B26685F02030303\tuidgot:-\tcookie:-\ntime:01/Jan/2023:12:00:05 +0900\thost:192.168.0.10\tforwardedfor:-\treq:GET / HTTP/2.0\tstatus:200\tmethod:GET\turi:/\tsize:528\treferer:-\tua:Mozilla/5.0 (X11; U; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36 Edg/85.0.564.44\treqtime:0.002\tcache:-\truntime:-\tapptime:0.000\tvhost:192.168.0.11\tuidset:uid=0B00A8C0FA28CA635B26685F02040303\tuidgot:-\tcookie:-\ntime:01/Jan/2023:12:00:06 +0900\thost:192.168.0.10\tforwardedfor:-\treq:GET / HTTP/2.0\tstatus:200\tmethod:GET\turi:/\tsize:528\treferer:-\tua:Mozilla/5.0 (X11; U; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36 Edg/85.0.564.44\treqtime:0.002\tcache:-\truntime:-\tapptime:0.000\tvhost:192.168.0.11\tuidset:-\tuidgot:uid=0B00A8C0FA28CA635B26685F02040303\tcookie:-\ntime:01/Jan/2023:12:00:07 +0900\thost:192.168.0.10\tforwardedfor:-\treq:GET / HTTP/2.0\tstatus:200\tmethod:GET\turi:/\tsize:528\treferer:-\tua:Mozilla/5.0 (X11; U; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36 Edg/85.0.564.44\treqtime:0.002\tcache:-\truntime:-\tapptime:0.000\tvhost:192.168.0.11\tuidset:uid=0B00A8C0FA28CA635C26725F02560303\tuidgot:-\tcookie:-\ntime:01/Jan/2023:12:00:07 +0900\thost:192.168.0.10\tforwardedfor:-\treq:GET / HTTP/2.0\tstatus:200\tmethod:GET\turi:/\tsize:528\treferer:-\tua:Mozilla/5.0 (X11; U; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36 Edg/85.0.564.44\treqtime:0.002\tcache:-\truntime:-\tapptime:0.000\tvhost:192.168.0.11\tuidset:uid=0B00A8C0FA28CA635B26685F02CA0303\tuidgot:-\tcookie:-\n")
	logReader, _ := log.NewLTSVReader(stdin, log.LTSVReadOpt{
		TimeFormat:     "02/Jan/2006:15:04:05 -0700",
		MatchingGroups: []string{"^/api/user/[^\\/]+$", "^/api/group/[^\\/]+$"},
	})

	scenarios, err := p.Profile(logReader)

	assert.NoError(t, err)
	assert.NotNil(t, scenarios)
	assert.Equal(t, 2, len(scenarios))
	assert.Equal(t, "POST /initialize", scenarios[0].Hash)
	assert.Equal(t, 1, scenarios[0].Count)
	assert.Equal(t, 0, scenarios[0].FirstReq)
	assert.Equal(t, 0, scenarios[0].LastReq)
	assert.Equal(t, "POST /initialize", scenarios[0].Pattern.String(true))
	assert.Equal(t, "(GET /)*", scenarios[1].Hash)
	assert.Equal(t, 1, scenarios[1].Count)
	assert.Equal(t, 5000, scenarios[1].FirstReq)
	assert.Equal(t, 6000, scenarios[1].LastReq)
	assert.Equal(t, "(GET /)*", scenarios[1].Pattern.String(true))
}
