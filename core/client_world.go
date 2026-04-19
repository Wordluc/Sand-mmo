package core

import (
	"encoding/binary"
	"sand-mmo/common"
)

type ClientWorld struct {
	world
	left_chunks  []Cell
	right_chunks []Cell
	up_chunks    []Cell
	down_chunks  []Cell
	w_chunks     int
	h_chunks     int
}

func NewCustomWorld(w, h, chunkSize int) (res ClientWorld) {
	res.world = newWorld(w, h, chunkSize)
	res.left_chunks = make([]Cell, h*chunkSize*chunkSize)
	res.right_chunks = make([]Cell, h*chunkSize*chunkSize)
	res.up_chunks = make([]Cell, w*chunkSize*chunkSize)
	res.down_chunks = make([]Cell, w*chunkSize*chunkSize)
	res.w_chunks = w / chunkSize
	res.h_chunks = h / chunkSize
	return res
}

func NewClientWorld() (res ClientWorld) {
	return NewCustomWorld(common.W_CELLS_CLIENT, common.H_CELLS_CLIENT, common.CHUNK_SIZE)
}

func (w *ClientWorld) ShiftWorld(dx, dy int) {

	if dx > 0 {
		for row := 0; row < w.H; row++ {
			base := row * w.W
			copy(w.cells[base:base+w.W-w.ChunkSize], w.cells[base+w.ChunkSize:base+w.W])
			copy(w.cells[base+w.W-w.ChunkSize:base+w.W], w.right_chunks[row*w.ChunkSize:(row+1)*w.ChunkSize])
		}
	} else if dx < 0 {
		for row := 0; row < w.H; row++ {
			base := row * w.W
			copy(w.cells[base+w.ChunkSize:base+w.W], w.cells[base:base+w.W-w.ChunkSize])
			copy(w.cells[base:base+w.ChunkSize], w.left_chunks[row*w.ChunkSize:(row+1)*w.ChunkSize])
		}
	}

	if dy > 0 {
		copy(w.cells[:w.W*w.H-w.ChunkSize*w.W], w.cells[w.ChunkSize*w.W:])
		for col := 0; col < w.W/w.ChunkSize; col++ {
			src := col * w.ChunkSize * w.ChunkSize
			for row := 0; row < w.ChunkSize; row++ {
				base := (w.H-w.ChunkSize+row)*w.W + col*w.ChunkSize
				copy(w.cells[base:base+w.ChunkSize], w.down_chunks[src+row*w.ChunkSize:src+(row+1)*w.ChunkSize])
			}
		}
	} else if dy < 0 {
		copy(w.cells[w.ChunkSize*w.W:], w.cells[:w.W*w.H-w.ChunkSize*w.W])
		for col := 0; col < w.W/w.ChunkSize; col++ {
			src := col * w.ChunkSize * w.ChunkSize
			for row := 0; row < w.ChunkSize; row++ {
				base := row*w.W + col*w.ChunkSize
				copy(w.cells[base:base+w.ChunkSize], w.up_chunks[src+row*w.ChunkSize:src+(row+1)*w.ChunkSize])
			}
		}
	}
	r := make([]int, w.GetNumberChucks())
	for i := range r {
		r[i] = i
	}
	w.activeChunks.SortedInsert(r...)

}
func (w *ClientWorld) SetDecodedCells(bytes []byte, xChunk, yChunk int) {
	const u32Size = 2
	var u16 uint16
	var c Cell

	iCell := 0
	iBorder := 0
	isXOut := func(xChunk int) bool {
		return xChunk < 0 || xChunk >= w.w_chunks
	}

	isYOut := func(yChunk int) bool {
		return yChunk < 0 || yChunk >= w.h_chunks
	}

	for i := 0; i < len(bytes); i = i + u32Size {
		u16 = binary.BigEndian.Uint16(bytes[i : i+u32Size])
		c = DecodeCell(u16)
		if xChunk == -1 {
			if !isYOut(yChunk) {
				w.left_chunks[yChunk*w.ChunkSize*w.ChunkSize+iBorder] = c
				iBorder++
			}
			continue
		}
		if xChunk == common.W_CHUNKS_CLIENT {
			if !isYOut(yChunk) {
				w.right_chunks[yChunk*w.ChunkSize*w.ChunkSize+iBorder] = c
				iBorder++
			}
			continue
		}
		if yChunk == -1 {
			if !isXOut(xChunk) {
				w.up_chunks[xChunk*w.ChunkSize*w.ChunkSize+iBorder] = c
				iBorder++
			}
			continue
		}
		if yChunk == common.H_CHUNKS_CLIENT {
			if !isXOut(xChunk) {
				w.down_chunks[xChunk*w.ChunkSize*w.ChunkSize+iBorder] = c
				iBorder++
			}
			continue
		}
		if isXOut(xChunk) {
			return
		}
		if isYOut(yChunk) {
			return
		}
		w.cells[iCell+yChunk*(w.W*w.ChunkSize)+xChunk*w.ChunkSize] = c
		iCell += 1
		if iCell%w.ChunkSize == 0 {
			iCell += (w.W - w.ChunkSize)
		}
	}
	w.activeChunks.SortedInsert(xChunk + yChunk*w.w_chunks)
}

func (w *ClientWorld) PopActiveChunks() (res []int) {
	res = w.activeChunks.Get()
	w.activeChunks.Clean()
	return res
}
