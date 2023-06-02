package dominoes

import "io"

type Listener interface {
	Listen()
}

type ListenCloser interface {
	Listener
	io.Closer
}

type logger interface {
	Printf(string, ...any)
}
