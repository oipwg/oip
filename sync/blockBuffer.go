package sync

import (
	"github.com/oipwg/oip/datastore"
)

const bbExp = 8 // 2e8 = 256 buckets, contain 255 recent blocks
const bbSize = 1 << bbExp
const bbCapacity = bbSize - 1

type blockBuffer struct {
	recentBlocks [bbSize]*datastore.BlockData
	front        int
	back         int
}

func (bb *blockBuffer) Push(b *datastore.BlockData) {
	bb.front = (bb.front + 1) & bbCapacity
	bb.recentBlocks[bb.front] = b
	if bb.front == bb.back {
		bb.back = (bb.back + 1) % bbCapacity
	}
}

func (bb *blockBuffer) PopFront() *datastore.BlockData {
	if bb.front == bb.back {
		return nil
	}

	b := bb.recentBlocks[bb.front]
	// decrement front index
	bb.front = (bb.front + bbCapacity) & bbCapacity

	return b
}

func (bb *blockBuffer) PopBack() *datastore.BlockData {
	if bb.front == bb.back {
		return nil
	}

	b := bb.recentBlocks[bb.back]
	// increment back index
	bb.back = (bb.back + 1) & bbCapacity

	return b
}

func (bb *blockBuffer) PeekFront() *datastore.BlockData {
	if bb.front == bb.back {
		return nil
	}

	return bb.recentBlocks[bb.front]
}

func (bb *blockBuffer) PeekBack() *datastore.BlockData {
	if bb.front == bb.back {
		return nil
	}

	return bb.recentBlocks[bb.back]
}

func (bb *blockBuffer) Len() int {
	return (bb.front + bbSize - bb.back) & bbCapacity
}

func (bb *blockBuffer) Get(i int) *datastore.BlockData {
	if i < 0 {
		return bb.recentBlocks[(bb.front+i+1)&bbCapacity]
	}
	return bb.recentBlocks[(bb.back+i)&bbCapacity]
}

func (bb *blockBuffer) Cap() int {
	return bbCapacity
}
