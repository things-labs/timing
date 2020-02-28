package timing

import (
	"log"
	"os"
	"sync/atomic"
)

// LogProvider RFC5424 log message levels only Debug and Error
type LogProvider interface {
	Error(format string, v ...interface{})
	Debug(format string, v ...interface{})
}

// 内部调试实现
type logger struct {
	logger LogProvider
	// is log output enabled,1: enable, 0: disable
	hasLog uint32
}

// newLogger new logger with prefix
func newLogger(prefix string) *logger {
	return &logger{
		logger: defaultLogger{log.New(os.Stdout, prefix, log.LstdFlags)},
		hasLog: 0,
	}
}

// LogMode set enable or disable log output when you has set defaultLogger
func (sf *logger) LogMode(enable bool) {
	if enable {
		atomic.StoreUint32(&sf.hasLog, 1)
	} else {
		atomic.StoreUint32(&sf.hasLog, 0)
	}
}

// SetLogProvider set defaultLogger provider
func (sf *logger) setLogProvider(p LogProvider) {
	if p != nil {
		sf.logger = p
	}
}

// Error Log ERROR level message.
func (sf logger) Error(format string, v ...interface{}) {
	if atomic.LoadUint32(&sf.hasLog) == 1 {
		sf.logger.Error(format, v...)
	}
}

// Debug Log DEBUG level message.
func (sf logger) Debug(format string, v ...interface{}) {
	if atomic.LoadUint32(&sf.hasLog) == 1 {
		sf.logger.Debug(format, v...)
	}
}

// default log implement LogProvider
type defaultLogger struct {
	*log.Logger
}

// check implement LogProvider interface
var _ LogProvider = (*defaultLogger)(nil)

// Error Log ERROR level message.
func (sf defaultLogger) Error(format string, v ...interface{}) {
	sf.Printf("[E]: "+format, v...)
}

// Debug Log DEBUG level message.
func (sf defaultLogger) Debug(format string, v ...interface{}) {
	sf.Printf("[D]: "+format, v...)
}
