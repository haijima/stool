package cmd

import (
	"fmt"
	"testing"

	"github.com/haijima/stool"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewTransitionCmd(t *testing.T) {
	p := stool.NewTransitionProfiler()
	v := viper.GetViper()
	fs := afero.NewOsFs()
	cmd := NewTransitionCmd(p, v, fs)

	assert.Equal(t, "transition", cmd.Name(), "NewTransitionCmd() should return command named \"transition\". but: \"%s\"", cmd.Name())
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
