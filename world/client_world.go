package world

import "sand-mmo/cell"

type ClientWorld struct {
	world
}

func NewClientWorld(w, h, chunkSize uint16) (res ClientWorld) {
	res.world = newWorld(w, h, chunkSize)
	return res
}

func (w *ClientWorld) GetCells() []cell.Cell {
	return w.cells
}
