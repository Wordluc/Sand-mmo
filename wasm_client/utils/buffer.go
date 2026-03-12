package utils

import (
	"slices"
	"sync"
)

type Buffer struct {
	stacks         map[uint16]*Stack
	chunkAvailable []uint16
	*sync.Mutex
}

func NewBuffer() Buffer {
	return Buffer{Mutex: &sync.Mutex{}, stacks: map[uint16]*Stack{}}
}

func (b *Buffer) Append(idChunk uint16, bytes []byte) {
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

func (b *Buffer) GetLast(idChunk uint16) []byte {
	b.Lock()
	defer b.Unlock()
	if _, ok := b.stacks[idChunk]; !ok {
		return []byte{}
	}
	res, _ := b.stacks[idChunk].PopAndClean()
	return res
}

func (b *Buffer) GetChunks() (res []uint16) {
	b.Lock()
	res = slices.Clone(b.chunkAvailable)
	b.chunkAvailable = make([]uint16, 0)
	b.Unlock()
	return res
}
