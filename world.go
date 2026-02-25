package sandmmo

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand/v2"
	"sand-mmo/common"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type World struct {
	W            uint16
	H            uint16
	ChunkSize    uint16
	cells        []Cell
	activeChunks []uint8
}

var GTouchedId uint8 = 0

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

func (w *World) forEachCell(idChunk uint16, f func(x, y uint16, center *Cell) error) error {

	chunkPerRow := w.W / w.ChunkSize
	chunkY := idChunk / chunkPerRow
	chunkX := idChunk % chunkPerRow
	x := chunkX*w.ChunkSize + w.ChunkSize - 1
	y := chunkY*w.ChunkSize + w.ChunkSize - 1
	for {
		if err := f(x, y, w.Get(x, y)); err != nil {
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

	type coordinate struct {
		x int32
		y int32
	}
	isFree := func(x, y int32) bool {
		cell := w.Get(uint16(x), uint16(y))

		return cell != nil && cell.IsEmpty()
	}
	simulateMovements := func(x, y int32, cell **Cell, groups [][]coordinate) bool {
		for _, g := range groups {
			i := rand.IntN(len(g))
			o := g[i]
			if isFree(o.x+x, o.y+y) {
				(*cell).Touched()
				w.Set(uint16(o.x+x), uint16(o.y+y), *(*cell))
				*cell = w.Get(uint16(o.x+x), uint16(o.y+y))
				w.Set(uint16(x), uint16(y), Cell{})

				return true
			}
		}
		return false
	}
	return w.forEachCell(idChunk, func(_x, _y uint16, center *Cell) error {
		if center == nil {
			return nil
		}
		if center.IsEmpty() || center.IsTouched() {
			return nil
		}
		x := int32(_x)
		y := int32(_y)
		switch center.CellType {
		case SAND_CELL:
			simulateMovements(x, y, &center, [][]coordinate{
				{
					{x: 0, y: 1},
				}, {
					{x: 1, y: 1},
					{x: -1, y: 1},
				},
			})
		case WATER_CELL:
			simulateMovements(x, y, &center, [][]coordinate{
				{
					{x: 0, y: 1},
				}, {
					{x: 1, y: 1},
					{x: -1, y: 1},
				}, {
					{x: -1, y: 0},
					{x: 1, y: 0},
				},
			})
		case SMOKE_CELL:
			if center.RemainingLife <= 0 {
				w.Set(_x, _y, Cell{})
				return nil
			}
			simulateMovements(x, y, &center, [][]coordinate{
				{
					{x: 0, y: -1},
				}, {
					{x: 1, y: -1},
					{x: -1, y: -1},
				}, {
					{x: -1, y: 0},
					{x: 1, y: 0},
				},
			})
			center.RemainingLife -= 1
			center.Touched()
		}

		return nil
	})
}

func (w *World) Draw() {
	var i, x, y uint16
	var color rl.Color
	for range w.cells {
		x = i % w.W * common.SIZE_CELL
		y = i / w.W * common.SIZE_CELL
		switch w.cells[i].CellType {
		case SAND_CELL:
			color = rl.Yellow
		case WATER_CELL:
			color = rl.Blue
		case SMOKE_CELL:
			color = rl.LightGray
		case NULL_CELL:
			color = rl.SkyBlue
		}
		rl.DrawRectangle(int32(x), int32(y), common.SIZE_CELL, common.SIZE_CELL, color)
		rl.DrawText(fmt.Sprint(y/common.SIZE_CELL), 0, int32(y), common.SIZE_CELL, rl.Black)
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

func (w *World) GetActiveChunksAndNeiboroud() (res []uint8) {
	chunks := slices.Compact(w.activeChunks)
	w.activeChunks = []uint8{}

	chunkPerRow := int(w.W / w.ChunkSize)
	totalChunks := chunkPerRow * int(w.H/w.ChunkSize)

	offsets := []int{
		0,
		-1, +1,
		-chunkPerRow, +chunkPerRow,
		-chunkPerRow - 1, -chunkPerRow + 1,
		+chunkPerRow - 1, +chunkPerRow + 1,
	}

	for _, c := range chunks {

		baseChunks := int(c)

		for _, off := range offsets {

			n := baseChunks + off

			if n < 0 || n >= totalChunks {
				continue
			}
			res = append(res, uint8(n))
		}
	}
	slices.SortFunc(res, func(a, b uint8) int {
		return int(a) - int(b)

	})
	return slices.Compact(res)
}

func (w *World) GetChunksToSend() []uint8 {
	r := slices.Clone(w.activeChunks)
	slices.Sort(r)
	r = slices.Compact(r)

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
