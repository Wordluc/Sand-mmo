package core

import (
	"encoding/binary"
	"sand-mmo/cell"
	"sand-mmo/common"
)

type ClientWorld struct {
	world
}

func NewCustomWorld(w, h, chunkSize int) (res ClientWorld) {
	res.world = newWorld(w, h, chunkSize)
	return res
}
func NewClientWorld() (res ClientWorld) {
	res.world = newWorld(common.W_CELLS_CLIENT, common.H_CELLS_CLIENT, common.CHUNK_SIZE)
	return res
}

func (w *ClientWorld) ShiftWorld(dx, dy int) {
	if dx == 0 && dy == 0 {
		return
	}
	if dx > 0 {
		for i := 0; i <= w.W*w.H-w.W; i += w.W {
			copy(w.cells[i:i+w.W-w.ChunkSize], w.cells[i+w.ChunkSize:i+w.W])
		}
	} else if dx < 0 {
		for i := 0; i <= w.W*w.H-w.W; i += w.W {
			copy(w.cells[i+w.ChunkSize:i+w.W], w.cells[i:i+w.W-w.ChunkSize])
		}
	}
	if dy > 0 {
		copy(w.cells[:w.W*w.H-w.ChunkSize*w.W], w.cells[w.ChunkSize*w.W:])
	} else if dy < 0 {
		copy(w.cells[w.ChunkSize*w.W:], w.cells[:w.W*w.H-w.ChunkSize*w.W])
	}
	r := make([]int, w.GetNumberChucks())
	for i := range r {
		r[i] = i
	}
	w.activeChunks.SortedInsert(r...)

}

func (w *ClientWorld) SetDecodedCells(bytes []byte, idChunk int) {
	const u32Size = 2
	var u16 uint16
	var c cell.Cell
	chunkPerRow := w.W / w.ChunkSize

	chunkY := idChunk / chunkPerRow
	chunkX := idChunk % chunkPerRow
	iCell := chunkY*(w.W*w.ChunkSize) + chunkX*w.ChunkSize
	for i := 0; i < len(bytes); i = i + u32Size {
		u16 = binary.BigEndian.Uint16(bytes[i : i+u32Size])
		c = cell.DecodeCell(u16)
		w.cells[iCell] = c
		iCell += 1
		if iCell%w.ChunkSize == 0 {
			iCell += (w.W - w.ChunkSize)
		}
	}
	w.activeChunks.SortedInsert(idChunk)
}

func (w *ClientWorld) PopActiveChunks() (res []int) {
	res = w.activeChunks.Get()
	w.activeChunks.Clean()
	return res
}
