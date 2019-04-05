package sync

import (
	"testing"

	"github.com/oipwg/oip/datastore"
)

func TestBlockBuffer(t *testing.T) {
	bb := blockBuffer{}

	insCount := int64(600)

	if insCount < 40 {
		t.Errorf("below tests require at least 40 insertions; got %d", insCount)
	}

	for i := int64(1); i <= insCount; i++ {
		bb.Push(&datastore.BlockData{SecSinceLastBlock: i})
	}

	c1 := bb.Len()
	if c1 != bbCapacity {
		t.Errorf("unexpected count %d; expected %d", c1, bbCapacity)
	}

	b := bb.PeekFront()
	if b == nil || b.SecSinceLastBlock != insCount {
		t.Errorf("unexpected front block received (%v); expected SecSinceLastBlock=%d", b, insCount)
	}

	b = bb.PeekBack()
	if b == nil || b.SecSinceLastBlock != insCount-bbCapacity {
		t.Errorf("unexpected front block received (%v); expected SecSinceLastBlock=%d", b, insCount-bbCapacity)
	}

	for i := insCount; i > insCount-20; i-- {
		b = bb.PopFront()
		if b == nil || b.SecSinceLastBlock != i {
			t.Errorf("unexpected front block received (%v); expected SecSinceLastBlock=%d", b, insCount)
		}
	}

	for i := int64(0); i < 20; i++ {
		b = bb.PopBack()
		if b == nil || b.SecSinceLastBlock != insCount-bbCapacity+i {
			t.Errorf("unexpected front block received (%v); expected SecSinceLastBlock=%d", b, insCount-bbCapacity+i)
		}
	}

	c2 := bb.Len()
	if c2 != c1-40 {
		t.Errorf("unexpected count %d; expected %d", c2, c1-40)
	}
}
