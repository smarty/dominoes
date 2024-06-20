package dominoes

import (
	"os"
	"testing"
	"time"

	"github.com/smarty/assertions/should"
	"github.com/smarty/gunit"
)

func TestListenerFixture(t *testing.T) {
	gunit.Run(new(ListenerFixture), t)
}

type ListenerFixture struct {
	*gunit.Fixture
}

func (this *ListenerFixture) TestPanicWhenProvidedNilListener() {
	fake := &fakeListener{}
	this.So(func() { New(Options.AddListeners(fake, nil, fake)) }, should.Panic)
}

func (this *ListenerFixture) TestWhenProvidedListenerConcludes_BlockUntilClosedIsCalledExplicitly() {
	fake := &fakeListener{}
	listener := New(Options.AddListeners(fake))

	go func() {
		time.Sleep(time.Millisecond)
		_ = listener.Close()
	}()

	started := time.Now().UTC()
	listener.Listen()

	this.So(time.Since(started), should.BeGreaterThan, time.Millisecond)
	this.So(fake.closeTime, should.HappenAfter, fake.listenTime)
	this.So(fake.closeTime, should.HappenAfter, started.Add(time.Millisecond))
	this.So(fake.listenCount, should.Equal, 1)
	this.So(fake.closeCount, should.Equal, 1)
}

func (this *ListenerFixture) TestWhenMultipleListenersProvided_ItShouldCascadeCloseWhenHeadListenerIsClosed() {
	fakeListeners := []*fakeListener{{}, {}, {}}
	listener := New(Options.AddListeners(fakeListeners[0], fakeListeners[1], fakeListeners[2]))

	go func() {
		time.Sleep(time.Millisecond)
		_ = listener.Close()
	}()

	started := time.Now().UTC()
	listener.Listen()

	this.So(time.Since(started), should.BeGreaterThan, time.Millisecond)
	this.So([]time.Time{fakeListeners[0].closeTime, fakeListeners[1].closeTime, fakeListeners[2].closeTime},
		should.BeChronological)
}

func (this *ListenerFixture) TestWhenWatchingForOSSignals_ItShouldCloseProvidedListenerWhenSignalReceived() {
	fake := &fakeListener{}
	listener := New(
		Options.AddListeners(fake),
		Options.WatchTerminateSignals(),
	)

	go func() {
		time.Sleep(time.Millisecond)
		listener.(*signalWatcher).channel <- os.Interrupt
	}()

	started := time.Now().UTC()
	listener.Listen()

	this.So(time.Since(started), should.BeGreaterThan, time.Millisecond)
	this.So(fake.closeTime, should.HappenAfter, fake.listenTime)
	this.So(fake.closeTime, should.HappenAfter, started.Add(time.Millisecond))
	this.So(fake.listenCount, should.Equal, 1)
	this.So(fake.closeCount, should.Equal, 1)
}
func (this *ListenerFixture) TestWhenWatchingForOSSignals_ItShouldCloseProvidedListenerWhenCloseIsInvoked() {
	fake := &fakeListener{}
	listener := New(
		Options.AddListeners(fake),
		Options.WatchTerminateSignals(),
	)

	go func() {
		time.Sleep(time.Millisecond)
		_ = listener.Close()
	}()

	started := time.Now().UTC()
	listener.Listen()

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
	closeError  error
}

func (this *fakeListener) Listen() {
	this.listenCount++
	this.listenTime = time.Now().UTC()
	time.Sleep(time.Microsecond * 10)
}
func (this *fakeListener) Close() error {
	this.closeCount++
	this.closeTime = time.Now().UTC()
	time.Sleep(time.Microsecond * 10)
	return this.closeError
}
