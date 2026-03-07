package world

import (
	"math"
	"sand-mmo/cell"
	"sand-mmo/common"
)

type ServerWorld struct {
	world
}

func NewServerWorld(w, h, chunkSize uint16) (res ServerWorld) {
	res.world = newWorld(w, h, chunkSize)
	return res
}

func (w *ServerWorld) forEachCell(idChunk uint16, f func(x, y uint16, center *cell.Cell) error) error {

	chunkPerRow := w.W / w.ChunkSize
	chunkY := idChunk / chunkPerRow
	chunkX := idChunk % chunkPerRow
	x := chunkX*w.ChunkSize + w.ChunkSize - 1
	y := chunkY*w.ChunkSize + w.ChunkSize - 1
	for {
		if err := f(x, y, w.Get(int32(x), int32(y))); err != nil {
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
func (w *ServerWorld) ApplyBrush(p common.BrushPackage) error {
	drawCircle := func(radius int) error {
		for iy := range radius * 2 {
			for ix := range radius * 2 {
				dx := (radius - ix)
				dy := (radius - iy)

				x := int(p.X) - dx
				if x < 0 {
					continue
				}
				y := int(p.Y) - dy
				if y < 0 {
					continue
				}
				if (dx*dx + dy*dy) <= radius*radius/4 {
					cell, err := cell.NewCell(p.CellType)
					if err != nil {
						return err
					}
					w.Set(uint16(x), uint16(y), cell)

				}
			}

		}
		return nil
	}
	drawBox := func(size int) error {
		for iy := range size * 2 {
			for ix := range size * 2 {
				dx := (size - ix)
				dy := (size - iy)

				x := int(p.X) - dx
				if x < 0 {
					continue
				}
				y := int(p.Y) - dy
				if y < 0 {
					continue
				}
				cell, err := cell.NewCell(p.CellType)
				if err != nil {
					return err
				}
				w.Set(uint16(x), uint16(y), cell)
			}
		}
		return nil
	}
	switch p.BrushType {
	case common.CIRCLE_SMALL:
		return drawCircle(4)
	case common.CIRCLE_BIG:
		return drawCircle(6)
	case common.SQUARE_SMALL:
		return drawBox(4)
	case common.SQUARE_BIG:
		return drawBox(6)
	}
	return nil
}

func (w *ServerWorld) AddGenerator(brush common.BrushPackage) {
	w.generators = append(w.generators, brush)
}

func (w *ServerWorld) ApplyGenerators() error {
	for i := range w.generators {
		err := w.ApplyBrush(w.generators[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *ServerWorld) Simulate(idChunk uint16) error {

	isFree := func(pos common.Vec2) bool {
		x, y := pos.Get()
		c := w.Get(x, y)

		return c != nil && (c.IsEmpty())
	}
	simulateCustomMovements := func(pos common.Vec2, maxSpeed int32, cell **cell.Cell, callbackBeforeMoving func(vec common.Vec2) bool, callbackAfterMoving func(x, y int32) error, groups []common.Vec2) bool {
		oldx, oldy := pos.Get()
		move := func(v common.Vec2) {
			pos.Add(v)
			x, y := pos.Get()
			(*cell).Touched()
			w.Set(uint16(x), uint16(y), *(*cell))

			*cell = w.Get(x, y)
			callbackAfterMoving(oldx, oldy)
		}

		for _, g := range groups {
			for s := maxSpeed; s > 0; s-- {
				o := g.Copy()
				o.MultConst(s)
				nPos := pos.Copy()
				nPos.Add(o)
				if callbackBeforeMoving(nPos) {
					move(o)
					return true

				}
			}
		}
		return false
	}

	simulateSimpleMovements := func(pos common.Vec2, maxSpeed int32, c **cell.Cell, groups []common.Vec2) bool {
		afterMoving := func(x, y int32) error {
			c, err := cell.NewCell(cell.EMPTY_CELL)
			if err != nil {
				return err
			}
			w.Set(uint16(x), uint16(y), c)
			return nil
		}
		return simulateCustomMovements(pos, maxSpeed, c, isFree, afterMoving, groups)
	}

	simulateWaterMovements := func(pos common.Vec2, maxSpeed int32, c **cell.Cell, groups []common.Vec2) bool {
		afterMoving := func(x, y int32) error {
			c, err := cell.NewCell(cell.EMPTY_CELL)
			if err != nil {
				return err
			}
			w.Set(uint16(x), uint16(y), c)
			return nil
		}
		beforeMoving := func(posToCheck common.Vec2) bool {
			x, y := posToCheck.Get()
			tcell := w.Get(x, y)
			if tcell == nil {
				return false
			}
			if tcell.CellType == cell.LAVA_CELL {
				smoke, _ := cell.NewCell(cell.SMOKE_CELL)
				w.SetVec(pos, smoke)
				return false
			}
			return isFree(posToCheck)
		}
		return simulateCustomMovements(pos, maxSpeed, c, beforeMoving, afterMoving, groups)
	}

	simulateFireMovements := func(pos common.Vec2, maxSpeed int32, c **cell.Cell, groups []common.Vec2) bool {
		afterMoving := func(x, y int32) error {
			cell := w.Get(x, y)
			cell.Touched()
			w.activeChunks.SortedInsert(idChunk)
			cell.DecreaseLife()
			return nil
		}
		isFree := func(pos common.Vec2) bool {
			x, y := pos.Get()
			tcell := w.Get(x, y)
			if tcell == nil {
				return false
			}
			if tcell.CellType == cell.WOOD_CELL && (*c).RemainingLife != 0 {
				(*c).RemainingLife = 3
				return true
			}
			return false
		}
		return simulateCustomMovements(pos, maxSpeed, c, isFree, afterMoving, groups)
	}

	simulateLavaMovements := func(pos common.Vec2, maxSpeed int32, c **cell.Cell, groups []common.Vec2) bool {
		afterMoving := func(x, y int32) error {
			c := w.Get(x, y)
			c.Touched()
			w.activeChunks.SortedInsert(idChunk)
			empty, _ := cell.NewCell(cell.EMPTY_CELL)
			w.Set(uint16(x), uint16(y), empty)
			return nil
		}
		isFree := func(pos common.Vec2) bool {
			x, y := pos.Get()
			tcell := w.Get(x, y)
			if tcell == nil {
				return false
			}
			if tcell.CellType == cell.WATER_CELL {
				smoke, _ := cell.NewCell(cell.SMOKE_CELL)
				w.SetVec(pos, smoke)
				return false
			}
			if tcell.CellType == cell.WOOD_CELL {
				fire, _ := cell.NewCell(cell.FIRE_CELL)
				w.Set(uint16(x), uint16(y), fire)
				return false
			}
			return isFree(pos)
		}
		return simulateCustomMovements(pos, maxSpeed, c, isFree, afterMoving, groups)
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
			simulateSimpleMovements(pos, 2, &center, []common.Vec2{
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

		case cell.LAVA_CELL:
			simulateLavaMovements(pos, 1, &center, []common.Vec2{
				common.NewVec2(0, 1),
				common.NewVec2(1, 1),
				common.NewVec2(-1, 1),
				common.NewVec2(-1, 0),
				common.NewVec2(1, 0),
			})
		case cell.WATER_CELL:
			simulateWaterMovements(pos, 2, &center, []common.Vec2{
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
			moved := simulateFireMovements(pos, 2, &center, []common.Vec2{
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
func (w *ServerWorld) GetActiveChunksAndNeiboroud() (res orderList[uint16]) {
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

func (w *ServerWorld) GetChunksToSend() []uint16 {
	return w.activeChunks
}
