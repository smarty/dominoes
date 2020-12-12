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
}

func newListener(config configuration) ListenCloser {
	if len(config.listeners) == 0 {
		return nil
	}

	current := config.listeners[0]
	if current == nil {
		panic("nil listener")
	}
	config.listeners = config.listeners[1:]
	next := newListener(config)
	ctx, shutdown := context.WithCancel(context.Background())
	return defaultListener{current: current, next: next, ctx: ctx, shutdown: shutdown}
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
	if listener, ok := listener.(io.Closer); ok {
		_ = listener.Close()
	}
}
