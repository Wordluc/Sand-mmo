package utils

import (
	"slices"
	"sync"
)

type Buffer struct {
	stacks         map[int]*Stack
	chunkAvailable []int
	*sync.Mutex
}

func NewBuffer() Buffer {
	return Buffer{Mutex: &sync.Mutex{}, stacks: map[int]*Stack{}}
}

func (b *Buffer) Clean() {
	b.stacks = map[int]*Stack{}
	b.chunkAvailable = make([]int, 0)
}

func (b *Buffer) Append(idChunk int, bytes []byte) {
	b.Lock()
	if _, ok := b.stacks[idChunk]; !ok {
		b.stacks[idChunk] = new(NewStack())
		b.chunkAvailable = append(b.chunkAvailable, idChunk)
	} else if b.stacks[idChunk].Len() == 0 {
		b.chunkAvailable = append(b.chunkAvailable, idChunk)
	}
	b.stacks[idChunk].Push(bytes)
	b.Unlock()
}

func (b *Buffer) GetLast(idChunk int) []byte {
	b.Lock()
	defer b.Unlock()
	if _, ok := b.stacks[idChunk]; !ok {
		return []byte{}
	}
	res, _ := b.stacks[idChunk].PopAndClean()
	return res
}

func (b *Buffer) GetChunks() (res []int) {
	b.Lock()
	res = slices.Clone(b.chunkAvailable)
	b.chunkAvailable = make([]int, 0)
	b.Unlock()
	return res
}
