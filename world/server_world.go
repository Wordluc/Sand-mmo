package world

import (
	"maps"
	"math"
	"sand-mmo/cell"
	"sand-mmo/common"
	"sync"

	ws "github.com/coder/websocket"
)

type ServerWorld struct {
	world
	webSockets     map[string]*ws.Conn
	webSocketMutex *sync.Mutex
}

func NewServerWorld(w, h, chunkSize uint16) (res ServerWorld) {
	res.world = newWorld(w, h, chunkSize)
	res.webSockets = map[string]*ws.Conn{}
	res.webSocketMutex = &sync.Mutex{}
	return res
}

func (w *ServerWorld) AddClient(addr string, conn *ws.Conn) int {
	w.webSocketMutex.Lock()
	defer w.webSocketMutex.Unlock()
	w.webSockets[addr] = conn
	return len(w.webSockets)
}

func (w *ServerWorld) RemoveClient(addr string) {
	w.webSocketMutex.Lock()
	defer w.webSocketMutex.Unlock()
	delete(w.webSockets, addr)
}

func (w *ServerWorld) GetLenSockets() int {
	w.webSocketMutex.Lock()
	defer w.webSocketMutex.Unlock()
	return len(w.webSockets)
}

func (w *ServerWorld) GetClients() (conns map[string]*ws.Conn) {
	w.webSocketMutex.Lock()
	conns = maps.Clone(w.webSockets)
	w.webSocketMutex.Unlock()
	return conns
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
					w.Set(uint16(x), uint16(y), cell.NewCell(p.CellType))

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
				w.Set(uint16(x), uint16(y), cell.NewCell(p.CellType))
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
			w.Set(uint16(x), uint16(y), cell.NewCell(cell.EMPTY_CELL))
			return nil
		}
		return simulateCustomMovements(pos, maxSpeed, c, isFree, afterMoving, groups)
	}

	simulateWaterMovements := func(pos common.Vec2, maxSpeed int32, c **cell.Cell, groups []common.Vec2) bool {
		afterMoving := func(x, y int32) error {
			w.Set(uint16(x), uint16(y), cell.NewCell(cell.EMPTY_CELL))
			return nil
		}
		beforeMoving := func(posToCheck common.Vec2) bool {
			x, y := posToCheck.Get()
			tcell := w.Get(x, y)
			if tcell == nil {
				return false
			}
			if tcell.CellType == cell.LAVA_CELL || tcell.CellType == cell.FIRE_CELL {
				smoke, isSmoke := NewCellByChance(cell.SMOKE_CELL, 1)
				if !isSmoke {
					tcell.RemainingLife = 0
					return false
				}
				w.SetVec(pos, smoke)
				return false
			}
			return isFree(posToCheck)
		}
		return simulateCustomMovements(pos, maxSpeed, c, beforeMoving, afterMoving, groups)
	}
	simulateLeafMovements := func(pos common.Vec2, maxSpeed int32, c **cell.Cell, groups []common.Vec2) bool {
		afterMoving := func(x, y int32) error {
			w.Set(uint16(x), uint16(y), cell.NewCell(cell.EMPTY_CELL))
			return nil
		}
		beforeMoving := func(posToCheck common.Vec2) bool {
			x, y := posToCheck.Get()
			tcell := w.Get(x, y)
			if tcell == nil {
				return false
			}
			if tcell.CellType == cell.LAVA_CELL || tcell.CellType == cell.FIRE_CELL {
				fire := cell.NewCell(cell.FIRE_CELL)
				w.SetVec(pos, fire)
				return false
			}
			return isFree(posToCheck)
		}
		return simulateCustomMovements(pos, maxSpeed, c, beforeMoving, afterMoving, groups)
	}

	simulateVacummMovements := func(pos common.Vec2, maxSpeed int32, c **cell.Cell, groups []common.Vec2) bool {
		afterMoving := func(x, y int32) error {
			return nil
		}
		isFree := func(pos common.Vec2) bool {
			x, y := pos.Get()
			tcell := w.Get(x, y)
			if tcell == nil {
				return false
			}
			if !isFree(pos) && tcell.CellType != cell.VACUUM_CELL {
				w.SetVec(pos, cell.NewCell(cell.EMPTY_CELL))
			}
			return false
		}
		return simulateCustomMovements(pos, maxSpeed, c, isFree, afterMoving, groups)
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
			if (tcell.CellType == cell.WOOD_CELL || tcell.CellType == cell.LEAF_CELL) && (*c).RemainingLife != 0 {
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
			w.Set(uint16(x), uint16(y), cell.NewCell(cell.EMPTY_CELL))
			return nil
		}
		isFree := func(pos common.Vec2) bool {
			x, y := pos.Get()
			tcell := w.Get(x, y)
			if tcell == nil {
				return false
			}
			if tcell.CellType == cell.WATER_CELL {
				smoke, _ := NewCellByChance(cell.SMOKE_CELL, 5)
				w.SetVec(pos, smoke)
				return false
			}
			if tcell.CellType == cell.WOOD_CELL {
				w.Set(uint16(x), uint16(y), cell.NewCell(cell.FIRE_CELL))
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
		case cell.LAVA_CELL:
			simulateLavaMovements(pos, 1, &center, []common.Vec2{
				common.NewVec2(0, 1),
				common.NewVec2(1, 1),
				common.NewVec2(-1, 1),
				common.NewVec2(-1, 0),
				common.NewVec2(1, 0),
			})
		case cell.LEAF_CELL:
			simulateLeafMovements(pos, 1, &center, []common.Vec2{
				common.NewVec2(0, 1),
				common.NewVec2(1, 1),
				common.NewVec2(-1, 1),
			})
		case cell.WATER_CELL:
			simulateWaterMovements(pos, 2, &center, []common.Vec2{
				common.NewVec2(0, 1),
				common.NewVec2(1, 1),
				common.NewVec2(-1, 1),
				common.NewVec2(-1, 0),
				common.NewVec2(1, 0),
			})
		case cell.VACUUM_CELL:
			if center.RemainingLife <= 0 {
				w.Set(_x, _y, cell.NewCell(cell.EMPTY_CELL))
				return nil
			}
			center.DecreaseLife()
			moved := simulateVacummMovements(pos, 1, &center, []common.Vec2{
				common.NewVec2(0, 1),
				common.NewVec2(0, -1),
				common.NewVec2(1, 0),
				common.NewVec2(-1, 0),
			})
			if !moved {
				center.Touched()
				w.activeChunks.SortedInsert(idChunk)
			}
		case cell.SMOKE_CELL:
			if center.RemainingLife <= 0 {
				w.Set(_x, _y, cell.NewCell(cell.EMPTY_CELL))
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
				smoke, _ := NewCellByChance(cell.SMOKE_CELL, 5)
				w.Set(_x, _y, smoke)
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
func (w *ServerWorld) GetActiveChunksAndNeiboroud() (res []uint16) {
	l := common.NewOrderList[uint16]()
	chunks := w.activeChunks.Get()
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
			l.SortedInsert(uint16(n))
		}
	}
	return l.GetReversSort()
}

func (w *ServerWorld) GetChunksToSend() []uint16 {
	return w.activeChunks.Get()
}

func (w *world) SetVec(pos common.Vec2, cell cell.Cell) {
	x, y := pos.Get()
	w.Set(uint16(x), uint16(y), cell)
}

func (w *world) Set(x, y uint16, cell cell.Cell) {
	if x >= w.W {
		return
	}
	if y >= w.H {
		return
	}
	w.activeChunks.SortedInsert(w.GetChunkId(x, y))
	indexCell := x + (y * w.W)
	w.cells[indexCell] = cell
}

func (w *world) Get(_x, _y int32) *cell.Cell {
	if _x < 0 {
		return nil
	}
	if _y < 0 {
		return nil
	}
	x := uint16(_x)
	y := uint16(_y)
	if x >= w.W {
		return nil
	}
	if y >= w.H {
		return nil
	}
	return &w.cells[x+(y*w.W)]
}
