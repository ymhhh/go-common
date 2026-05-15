package logger

import (
	"fmt"
	"io"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/ymhhh/go-common/config"
)

var (
	globalMu     sync.RWMutex
	global       *logrus.Logger
	globalCloser io.Closer
)

// InitGlobal initializes the global logger from config.
// By default it reads the "logger" subtree; you may pass an optional subtree path.
// It closes the previous output after the shared global logger has been moved to
// the new output, so entries returned by L before a reload keep writing.
func InitGlobal(c config.Config, path ...string) error {
	l, err := FromConfig(c, path...)
	if err != nil {
		return err
	}

	var oldCloser io.Closer
	globalMu.Lock()
	if global == nil {
		global = l.Logger
	} else {
		applyLoggerConfig(global, l.Logger)
	}
	oldCloser = globalCloser
	globalCloser = l.closer
	globalMu.Unlock()

	if oldCloser != nil {
		_ = oldCloser.Close()
	}
	return nil
}

// L returns the global logger entry. If InitGlobal was never called, it returns
// logrus.StandardLogger().WithField("logger", "default").
func L() *logrus.Entry {
	globalMu.RLock()
	l := global
	globalMu.RUnlock()

	if l == nil {
		return logrus.StandardLogger().WithField("logger", "default")
	}
	return logrus.NewEntry(l)
}

// MustInitGlobal panics on init error.
func MustInitGlobal(c config.Config, path ...string) {
	if err := InitGlobal(c, path...); err != nil {
		panic(fmt.Errorf("logger: init global: %w", err))
	}
}

func applyLoggerConfig(dst, src *logrus.Logger) {
	dst.SetLevel(src.GetLevel())
	dst.SetReportCaller(src.ReportCaller)
	dst.SetFormatter(src.Formatter)
	dst.SetOutput(src.Out)
	dst.ReplaceHooks(src.Hooks)
}
