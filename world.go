package sandmmo

import (
	"encoding/binary"
	"math"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const W_WINDOWS = 50
const H_WINDOWS = 50
const SIZE_CELL = 10
const CHUNK_SIZE = 25

type World struct {
	W            uint16
	H            uint16
	ChunkSize    uint16
	cells        []Cell
	activeChunks []uint8
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

func (w *World) forEachCell(idChunk uint16, f func(x, y uint16, center, up, down, right, left *Cell) error) error {

	chunkPerRow := w.W / w.ChunkSize
	chunkY := idChunk / chunkPerRow
	chunkX := idChunk % chunkPerRow
	x := chunkX*w.ChunkSize + w.ChunkSize
	y := chunkY*w.ChunkSize + w.ChunkSize
	for {
		if err := f(x, y, w.Get(x, y), w.Get(x, y-1), w.Get(x, y+1), w.Get(x+1, y), w.Get(x-1, y)); err != nil {
			return err
		}
		x--
		if x < chunkX*w.ChunkSize || x == math.MaxUint16 {
			x = chunkX*w.ChunkSize + w.ChunkSize
			y--
			if y < chunkY*w.ChunkSize || y == math.MaxUint16 {
				return nil
			}
		}
	}
}

func (w *World) Simulate(idChunk uint16) error {
	return w.forEachCell(idChunk, func(x, y uint16, center, up, down, right, left *Cell) error {
		if center == nil {
			return nil
		}
		if center.IsEmpty() || center.Touched {
			return nil
		}
		if down != nil && down.IsEmpty() {
			center.Touched = true
			w.Set(x, y+1, *center)
			w.Set(x, y, Cell{})
		}

		return nil
	})
}

func (w *World) Draw() {
	var i, x, y uint16
	var color rl.Color
	for range w.cells {
		x = i % w.W * SIZE_CELL
		y = i / w.W * SIZE_CELL
		if w.cells[i].Cell == 1 {
			color = rl.Black
		} else {
			color = rl.Beige
		}
		rl.DrawRectangle(int32(x), int32(y), SIZE_CELL, SIZE_CELL, color)
		i++
	}
}

// For test
func (w *World) importCell(cells []uint32) {
	w.cells = []Cell{}
	for i := range cells {
		w.cells = append(w.cells, DecodeCell(cells[i]))
	}
}

func (w *World) GetChunkId(x, y uint16) uint16 {
	chunkPerRow := w.W / w.ChunkSize
	return (y/w.ChunkSize)*chunkPerRow + x/w.ChunkSize
}
func (w *World) GetNumberChucks() uint16 {
	return w.W / w.ChunkSize * w.H / w.ChunkSize
}

func (w *World) GetTouchedChunks(reset bool) []uint8 {
	r := w.activeChunks
	r = slices.Compact(r)
	w.activeChunks = r
	if reset {
		w.activeChunks = []uint8{}
	}
	return r
}

func (w *World) Set(x, y uint16, cell Cell) {
	if x < 0 {
		return
	}
	if x >= w.W {
		return
	}
	if y < 0 {
		return
	}
	if y >= w.H {
		return
	}
	w.activeChunks = append(w.activeChunks, uint8(w.GetChunkId(x, y)))
	indexCell := x + (y * w.W)
	w.cells[indexCell] = cell
}

func (w *World) Get(x, y uint16) *Cell {
	if x < 0 {
		return nil
	}
	if x >= w.W {
		return nil
	}
	if y < 0 {
		return nil
	}
	if y >= w.H {
		return nil
	}
	return &w.cells[x+(y*w.W)]
}

func (w *World) GetChunk(idChunk uint16) []uint32 {
	var decoded []uint32

	chunkPerRow := w.W / w.ChunkSize

	chunkY := idChunk / chunkPerRow
	chunkX := idChunk % chunkPerRow

	iCell := chunkY*(w.W*w.ChunkSize) + chunkX*w.ChunkSize
	var i uint16
	for range uint16(w.ChunkSize) {
		i = 0
		for _, cell := range w.cells[iCell : iCell+w.ChunkSize] {
			w.cells[iCell+i].Touched = false
			decoded = append(decoded, EncodeCell(cell))
			i++
		}
		iCell += (w.W)
	}

	return decoded
}
func (w *World) GetChunkBytesToSend(idChunk uint16) []byte {
	chunk := w.GetChunk(idChunk)
	var bytes []byte
	bytes = binary.BigEndian.AppendUint16(bytes, idChunk)
	for i := range chunk {
		bytes = binary.BigEndian.AppendUint32(bytes, chunk[i])
	}
	return bytes
}
