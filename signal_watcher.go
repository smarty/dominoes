package dominoes

import (
	"context"
	"os"
	"os/signal"
)

type signalWatcher struct {
	channel  chan os.Signal
	ctx      context.Context
	shutdown context.CancelFunc

	inner   ListenCloser
	signals []os.Signal
	logger  logger
}

func newSignalWatcher(inner ListenCloser, config configuration) ListenCloser {
	if len(config.signals) == 0 {
		return inner
	}

	ctx, shutdown := context.WithCancel(context.Background())
	return signalWatcher{
		channel:  make(chan os.Signal, 4),
		ctx:      ctx,
		shutdown: shutdown,
		inner:    inner,
		logger:   config.logger,
		signals:  config.signals,
	}
}

func (this signalWatcher) Listen() {
	defer this.shutdown() // unblock wait method so we don't wait for signals anymore
	go this.wait()
	this.inner.Listen()
}
func (this signalWatcher) wait() {
	signal.Notify(this.channel, this.signals...)

	select {
	case value := <-this.channel:
		this.logger.Printf("[INFO] Received signal [%s]. Closing...", value)
	case <-this.ctx.Done():
	}

	signal.Stop(this.channel)
	close(this.channel)
	_ = this.inner.Close()
}

func (this signalWatcher) Close() error {
	this.shutdown()
	return nil
}
