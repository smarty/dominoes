package dominoes

type sequentialListener struct {
	listeners []Listener
}

func NewSequentialListener(listeners ...Listener) ListenCloser {
	return &sequentialListener{listeners: listeners}
}

func (this *sequentialListener) Listen() {
	for _, listener := range this.listeners {
		listener.Listen()
	}
}

func (this *sequentialListener) Close() error {
	var err error

	for _, listener := range this.listeners {
		if closer, ok := listener.(ListenCloser); ok {
			err = closer.Close()
		}
	}

	return err
}
