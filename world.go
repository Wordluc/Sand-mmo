package sandmmo

import (
	"encoding/binary"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const W_WINDOWS = 50
const H_WINDOWS = 50
const SIZE_CELL = 10

type World struct {
	W            uint16
	H            uint16
	ChunkSize    uint16
	cells        []Cell
	activeChunks []struct {
		id    uint16
		cells []Cell
	}
}

func NewWorld(w, h, chunkSize uint16) World {
	world := World{}
	world.cells = make([]Cell, w*h)
	world.H = h
	world.W = w
	world.ChunkSize = chunkSize
	return world
}

func (w *World) SetCellsByte(bytes []byte, idChunk uint16) {
	const u32Size = 4
	var u32 uint32
	var cell Cell
	chunkPerRow := w.W / w.ChunkSize

	chunkY := idChunk / chunkPerRow
	chunkX := idChunk % chunkPerRow
	iCell := chunkY*(w.W*w.ChunkSize) + chunkX*w.ChunkSize
	for i := 0; i < len(bytes); i = i + u32Size {
		u32 = binary.BigEndian.Uint32(bytes[i : i+u32Size])
		cell = DecodeCell(u32)
		w.cells[iCell] = cell
		iCell += 1
		if iCell%w.ChunkSize == 0 {
			iCell += (w.W - w.ChunkSize)
		}
	}

}

func (w *World) Draw() {
	var x int32
	var y int32
	var color rl.Color
	for i := range w.cells {
		x = int32(i%int(w.W)) * SIZE_CELL
		y = int32(i/int(w.W)) * SIZE_CELL
		if w.cells[i].Cell == 1 {
			color = rl.Black
		} else {
			color = rl.Beige
		}
		rl.DrawRectangle(x, y, SIZE_CELL, SIZE_CELL, color)
	}
}

// For test
func (w *World) importCell(cells []uint32) {
	w.cells = []Cell{}
	for i := range cells {
		w.cells = append(w.cells, DecodeCell(cells[i]))
	}
}

func (w *World) GetChuck(x, y uint16) uint16 {
	chunkPerRow := w.W / w.ChunkSize
	return (y/w.ChunkSize)*chunkPerRow + x/w.ChunkSize
}
func (w *World) Set(x, y uint16, cell Cell) {
	indexCell := x + (y * w.W)
	w.cells[indexCell] = cell
}

func (w *World) Get(x, y uint16) Cell {
	indexCell := x * y
	return w.cells[indexCell]
}

func (w *World) GetChunk(idChunk uint16) []uint32 {
	var decoded []uint32

	chunkPerRow := w.W / w.ChunkSize

	chunkY := idChunk / chunkPerRow
	chunkX := idChunk % chunkPerRow

	iCell := chunkY*(w.W*w.ChunkSize) + chunkX*w.ChunkSize

	for range uint16(w.ChunkSize) {
		for _, cell := range w.cells[iCell : iCell+w.ChunkSize] {
			decoded = append(decoded, EncodeCell(cell))
		}
		iCell += (w.W)
	}

	return decoded
}
