package dominoes

import (
	"io"
	"os"
	"syscall"
)

type configuration struct {
	listeners []Listener
	managed   []io.Closer
	signals   []os.Signal
	logger    logger
}

func New(options ...option) ListenCloser {
	var config configuration
	Options.apply(options...)(&config)
	return newSignalWatcher(newListener(config, 0), config)
}

var Options singleton

type singleton struct{}
type option func(*configuration)

func (singleton) AddOptionalListeners(value ...Listener) option {
	var populated []Listener

	for _, listener := range value {
		if listener != nil {
			populated = append(populated, listener)
		}
	}

	return Options.AddListeners(populated...)
}

func (singleton) AddListeners(value ...Listener) option {
	return func(this *configuration) { this.listeners = append(this.listeners, value...) }
}
func (singleton) AddManagedResource(value ...io.Closer) option {
	return func(this *configuration) { this.managed = append(this.managed, value...) }
}
func (singleton) WatchTerminateSignals() option {
	return Options.WatchSignals(syscall.SIGINT, syscall.SIGTERM)
}
func (singleton) WatchSignals(value ...os.Signal) option {
	return func(this *configuration) { this.signals = append(this.signals, value...) }
}
func (singleton) Logger(value logger) option {
	return func(this *configuration) { this.logger = value }
}

func (singleton) apply(options ...option) option {
	return func(this *configuration) {
		for _, option := range Options.defaults(options...) {
			option(this)
		}

		if len(this.listeners) == 0 {
			this.listeners = append(this.listeners, &nop{})
		}
	}
}
func (singleton) defaults(options ...option) []option {
	return append([]option{
		Options.Logger(&nop{}),
	}, options...)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type nop struct{}

func (*nop) Printf(_ string, _ ...interface{}) {}
func (*nop) Println(_ ...interface{})          {}

func (*nop) Listen() {}
