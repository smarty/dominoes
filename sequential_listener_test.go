package dominoes

import (
	"errors"
	"testing"

	"github.com/smarty/gunit"
	"github.com/smarty/gunit/assert/should"
)

func TestSequentialListenerFixture(t *testing.T) {
	gunit.Run(new(SequentialListenerFixture), t)
}

type SequentialListenerFixture struct {
	*gunit.Fixture

	inner1 *fakeListener
	inner2 *fakeListener
	inner3 *fakeListener
	outer  ListenCloser
}

func (this *SequentialListenerFixture) Setup() {
	this.inner1 = newFakeListener()
	this.inner2 = newFakeListener()
	this.inner3 = newFakeListener()
	this.outer = NewSequentialListener(this.inner1, this.inner2, this.inner3)
}

func (this *SequentialListenerFixture) TestListen() {
	this.outer.Listen()
	this.So(this.inner1.listenCount, should.Equal, 1)
	this.So(this.inner2.listenCount, should.Equal, 1)
	this.So(this.inner3.listenCount, should.Equal, 1)
	this.So(this.inner1.listenTime, should.HappenBefore, this.inner2.listenTime)
	this.So(this.inner2.listenTime, should.HappenBefore, this.inner3.listenTime)
}
func (this *SequentialListenerFixture) TestClose_NoError() {
	err := this.outer.Close()
	this.So(err, should.BeNil)
	this.So(this.inner1.closeCount, should.Equal, 1)
	this.So(this.inner2.closeCount, should.Equal, 1)
	this.So(this.inner3.closeCount, should.Equal, 1)
	this.So(this.inner1.closeTime, should.HappenBefore, this.inner2.closeTime)
	this.So(this.inner2.closeTime, should.HappenBefore, this.inner3.closeTime)
}
func (this *SequentialListenerFixture) TestClose_LastErrorReturned() {
	err1 := errors.New("1")
	err3 := errors.New("3")
	this.inner1.closeError = err1
	this.inner3.closeError = err3
	err := this.outer.Close()
	this.So(err, should.Equal, err3)
}
