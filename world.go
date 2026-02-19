package sandmmo

import ()

const SizeChunk = 2

type World struct {
	W            uint16
	H            uint16
	cells        []Cell
	activeChunks []struct {
		id    uint16
		cells []Cell
	}
}

func NewWorld(w, h uint16) World {
	world := World{}
	world.cells = make([]Cell, 0)
	world.H = h
	world.W = w
	return world
}

// For tests purpose
func (w *World) importCell(cells []uint32) {
	for i := range cells {
		w.cells = append(w.cells, DecodeCell(cells[i]))
	}
}
func (w *World) Set(x, y uint16, cell Cell) {
	indexCell := x * y
	w.cells[indexCell] = cell
}

func (w *World) Get(x, y uint16) Cell {
	indexCell := x * y
	return w.cells[indexCell]
}

func (w *World) GetChunk(idChunk uint16) []uint32 {
	var decoded []uint32

	chunkPerRow := w.W / SizeChunk

	chunkY := idChunk / chunkPerRow
	chunkX := idChunk % chunkPerRow

	iCell := chunkY*(w.W*SizeChunk) + chunkX*SizeChunk

	for range uint16(SizeChunk) {
		for _, cell := range w.cells[iCell : iCell+SizeChunk] {
			decoded = append(decoded, EncodeCell(cell))
		}
		iCell += w.W
	}

	return decoded
}
