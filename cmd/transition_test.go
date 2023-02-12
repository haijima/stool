package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/haijima/stool/internal"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewTransitionCmd(t *testing.T) {
	p := internal.NewTransitionProfiler()
	v := viper.New()
	fs := afero.NewMemMapFs()
	cmd := NewTransitionCmd(p, v, fs)

	assert.Equal(t, "transition", cmd.Name(), "NewTransitionCmd() should return command named \"transition\". but: \"%s\"", cmd.Name())
}

func TestNewTransitionCmd_Flag(t *testing.T) {
	p := internal.NewTransitionProfiler()
	v := viper.New()
	fs := afero.NewMemMapFs()
	cmd := NewTransitionCmd(p, v, fs)
	fileFlag := cmd.PersistentFlags().Lookup("file")
	matchingGroupsFlag := cmd.PersistentFlags().Lookup("matching_groups")
	ignorePatternsFlag := cmd.PersistentFlags().Lookup("ignore_patterns")
	timeFormatFlag := cmd.PersistentFlags().Lookup("time_format")
	formatFlag := cmd.PersistentFlags().Lookup("format")

	assert.True(t, cmd.HasAvailablePersistentFlags(), "transition command should have available flag")
	assert.NotNil(t, fileFlag, "transition command should have \"file\" flag")
	assert.Equal(t, "f", fileFlag.Shorthand, "\"file\" flag's shorthand is \"f\"")
	assert.Equal(t, "string", fileFlag.Value.Type(), "\"file\" flag is string")
	assert.NotNil(t, matchingGroupsFlag, "transition command should have \"matching_groups\" flag")
	assert.Equal(t, "m", matchingGroupsFlag.Shorthand, "\"matching_groups\" flag's shorthand is \"m\"")
	assert.Equal(t, "stringSlice", matchingGroupsFlag.Value.Type(), "\"matching_groups\" flag is string slice")
	assert.NotNil(t, ignorePatternsFlag, "transition command should have \"ignore_patterns\" flag")
	assert.Equal(t, "stringSlice", ignorePatternsFlag.Value.Type(), "\"ignore_patterns\" flag is string slice")
	assert.NotNil(t, timeFormatFlag, "transition command should have \"time_format\" flag")
	assert.Equal(t, "string", timeFormatFlag.Value.Type(), "\"time_format\" flag is string")
	assert.NotNil(t, formatFlag, "transition command should have \"format\" flag")
	assert.Equal(t, "string", formatFlag.Value.Type(), "\"format\" flag is string")
}

func Test_TransitionCmd_RunE(t *testing.T) {
	p := internal.NewTransitionProfiler()
	v := viper.New()
	fs := afero.NewMemMapFs()
	cmd := NewTransitionCmd(p, v, fs)

	fileName := "./access.log"
	v.Set("file", fileName)
	_, _ = fs.Create(fileName)
	_ = afero.WriteFile(fs, fileName, []byte("time:01/Jan/2023:12:00:01 +0900\treq:POST /initialize HTTP/2.0\tstatus:200\tuidset:uid=0B00A8C0F528CA635B26685F02030303\tuidgot:-\ntime:01/Jan/2023:12:00:02 +0900\treq:GET / HTTP/2.0\tstatus:200\tuidset:uid=0B00A8C0FA28CA635B26685F02040303\tuidgot:-\ntime:01/Jan/2023:12:00:03 +0900\treq:GET / HTTP/2.0\tstatus:200\tuidset:-\tuidgot:uid=0B00A8C0FA28CA635B26685F02040303\n"), 0777)

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	err := cmd.RunE(cmd, []string{})

	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "digraph")
	assert.Contains(t, stdout.String(), "start")
	assert.Contains(t, stdout.String(), "end")
	assert.Contains(t, stdout.String(), "\"POST /initialize\"")
	assert.Contains(t, stdout.String(), "\"GET /\"")
}

func Test_TransitionCmd_RunE_file_not_exists(t *testing.T) {
	p := internal.NewTransitionProfiler()
	v := viper.New()
	fs := afero.NewMemMapFs()
	cmd := NewTransitionCmd(p, v, fs)

	fileName := "./not_exists.log"
	v.Set("file", fileName)

	err := cmd.RunE(cmd, []string{})

	assert.ErrorContains(t, err, "not_exists.log")
}

func Test_TransitionCmd_RunE_format_csv(t *testing.T) {
	p := internal.NewTransitionProfiler()
	v := viper.New()
	fs := afero.NewMemMapFs()
	cmd := NewTransitionCmd(p, v, fs)

	fileName := "./access.log"
	v.Set("file", fileName)
	v.Set("format", "csv")
	_, _ = fs.Create(fileName)
	_ = afero.WriteFile(fs, fileName, []byte("time:01/Jan/2023:12:00:01 +0900\treq:POST /initialize HTTP/2.0\tstatus:200\tuidset:uid=0B00A8C0F528CA635B26685F02030303\tuidgot:-\ntime:01/Jan/2023:12:00:02 +0900\treq:GET / HTTP/2.0\tstatus:200\tuidset:uid=0B00A8C0FA28CA635B26685F02040303\tuidgot:-\ntime:01/Jan/2023:12:00:03 +0900\treq:GET / HTTP/2.0\tstatus:200\tuidset:-\tuidgot:uid=0B00A8C0FA28CA635B26685F02040303\n"), 0777)

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	err := cmd.RunE(cmd, []string{})

	assert.NoError(t, err)
	assert.Equal(t, ",,GET /,POST /initialize\n,0,1,1\nGET /,1,1,0\nPOST /initialize,1,0,0\n", stdout.String())
}

func Test_TransitionCmd_RunE_invalid_format(t *testing.T) {
	p := internal.NewTransitionProfiler()
	v := viper.New()
	fs := afero.NewMemMapFs()
	cmd := NewTransitionCmd(p, v, fs)

	fileName := "./access.log"
	v.Set("file", fileName)
	v.Set("format", "hoge")
	_, _ = fs.Create(fileName)
	_ = afero.WriteFile(fs, fileName, []byte("time:01/Jan/2023:12:00:01 +0900\treq:POST /initialize HTTP/2.0\tstatus:200\tuidset:uid=0B00A8C0F528CA635B26685F02030303\tuidgot:-\ntime:01/Jan/2023:12:00:02 +0900\treq:GET / HTTP/2.0\tstatus:200\tuidset:uid=0B00A8C0FA28CA635B26685F02040303\tuidgot:-\ntime:01/Jan/2023:12:00:03 +0900\treq:GET / HTTP/2.0\tstatus:200\tuidset:-\tuidgot:uid=0B00A8C0FA28CA635B26685F02040303\n"), 0777)

	err := cmd.RunE(cmd, []string{})

	assert.ErrorContains(t, err, "invalid format flag")
}

func Test_logNorm(t *testing.T) {
	type args struct {
		num    int
		src    int
		target int
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr assert.ErrorAssertionFunc
	}{{
		name:    "1000 -> 3 on [1, 1000000] -> [0, 6]",
		args:    args{num: 1000, src: 1000000, target: 6},
		want:    3,
		wantErr: nil,
	}, {
		name:    "32 -> 5 on [1, 1024] -> [0, 10]",
		args:    args{num: 32, src: 1024, target: 10},
		want:    5,
		wantErr: nil,
	}, {
		name:    "20 -> 1.30102999566 on [1, 1000000] -> [0, 6]",
		args:    args{num: 20, src: 1000000, target: 6},
		want:    1.30102999566,
		wantErr: nil,
	}, {
		name:    "1 -> 0 on [1, 1024] -> [0, 10]",
		args:    args{num: 1, src: 1024, target: 10},
		want:    0,
		wantErr: nil,
	}, {
		name: "error when num = 0",
		args: args{num: 0, src: 1024, target: 10},
		want: 0,
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return assert.ErrorContains(t, err, "num should be more than 0")
		},
	}, {
		name: "error when num = -1",
		args: args{num: -1, src: 1024, target: 10},
		want: 0,
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return assert.ErrorContains(t, err, "num should be more than 0")
		},
	}, {
		name: "error when src = 1",
		args: args{num: 1, src: 1, target: 0},
		want: 0,
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return assert.ErrorContains(t, err, "src should be more than 1")
		},
	}, {
		name: "error when target = 0",
		args: args{num: 1, src: 2, target: 0},
		want: 0,
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return assert.ErrorContains(t, err, "target should be more than 0")
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := logNorm(tt.args.num, tt.args.src, tt.args.target)
			if tt.wantErr != nil && !tt.wantErr(t, err, fmt.Sprintf("logNorm(%v, %v, %v)", tt.args.num, tt.args.src, tt.args.target)) {
				return
			}
			assert.InDeltaf(t, tt.want, got, 0.0000000001, "logNorm(%v, %v, %v)", tt.args.num, tt.args.src, tt.args.target)
		})
	}
}

func BenchmarkNewTransitionCommand_Trend_RunE(b *testing.B) {
	p := internal.NewTransitionProfiler()
	v := viper.New()
	fs := afero.NewMemMapFs()
	cmd := NewTransitionCmd(p, v, fs)

	fileName := "./access.log"
	v.Set("file", fileName)
	v.Set("interval", "5")
	_, _ = fs.Create(fileName)
	_ = afero.WriteFile(fs, fileName, []byte("time:20/Jan/2023:14:39:01 +0900\thost:192.168.0.10\tforwardedfor:-\treq:POST /initialize HTTP/2.0\tstatus:200\tmethod:POST\turi:/initialize\tsize:18\treferer:-\tua:benchmarker-initializer\treqtime:0.268\tcache:-\truntime:-\tapptime:0.268\tvhost:192.168.0.11\tuidset:uid=0B00A8C0F528CA635B26685F02030303\tuidgot:-\tcookie:-\ntime:20/Jan/2023:14:39:06 +0900\thost:192.168.0.10\tforwardedfor:-\treq:GET / HTTP/2.0\tstatus:200\tmethod:GET\turi:/\tsize:528\treferer:-\tua:Mozilla/5.0 (X11; U; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36 Edg/85.0.564.44\treqtime:0.002\tcache:-\truntime:-\tapptime:0.000\tvhost:192.168.0.11\tuidset:uid=0B00A8C0FA28CA635B26685F02040303\tuidgot:-\tcookie:-"), 0777)

	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)

	for i := 0; i < b.N; i++ {
		_ = cmd.RunE(cmd, []string{})
	}
}
