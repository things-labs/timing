package timing

// Option user's option
type Option func(tim *Base)

// WithLoggerProvider override default logger provider
func WithLoggerProvider(p LogProvider) Option {
	return func(tim *Base) {
		tim.setLogProvider(p)
	}
}

// WithEnableLogger enable logger
func WithEnableLogger() Option {
	return func(tim *Base) {
		tim.LogMode(true)
	}
}
