package dominoes

import (
	"os"
	"testing"
	"time"

	"github.com/smartystreets/assertions/should"
	"github.com/smartystreets/gunit"
)

func TestListenerFixture(t *testing.T) {
	gunit.Run(new(ListenerFixture), t)
}

type ListenerFixture struct {
	*gunit.Fixture

	listener     ListenCloser
	listeners    []Listener
	watchSignals bool
}

func (this *ListenerFixture) initialize() {
	options := []option{
		Options.AddListeners(this.listeners...),
	}

	if this.watchSignals {
		options = append(options, Options.WatchTerminateSignals())
	}

	this.listener = New(options...)
}

func (this *ListenerFixture) TestPanicWhenProvidedNilListener() {
	fake := &fakeListener{}
	this.listeners = append(this.listeners, fake, nil, fake)
	this.So(this.initialize, should.Panic)
}

func (this *ListenerFixture) TestWhenProvidedListenerConcludes_BlockUntilClosedIsCalledExplicitly() {
	fake := &fakeListener{}
	this.listeners = append(this.listeners, fake)
	this.initialize()

	go func() {
		time.Sleep(time.Millisecond)
		_ = this.listener.Close()
	}()

	started := time.Now().UTC()
	this.listener.Listen()

	this.So(time.Since(started), should.BeGreaterThan, time.Millisecond)
	this.So(fake.closeTime, should.HappenAfter, fake.listenTime)
	this.So(fake.closeTime, should.HappenAfter, started.Add(time.Millisecond))
	this.So(fake.listenCount, should.Equal, 1)
	this.So(fake.closeCount, should.Equal, 1)
}

func (this *ListenerFixture) TestWhenMultipleListenersProvided_ItShouldCascadeCloseWhenHeadListenerIsClosed() {
	fakeListeners := []*fakeListener{{}, {}, {}}
	this.listeners = append(this.listeners, fakeListeners[0], fakeListeners[1], fakeListeners[2])
	this.initialize()

	go func() {
		time.Sleep(time.Millisecond)
		_ = this.listener.Close()
	}()

	started := time.Now().UTC()
	this.listener.Listen()

	this.So(time.Since(started), should.BeGreaterThan, time.Millisecond)
	this.So([]time.Time{fakeListeners[0].closeTime, fakeListeners[1].closeTime, fakeListeners[2].closeTime},
		should.BeChronological)
}

func (this *ListenerFixture) TestWhenWatchingForOSSignals_ItShouldCloseProvidedListenerWhenSignalReceived() {
	fake := &fakeListener{}
	this.listeners = append(this.listeners, fake)
	this.watchSignals = true
	this.initialize()

	go func() {
		time.Sleep(time.Millisecond)
		this.listener.(*signalWatcher).channel <- os.Interrupt
	}()

	started := time.Now().UTC()
	this.listener.Listen()

	this.So(time.Since(started), should.BeGreaterThan, time.Millisecond)
	this.So(fake.closeTime, should.HappenAfter, fake.listenTime)
	this.So(fake.closeTime, should.HappenAfter, started.Add(time.Millisecond))
	this.So(fake.listenCount, should.Equal, 1)
	this.So(fake.closeCount, should.Equal, 1)
}
func (this *ListenerFixture) TestWhenWatchingForOSSignals_ItShouldCloseProvidedListenerWhenCloseIsInvoked() {
	fake := &fakeListener{}
	this.listeners = append(this.listeners, fake)
	this.watchSignals = true
	this.initialize()

	go func() {
		time.Sleep(time.Millisecond)
		_ = this.listener.Close()
	}()

	started := time.Now().UTC()
	this.listener.Listen()

	this.So(time.Since(started), should.BeGreaterThan, time.Millisecond)
	this.So(fake.closeTime, should.HappenAfter, fake.listenTime)
	this.So(fake.closeTime, should.HappenAfter, started.Add(time.Millisecond))
	this.So(fake.listenCount, should.Equal, 1)
	this.So(fake.closeCount, should.Equal, 1)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type fakeListener struct {
	listenCount int
	listenTime  time.Time
	closeCount  int
	closeTime   time.Time
}

func (this *fakeListener) Listen() {
	this.listenCount++
	this.listenTime = time.Now().UTC()
}
func (this *fakeListener) Close() error {
	this.closeCount++
	this.closeTime = time.Now().UTC()
	return nil
}
