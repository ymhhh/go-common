package logger

import (
	"fmt"
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
// It closes the previous global logger if it owned a resource (file output).
func InitGlobal(c config.Config, path ...string) error {
	l, err := FromConfig(c, path...)
	if err != nil {
		return err
	}

	globalMu.Lock()
	old := global
	global = l
	globalMu.Unlock()

	if old != nil {
		_ = old.Close()
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
