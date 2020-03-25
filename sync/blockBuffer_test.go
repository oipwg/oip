package sync

import (
	"testing"

	"github.com/oipwg/oip/datastore"
)

func TestBlockBuffer(t *testing.T) {
	bb := blockBuffer{}

	insCount := int64(600)

	if insCount < 40 {
		t.Fatalf("below tests require at least 40 insertions; got %d", insCount)
	}

	if bb.Len() != 0 {
		t.Fatalf("unexpected length %d; expected %d", bb.Len(), 0)
	}

	for i := 1; i <= bbCapacity; i++ {
		bb.Push(&datastore.BlockData{SecSinceLastBlock: int64(i)})
		l := bb.Len()
		if l != i {
			t.Errorf("unexpected length %d; expected %d", l, i)
		}
	}

	for i := 1; i <= bbCapacity; i++ {
		b := bb.Get(i)
		if b == nil || b.SecSinceLastBlock != int64(i) {
			t.Errorf("unexpected positive index block received (%v); expected SecSinceLastBlock=%d", b, i)
		}
		b = bb.Get(-i)
		if b == nil || b.SecSinceLastBlock != int64(bbCapacity-i+1) {
			t.Errorf("unexpected negative index block received (%v); expected SecSinceLastBlock=%d", b, bbCapacity-i+1)
		}
	}

	for i := int64(bbCapacity + 1); i <= bbCapacity; i++ {
		bb.Push(&datastore.BlockData{SecSinceLastBlock: i})
	}

	for i := 1; i <= bbCapacity; i++ {
		b := bb.Get(i)
		if b == nil || b.SecSinceLastBlock != int64(i) {
			t.Errorf("unexpected positive index block received (%v); expected SecSinceLastBlock=%d", b, i)
		}
		b = bb.Get(-i)
		if b == nil || b.SecSinceLastBlock != int64(bbCapacity-i+1) {
			t.Errorf("unexpected negative index block received (%v); expected SecSinceLastBlock=%d", b, bbCapacity-i+1)
		}
	}

	for i := int64(bbCapacity + 1); i <= insCount; i++ {
		bb.Push(&datastore.BlockData{SecSinceLastBlock: i})
	}

	c1 := bb.Len()
	if c1 != bbCapacity {
		t.Errorf("unexpected count %d; expected %d", c1, bbCapacity)
	}

	b := bb.PeekFront()
	if b == nil || b.SecSinceLastBlock != insCount {
		t.Errorf("unexpected front block peeked (%v); expected SecSinceLastBlock=%d", b, insCount)
	}

	b = bb.PeekBack()
	if b == nil || b.SecSinceLastBlock != insCount-bbCapacity {
		t.Errorf("unexpected back block peeked (%v); expected SecSinceLastBlock=%d", b, insCount-bbCapacity)
	}

	for i := insCount; i > insCount-20; i-- {
		b = bb.PopFront()
		if b == nil || b.SecSinceLastBlock != i {
			t.Errorf("unexpected front block popped (%v); expected SecSinceLastBlock=%d", b, insCount-(insCount-i))
		}
	}

	for i := int64(0); i < 20; i++ {
		b = bb.PopBack()
		if b == nil || b.SecSinceLastBlock != insCount-bbCapacity+i {
			t.Errorf("unexpected back block popped (%v); expected SecSinceLastBlock=%d", b, insCount-bbCapacity+i)
		}
	}

	c2 := bb.Len()
	if c2 != c1-40 {
		t.Errorf("unexpected count %d; expected %d", c2, c1-40)
	}
}

func TestFront(t *testing.T) {
	bb := blockBuffer{}
	bb.Push(&datastore.BlockData{SecSinceLastBlock: 1})
	bb.Push(&datastore.BlockData{SecSinceLastBlock: 2})
	a := bb.PeekFront().SecSinceLastBlock // 2
	b := bb.PopFront().SecSinceLastBlock  // 2
	c := bb.PeekFront().SecSinceLastBlock // 1
	d := bb.PopFront().SecSinceLastBlock  // 1

	if a != b {
		t.Errorf("Value mismatch. %d != %d", a, b)
	}
	if b == c {
		t.Errorf("Values should not match. %d - %d", b, c)
	}
	if c != d {
		t.Errorf("Value mismatch. %d != %d", c, d)
	}
}
