package sandmmo

import (
	"cmp"
	"encoding/binary"
	"fmt"
	"math"
	"sand-mmo/cell"
	"sand-mmo/common"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type orderList[t cmp.Ordered] []t

func (a *orderList[t]) SortedInsert(newValue t) {
	i, found := slices.BinarySearch(*a, newValue)
	if found {
		return
	}
	*a = slices.Insert(*a, i, newValue)
}

func (a *orderList[t]) GetReversSort() []t {
	var res []t = make([]t, len(*a))
	for i := range *a {
		res[len(*a)-i-1] = (*a)[i]
	}
	return res
}

func (a *orderList[t]) Clean() {

	*a = []t{}
}

type World struct {
	W            uint16
	H            uint16
	ChunkSize    uint16
	cells        []cell.Cell
	activeChunks orderList[uint16]
}

func NewWorld(w, h, chunkSize uint16) World {
	world := World{}
	world.cells = make([]cell.Cell, w*h)
	world.H = h
	world.W = w
	world.ChunkSize = chunkSize
	return world
}

func (w *World) SetCellsByte(bytes []byte, idChunk uint16) {
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

}

func (w *World) forEachCell(idChunk uint16, f func(x, y uint16, center *cell.Cell) error) error {

	chunkPerRow := w.W / w.ChunkSize
	chunkY := idChunk / chunkPerRow
	chunkX := idChunk % chunkPerRow
	x := chunkX*w.ChunkSize + w.ChunkSize - 1
	y := chunkY*w.ChunkSize + w.ChunkSize - 1
	for {
		if err := f(x, y, w.Get(x, y)); err != nil {
			return err
		}
		x = x - 1
		if x < chunkX*w.ChunkSize || x == math.MaxUint16 {
			x = chunkX*w.ChunkSize + w.ChunkSize - 1
			y = y - 1
		}
		if y < chunkY*w.ChunkSize || y == math.MaxUint16 {
			return nil
		}
	}
}

func (w *World) Simulate(idChunk uint16) error {

	isFree := func(pos common.Vec2) bool {
		x, y := pos.Get()
		cell := w.Get(uint16(x), uint16(y))

		return cell != nil && cell.IsEmpty()
	}
	simulateCustomMovements := func(pos common.Vec2, maxSpeed int32, cell **cell.Cell, isFree func(vec common.Vec2) bool, callbackAfter func(x, y int32) error, groups []common.Vec2) bool {
		oldx, oldy := pos.Get()
		move := func(v common.Vec2) {
			pos.Add(v)
			x, y := pos.Get()
			(*cell).Touched()
			w.Set(uint16(x), uint16(y), *(*cell))

			*cell = w.Get(uint16(x), uint16(y))
			callbackAfter(oldx, oldy)
		}

		for _, g := range groups {
			for s := maxSpeed; s > 0; s-- {
				o := g.Copy()
				o.MultConst(s)
				nPos := pos.Copy()
				nPos.Add(o)
				if isFree(nPos) {
					move(o)
					return true

				}
			}
		}
		return false
	}

	simulateSimpleMovements := func(pos common.Vec2, maxSpeed int32, c **cell.Cell, groups []common.Vec2) bool {
		removeOldCell := func(x, y int32) error {
			c, err := cell.NewCell(cell.EMPTY_CELL)
			if err != nil {
				return err
			}
			w.Set(uint16(x), uint16(y), c)
			return nil
		}
		return simulateCustomMovements(pos, maxSpeed, c, isFree, removeOldCell, groups)
	}

	simulateFireMovements := func(pos common.Vec2, maxSpeed int32, c **cell.Cell, groups []common.Vec2) bool {
		removeOldCell := func(x, y int32) error {
			cell := w.Get(uint16(x), uint16(y))
			cell.Touched()
			w.activeChunks.SortedInsert(idChunk)
			cell.DecreaseLife()
			return nil
		}
		isFree := func(pos common.Vec2) bool {
			x, y := pos.Get()
			tcell := w.Get(uint16(x), uint16(y))
			if tcell == nil {
				return false
			}
			if tcell.CellType == cell.WATER_CELL {
				(*c).RemainingLife = 0
			}
			if tcell.CellType == cell.WOOD_CELL && (*c).RemainingLife != 0 {
				(*c).RemainingLife = 3
				return true
			}
			return false
		}
		return simulateCustomMovements(pos, maxSpeed, c, isFree, removeOldCell, groups)
	}

	return w.forEachCell(idChunk, func(_x, _y uint16, center *cell.Cell) error {
		if center == nil {
			return nil
		}

		if center.IsNew() {

			w.activeChunks.SortedInsert(idChunk)
		}
		if center.IsEmpty() || center.IsTouched() {
			return nil
		}
		pos := common.NewVec2(int32(_x), int32(_y))
		switch center.CellType {
		case cell.SAND_CELL:
			simulateSimpleMovements(pos, 1, &center, []common.Vec2{
				common.NewVec2(0, 1),
				common.NewVec2(1, 1),
				common.NewVec2(-1, 1),
			})
		case cell.EMPTY_CELL:
			c, err := cell.NewCell(cell.EMPTY_CELL)
			if err != nil {
				return err
			}
			w.Set(_x, _y, c)
		case cell.WATER_CELL:
			simulateSimpleMovements(pos, 2, &center, []common.Vec2{
				common.NewVec2(0, 1),
				common.NewVec2(1, 1),
				common.NewVec2(-1, 1),
				common.NewVec2(-1, 0),
				common.NewVec2(1, 0),
			})
		case cell.SMOKE_CELL:
			if center.RemainingLife <= 0 {
				c, err := cell.NewCell(cell.EMPTY_CELL)
				if err != nil {
					return err
				}
				w.Set(_x, _y, c)
				return nil
			}
			center.DecreaseLife()
			moved := simulateSimpleMovements(pos, 2, &center, []common.Vec2{
				common.NewVec2(0, -1),
				common.NewVec2(1, -1),
				common.NewVec2(-1, -1),
				common.NewVec2(-1, 0),
				common.NewVec2(1, 0),
			})
			if !moved {
				center.Touched()
				w.activeChunks.SortedInsert(idChunk)
			}
		case cell.FIRE_CELL:
			moved := simulateFireMovements(pos, 1, &center, []common.Vec2{
				common.NewVec2(0, 1),
				common.NewVec2(0, -1),
				common.NewVec2(1, 0),
				common.NewVec2(-1, 0),
			})
			if center.RemainingLife <= 0 {
				c, err := cell.NewCell(cell.SMOKE_CELL)
				if err != nil {
					return err
				}
				w.Set(_x, _y, c)
				return nil
			}
			if !moved {
				center.Touched()
				w.activeChunks.SortedInsert(idChunk)
				center.DecreaseLife()
			}
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
		case cell.SAND_CELL:
			color = rl.Yellow
		case cell.WATER_CELL:
			color = rl.Blue
		case cell.SMOKE_CELL:
			color = rl.LightGray
		case cell.EMPTY_CELL:
			color = rl.SkyBlue
		case cell.STONE_CELL:
			color = rl.Gray
		case cell.FIRE_CELL:
			color = rl.Red
		case cell.WOOD_CELL:
			color = rl.Brown
		}
		rl.DrawRectangle(int32(x), int32(y), common.SIZE_CELL, common.SIZE_CELL, color)
		rl.DrawText(fmt.Sprint(y/common.SIZE_CELL), 0, int32(y), common.SIZE_CELL, rl.Black)
		i++
	}
}

// For test
func (w *World) importCell(cells []uint16) {
	w.cells = []cell.Cell{}
	for i := range cells {
		w.cells = append(w.cells, cell.DecodeCell(cells[i]))
	}
}

func (w *World) GetChunkId(x, y uint16) uint16 {
	chunkPerRow := w.W / w.ChunkSize
	id := (y/w.ChunkSize)*chunkPerRow + x/w.ChunkSize
	return id
}
func (w *World) GetNumberChucks() uint16 {
	return w.W / w.ChunkSize * w.H / w.ChunkSize
}

func (w *World) GetActiveChunksAndNeiboroud() (res orderList[uint16]) {
	chunks := w.activeChunks
	w.activeChunks.Clean()

	chunkPerRow := int(w.W / w.ChunkSize)
	totalChunks := chunkPerRow * int(w.H/w.ChunkSize)
	offsets := []int{
		0,
		-1, +1,
		-chunkPerRow, +chunkPerRow,
		-(chunkPerRow + 1), -(chunkPerRow - 1),
		+chunkPerRow - 1, +chunkPerRow + 1,
	}

	for _, c := range chunks {

		baseChunks := int(c)
		for _, off := range offsets {

			n := baseChunks + off

			if n < 0 || n >= totalChunks {
				continue
			}
			res.SortedInsert(uint16(n))
		}
	}
	return res.GetReversSort()
}

func (w *World) GetChunksToSend() []uint16 {
	return w.activeChunks
}

func (w *World) Set(x, y uint16, cell cell.Cell) {
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
	w.activeChunks.SortedInsert(w.GetChunkId(x, y))
	indexCell := x + (y * w.W)
	w.cells[indexCell] = cell
}

func (w *World) Get(x, y uint16) *cell.Cell {
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

func (w *World) GetChunk(idChunk uint16) []uint16 {
	var decoded []uint16

	chunkPerRow := w.W / w.ChunkSize

	chunkY := idChunk / chunkPerRow
	chunkX := idChunk % chunkPerRow

	iCell := chunkY*(w.W*w.ChunkSize) + chunkX*w.ChunkSize
	var i uint16
	for range uint16(w.ChunkSize) {
		i = 0
		for _, c := range w.cells[iCell : iCell+w.ChunkSize] {
			decoded = append(decoded, cell.EncodeCell(c))
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
		bytes = binary.BigEndian.AppendUint16(bytes, chunk[i])
	}
	return bytes
}
