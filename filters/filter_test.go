package filters

import (
	"encoding/binary"
	"encoding/hex"
	"math"
	"math/rand"
	"sort"
	"testing"
)

func BenchmarkInsert(b *testing.B) {
	list := sort.IntSlice{}
	for i := 0; i < b.N; i++ {
		insert(&list, i)
	}
}

func TestFilter(t *testing.T) {
	caseCount := 100000
	max := math.MaxUint32
	step := max / caseCount
	Clear()

	testCases := make([]string, caseCount)
	for i := 0; i < caseCount; i++ {
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(i*step+rand.Intn(step)))
		s := hex.EncodeToString(b)
		testCases[i] = s
	}

	rand.Shuffle(len(testCases), func(i, j int) {
		testCases[i], testCases[j] = testCases[j], testCases[i]
	})

	for k, v := range testCases {
		if k%2 == 0 {
			Add(v)
		}
	}

	if len(filterList) != caseCount/2 {
		t.Error("incorrect filterList length")
	}

	for k, v := range testCases {
		if Contains(v) != (k%2 == 0) {
			t.Error("filter contents mismatch")
			return
		}
	}
}
