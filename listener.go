package dominoes

import (
	"context"
	"io"
)

type linkedListener struct {
	current  Listener
	next     Listener
	ctx      context.Context
	shutdown context.CancelFunc
	managed  []io.Closer
	logger   logger
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
	if len(config.listeners) > 0 {
		managed = nil
	}

	return &linkedListener{
		ctx:      ctx,
		shutdown: shutdown,
		current:  current,
		next:     newListener(config),
		managed:  managed,
		logger:   config.logger,
	}
}

func (this *linkedListener) Listen() {
	if this.next == nil {
		this.listen()
	} else {
		go this.listen()
		this.next.Listen()
	}
}
func (this *linkedListener) listen() {
	defer this.onListenComplete()
	this.current.Listen()
	<-this.ctx.Done()
	closeListener(this.next) // current just completed, now cause the next in line to conclude (if one exists)
}
func (this *linkedListener) onListenComplete() {
	if this.next != nil {
		return
	}

	CloseResources(this.managed...) // after the last/inner-most resource is closed, close these managed resources
	this.managed = this.managed[0:0]
	this.logger.Printf("[INFO] All listeners have concluded.")
}

func (this *linkedListener) Close() error {
	defer this.shutdown()
	closeListener(this.current)
	return nil
}
func closeListener(listener any) {
	if resource, ok := listener.(io.Closer); ok {
		CloseResources(resource)
	}
}

func CloseResources(resources ...io.Closer) {
	for _, resource := range resources {
		if resource != nil {
			_ = resource.Close()
		}
	}
}
