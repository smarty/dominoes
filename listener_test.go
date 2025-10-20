package dominoes

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/smarty/gunit"
	"github.com/smarty/gunit/assert/should"
)

func TestListenerFixture(t *testing.T) {
	gunit.Run(new(ListenerFixture), t)
}

type ListenerFixture struct {
	*gunit.Fixture
}

func closeAfterDelay(l ListenCloser) {
	time.Sleep(time.Millisecond)
	_ = l.Close()
}

func (this *ListenerFixture) New(options ...option) ListenCloser {
	return New(append(options, Options.Logger(this))...)
}

func (this *ListenerFixture) TestPanicWhenProvidedNilListener() {
	fake := newFakeListener()
	this.So(func() { this.New(Options.AddListeners(fake, nil, fake)) }, should.Panic)
}
func (this *ListenerFixture) TestNoPanicWhenListenersAreOptional() {
	fake := newFakeListener()
	this.So(func() { this.New(Options.AddOptionalListeners(fake, nil, fake)) }, should.NotPanic)
}
func (this *ListenerFixture) TestNoListeners_FinishesWithoutPanic() {
	listener := this.New()
	go closeAfterDelay(listener)
	this.So(listener.Listen, should.NotPanic)
}
func (this *ListenerFixture) TestWhenProvidedListenerConcludes_BlockUntilClosedIsCalledExplicitly() {
	fake := newFakeListener()
	listener := this.New(Options.AddListeners(fake))

	go closeAfterDelay(listener)

	started := time.Now().UTC()
	listener.Listen()

	this.So(time.Since(started), should.BeGreaterThan, time.Millisecond)
	this.So(fake.closeTime, should.HappenAfter, fake.listenTime)
	this.So(fake.closeTime, should.HappenAfter, started.Add(time.Millisecond))
	this.So(fake.listenCount, should.Equal, 1)
	this.So(fake.closeCount, should.Equal, 1)
}
func (this *ListenerFixture) TestWhenMultipleListenersProvided_ItShouldCascadeCloseWhenHeadListenerIsClosed() {
	fakeListeners := []*fakeListener{newFakeListener(), newFakeListener(), newFakeListener()}
	listener := this.New(Options.AddListeners(fakeListeners[0], fakeListeners[1], fakeListeners[2]))

	go closeAfterDelay(listener)

	started := time.Now().UTC()
	listener.Listen()

	this.So(time.Since(started), should.BeGreaterThan, time.Millisecond)
	this.So([]time.Time{fakeListeners[0].closeTime, fakeListeners[1].closeTime, fakeListeners[2].closeTime},
		should.BeChronological)
}
func (this *ListenerFixture) TestWhenWatchingForOSSignals_ItShouldCloseProvidedListenerWhenSignalReceived() {
	fake := newFakeListener()
	listener := this.New(
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
	fake := newFakeListener()
	listener := this.New(
		Options.AddListeners(fake),
		Options.WatchTerminateSignals(),
	)

	go closeAfterDelay(listener)

	started := time.Now().UTC()
	listener.Listen()

	this.So(time.Since(started), should.BeGreaterThan, time.Millisecond)
	this.So(fake.closeTime, should.HappenAfter, fake.listenTime)
	this.So(fake.closeTime, should.HappenAfter, started.Add(time.Millisecond))
	this.So(fake.listenCount, should.Equal, 1)
	this.So(fake.closeCount, should.Equal, 1)
}
func (this *ListenerFixture) TestManagedResourcesClosedAlongsideListeners() {
	fake1 := newFakeListener()
	fake2 := newFakeListener()
	resource1 := newFakeResource()
	resource2 := newFakeResource()
	listener := this.New(
		Options.AddListeners(fake1, fake2),
		Options.AddManagedResource(resource1, resource2),
	)

	go closeAfterDelay(listener)

	listener.Listen()

	this.So(resource1.closeCount, should.Equal, 1)
	this.So(resource2.closeCount, should.Equal, 1)
}
func (this *ListenerFixture) TestContextCancellationAsManagedResource() {
	ctx, cancel := context.WithCancel(context.Background())
	listener := this.New(Options.AddContextShutdown(cancel))
	go closeAfterDelay(listener)
	listener.Listen()
	<-ctx.Done() // if cancel() was not called, this will block forever.
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type fakeListener struct {
	listenCount int
	listenTime  time.Time
	*fakeCloser
}

func newFakeListener() *fakeListener {
	return &fakeListener{fakeCloser: newFakeResource()}
}

func (this *fakeListener) Listen() {
	this.listenCount++
	this.listenTime = time.Now().UTC()
	time.Sleep(time.Microsecond * 10)
}

type fakeCloser struct {
	closeCount int
	closeTime  time.Time
	closeError error
}

func newFakeResource() *fakeCloser {
	return &fakeCloser{}
}

func (this *fakeCloser) Close() error {
	this.closeCount++
	this.closeTime = time.Now().UTC()
	time.Sleep(time.Microsecond * 10)
	return this.closeError
}
