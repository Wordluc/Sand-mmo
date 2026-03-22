package core

import (
	"sand-mmo/cell"
)

type ClientWorld struct {
	world
}

func NewClientWorld(w, h, chunkSize int) (res ClientWorld) {
	res.world = newWorld(w, h, chunkSize)
	return res
}

func (w *ClientWorld) GetCells() []cell.Cell {
	return w.cells
}

func (w *ClientWorld) ShiftWorld(dx, dy int) {
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
		clear(w.cells[w.W*w.H-w.ChunkSize*w.W:])
	} else if dy < 0 {
		copy(w.cells[w.ChunkSize*w.W:], w.cells[:w.W*w.H-w.ChunkSize*w.W])
		clear(w.cells[:w.ChunkSize*w.W])
	}
}
