package internal

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/fatih/color"
)

const TraceErrorKey = "trace_error"

type CliSlogHandler struct {
	mu   *sync.Mutex
	w    io.Writer
	opts slog.HandlerOptions
}

func NewCliSlogHandler(opts *slog.HandlerOptions) *CliSlogHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	if opts.Level == nil {
		opts.Level = slog.LevelError
	}
	return &CliSlogHandler{
		mu:   &sync.Mutex{},
		w:    os.Stderr,
		opts: *opts,
	}
}

func (h *CliSlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *CliSlogHandler) Handle(ctx context.Context, r slog.Record) error {
	var level string
	switch r.Level {
	case slog.LevelError:
		level = color.New(color.Bold, color.FgRed).Sprint("ERRO")
	case slog.LevelWarn:
		level = color.New(color.Bold, color.FgYellow).Sprint("WARN")
	case slog.LevelInfo:
		level = color.New(color.Bold, color.FgHiCyan).Sprint("INFO")
	case slog.LevelDebug:
		level = color.New(color.Bold, color.FgBlue).Sprint("DEBU")
	}

	var tracableErr error
	if r.Level == slog.LevelError {
		r.Attrs(func(attr slog.Attr) bool {
			if attr.Key == TraceErrorKey {
				if err, ok := attr.Value.Any().(error); ok {
					if _, _, _, ok = errors.GetOneLineSource(err); ok {
						tracableErr = err
						return false
					}
				}
			}
			return true
		})
	}

	buf := make([]byte, 0, 1024)
	buf = append(buf, r.Time.Format(time.TimeOnly)...)
	buf = append(buf, " "...)
	buf = append(buf, level...)
	buf = append(buf, "\t"...)
	buf = append(buf, r.Message...)
	if r.Level == slog.LevelError && tracableErr != nil {
		if h.opts.Level.Level() < slog.LevelDebug.Level() {
			buf = append(buf, "\n"...)
			buf = append(buf, fmt.Sprintf("%+v", tracableErr)...)
		}
	}

	if r.NumAttrs() > 0 {
		buf = append(buf, "\t"...)
	}

	r.Attrs(func(attr slog.Attr) bool {
		if attr.Key == TraceErrorKey {
			return true
		}
		buf = append(buf, " "...)
		buf = append(buf, color.HiBlackString(attr.Key)...)
		buf = append(buf, color.HiBlackString("=")...)
		buf = append(buf, attr.Value.String()...)

		return true
	})

	if h.opts.Level.Level() < slog.LevelDebug.Level() && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()

		file := f.File
		funcName := f.Function
		lineno := f.Line
		rst := errors.GetReportableStackTrace(tracableErr)
		if rst != nil && len(rst.Frames) > 0 {
			frame := rst.Frames[len(rst.Frames)-1]
			//file = frame.Filename
			file = frame.AbsPath
			funcName = frame.Function
			lineno = frame.Lineno
		}

		buf = append(buf, "\t"...)
		buf = append(buf, file...)
		buf = append(buf, ":"...)
		buf = append(buf, strconv.Itoa(lineno)...)
		buf = append(buf, " "...)
		buf = append(buf, funcName...)
	}

	buf = append(buf, "\n"...)

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(buf)
	return err
}

func (h *CliSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *CliSlogHandler) WithGroup(name string) slog.Handler {
	return h
}
