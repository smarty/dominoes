package dominos

import "io"

type Listener interface {
	Listen()
}

type ListenCloser interface {
	Listener
	io.Closer
}

type logger interface {
	Printf(string, ...interface{})
}
