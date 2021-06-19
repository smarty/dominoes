package dominoes

import (
	"context"
	"io"
)

type defaultListener struct {
	current  Listener
	next     Listener
	ctx      context.Context
	shutdown context.CancelFunc
	managed  []io.Closer
}

func newListener(config configuration) ListenCloser {
	if len(config.listeners) == 0 {
		return nil
	}

	current := config.listeners[0]
	if current == nil {
		panic("nil listener")
	}

	ctx, shutdown := context.WithCancel(context.Background())
	config.listeners = config.listeners[1:]
	managed := config.managed
	config.managed = nil

	return defaultListener{
		ctx:      ctx,
		shutdown: shutdown,
		current:  current,
		next:     newListener(config),
		managed:  managed,
	}
}

func (this defaultListener) Listen() {
	if this.next == nil {
		this.listen()
	} else {
		go this.listen()
		this.next.Listen()
	}
}
func (this defaultListener) listen() {
	defer closeResources(this.managed...) // after the last/inner-most resource is closed, close these managed resources
	this.current.Listen()
	<-this.ctx.Done()
	closeListener(this.next) // current just completed, now cause the next in line to conclude (if one exists)
}

func (this defaultListener) Close() error {
	defer this.shutdown()
	closeListener(this.current)
	return nil
}
func closeListener(listener interface{}) {
	if resource, ok := listener.(io.Closer); ok {
		closeResources(resource)
	}
}
func closeResources(resources ...io.Closer) {
	for _, resource := range resources {
		if resource != nil {
			_ = resource.Close()
		}
	}
}
