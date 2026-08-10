package logger

import (
	"fmt"
	"io"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/ymhhh/go-common/config"
)

var (
	globalMu sync.RWMutex
	global   *Logger
)

// InitGlobal initializes the global logger from config.
// By default it reads the "logger" subtree; you may pass an optional subtree path.
// Existing entries returned by L keep using the same underlying logger after a
// reload, so they continue writing to the current output.
func InitGlobal(c config.Config, path ...string) error {
	l, err := FromConfig(c, path...)
	if err != nil {
		return err
	}

	var oldCloser io.Closer

	globalMu.Lock()
	old := global
	if old == nil {
		global = l
	} else {
		// Take ownership of the previous closer before installing the new one.
		// If Close already ran, closer is nil and must not be closed again
		// (rotatorr panics on double Close).
		oldCloser = old.closer
		old.closer = nil
		old.SetLevel(l.Level)
		old.SetReportCaller(l.ReportCaller)
		old.SetFormatter(l.Formatter)
		old.SetOutput(l.Out)
		old.closer = l.closer
		l.closer = nil
		old.closeOnce = sync.Once{}
	}
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
	return logrus.NewEntry(l.Logger)
}

// MustInitGlobal panics on init error.
func MustInitGlobal(c config.Config, path ...string) {
	if err := InitGlobal(c, path...); err != nil {
		panic(fmt.Errorf("logger: init global: %w", err))
	}
}
