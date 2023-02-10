package internal

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrendProfiler_Profile(t *testing.T) {
	p := NewTrendProfiler()
	stdin := bytes.NewBufferString("time:20/Jan/2023:14:39:01 +0900\thost:192.168.0.10\tforwardedfor:-\treq:POST /initialize HTTP/2.0\tstatus:200\tmethod:POST\turi:/initialize\tsize:18\treferer:-\tua:benchmarker-initializer\treqtime:0.268\tcache:-\truntime:-\tapptime:0.268\tvhost:192.168.0.11\tuidset:uid=0B00A8C0F528CA635B26685F02030303\tuidgot:-\tcookie:-\ntime:20/Jan/2023:14:39:06 +0900\thost:192.168.0.10\tforwardedfor:-\treq:GET / HTTP/2.0\tstatus:200\tmethod:GET\turi:/\tsize:528\treferer:-\tua:Mozilla/5.0 (X11; U; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36 Edg/85.0.564.44\treqtime:0.002\tcache:-\truntime:-\tapptime:0.000\tvhost:192.168.0.11\tuidset:uid=0B00A8C0FA28CA635B26685F02040303\tuidgot:-\tcookie:-\n")
	logReader, _ := NewLTSVReader(stdin, LTSVReadOpt{
		TimeFormat:     "02/Jan/2006:15:04:05 -0700",
		MatchingGroups: []string{"^/api/user/[^\\/]+$", "^/api/group/[^\\/]+$"},
	})

	trend, err := p.Profile(logReader, 5)

	assert.NoError(t, err)
	assert.NotNil(t, trend)
	assert.Equal(t, 2, len(trend.data))
	assert.Equal(t, 2, trend.Step)
}

func TestTrendProfiler_Profile_invalid_TimeFormat(t *testing.T) {
	p := NewTrendProfiler()
	stdin := bytes.NewBufferString("time:20/Jan/2023:14:39:01 +0900\thost:192.168.0.10\tforwardedfor:-\treq:POST /initialize HTTP/2.0\tstatus:200\tmethod:POST\turi:/initialize\tsize:18\treferer:-\tua:benchmarker-initializer\treqtime:0.268\tcache:-\truntime:-\tapptime:0.268\tvhost:192.168.0.11\tuidset:uid=0B00A8C0F528CA635B26685F02030303\tuidgot:-\tcookie:-\ntime:20/Jan/2023:14:39:06 +0900\thost:192.168.0.10\tforwardedfor:-\treq:GET / HTTP/2.0\tstatus:200\tmethod:GET\turi:/\tsize:528\treferer:-\tua:Mozilla/5.0 (X11; U; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36 Edg/85.0.564.44\treqtime:0.002\tcache:-\truntime:-\tapptime:0.000\tvhost:192.168.0.11\tuidset:uid=0B00A8C0FA28CA635B26685F02040303\tuidgot:-\tcookie:-\n")
	logReader, _ := NewLTSVReader(stdin, LTSVReadOpt{
		TimeFormat:     "02/01/2006:15:04:05 -0700", // invalid
		MatchingGroups: []string{"^/api/user/[^\\/]+$", "^/api/group/[^\\/]+$"},
	})

	trend, err := p.Profile(logReader, 5)

	assert.ErrorContains(t, err, "cannot parse")
	assert.Nil(t, trend)
}

func TestTrendProfiler_Profile_invalid_ltsv_format(t *testing.T) {
	p := NewTrendProfiler()
	// invalid ltsv format
	stdin := bytes.NewBufferString("time:20/Jan/2023:14:39:01 +0900\thost=192.168.0.10\tforwardedfor:-\treq:POST /initialize HTTP/2.0\tstatus:200\tmethod:POST\turi:/initialize\tsize:18\treferer:-\tua:benchmarker-initializer\treqtime:0.268\tcache:-\truntime:-\tapptime:0.268\tvhost:192.168.0.11\tuidset:uid=0B00A8C0F528CA635B26685F02030303\tuidgot:-\tcookie:-\n")
	logReader, _ := NewLTSVReader(stdin, LTSVReadOpt{
		TimeFormat:     "02/Jan/2006:15:04:05 -0700",
		MatchingGroups: []string{"^/api/user/[^\\/]+$"},
	})

	trend, err := p.Profile(logReader, 5)

	assert.ErrorContains(t, err, "bad line syntax")
	assert.Nil(t, trend)
}

func TestTrend(t *testing.T) {
	data := map[string]map[int]int{"GET /": {0: 1, 1: 2, 2: 3, 3: 4}, "POST /": {0: 1}}

	trend := NewTrend(data, 5, 5)

	assert.Equal(t, 5, trend.Interval)
	assert.Equal(t, 5, trend.Step)
}

func TestTrend_Counts(t *testing.T) {
	data := map[string]map[int]int{"GET /": {0: 1, 1: 2, 2: 3, 3: 4}, "POST /": {0: 1}}

	trend := NewTrend(data, 5, 5)

	assert.Equal(t, 5, len(trend.Counts("GET /")))
	assert.Equal(t, 1, trend.Counts("GET /")[0])
	assert.Equal(t, 2, trend.Counts("GET /")[1])
	assert.Equal(t, 3, trend.Counts("GET /")[2])
	assert.Equal(t, 4, trend.Counts("GET /")[3])
	assert.Equal(t, 0, trend.Counts("GET /")[4])
	assert.Equal(t, 5, len(trend.Counts("POST /")))
	assert.Equal(t, 1, trend.Counts("POST /")[0])
	assert.Equal(t, 0, trend.Counts("POST /")[1])
	assert.Equal(t, 0, trend.Counts("POST /")[2])
	assert.Equal(t, 0, trend.Counts("POST /")[3])
	assert.Equal(t, 0, trend.Counts("POST /")[4])
}

func TestTrend_Counts_not_found(t *testing.T) {
	data := map[string]map[int]int{"GET /": {0: 1, 1: 2, 2: 3, 3: 4}, "POST /": {0: 1}}

	trend := NewTrend(data, 5, 5)

	assert.NotNil(t, trend.Counts("PUT /"))
	assert.Equal(t, 0, len(trend.Counts("PUT /")))
}

func TestTrend_Endpoints(t *testing.T) {
	data := map[string]map[int]int{"GET /": {0: 1, 1: 2, 2: 3, 3: 4}, "POST /": {0: 1}}

	trend := NewTrend(data, 5, 5)

	assert.Equal(t, 2, len(trend.Endpoints()))
	assert.Equal(t, "GET /", trend.Endpoints()[0])
	assert.Equal(t, "POST /", trend.Endpoints()[1])
}
