package world

import (
	"context"
	"encoding/binary"
	"fmt"
	"maps"
	"sand-mmo/cell"
	"sand-mmo/common"
	"sync"
	"time"

	ws "github.com/coder/websocket"
	"github.com/redis/go-redis/v9"
)

type ServerWorld struct {
	world
	webSockets     map[string]*ws.Conn
	webSocketMutex *sync.Mutex
	redis          *redis.Client
}

const REDIS_KEY_BYTES_BYTES = "world:bytes"
const REDIS_KEY_BYTES_GENERATOR = "world:generator"

func NewServerWorld(w, h, chunkSize uint16, redisClient *redis.Client) (res ServerWorld) {
	res.world = newWorld(w, h, chunkSize)
	res.webSockets = map[string]*ws.Conn{}
	res.webSocketMutex = &sync.Mutex{}
	res.redis = redisClient
	err := res.LoadSnapshot()
	if err != nil {
		fmt.Println("Falling loading world")
	}
	return res
}

func (w *ServerWorld) LoadSnapshot() error {
	get := func(key string) ([]byte, error) {
		ctx, p := context.WithTimeout(context.Background(), common.SLEEP*time.Millisecond)
		defer p()
		worldBytes, err := w.redis.Get(ctx, key).Result()
		if err == nil {
			return []byte(worldBytes), nil
		}
		switch err {
		case redis.Nil:
			return []byte{}, nil
		default:
			return []byte{}, err
		}
	}
	worldBytes, err := get(REDIS_KEY_BYTES_BYTES)
	if err != nil {
		return err
	}
	w.ImportCells(worldBytes)

	generatorBytes, err := get(REDIS_KEY_BYTES_GENERATOR)
	if err != nil {
		return err
	}
	w.ImportGenerators(generatorBytes)
	return nil
}

func (w *ServerWorld) SaveSnapshot() {
	if w.redis == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	worldBytes := w.GetWorldBytes()
	w.redis.Set(ctx, REDIS_KEY_BYTES_BYTES, string(worldBytes), 0)
	generatorsBytes := w.GetGeneratorsBytes()
	w.redis.Set(ctx, REDIS_KEY_BYTES_GENERATOR, string(generatorsBytes), 0)
	println("World Saved")
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

func (w *ServerWorld) GetLenClients() int {
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

func (w *ServerWorld) ApplyBrush(p common.BrushPackage) (err error, metVacuum bool) {
	var c *cell.Cell
	drawCircle := func(radius int) error {
		for iy := range radius * 2 {
			for ix := range radius * 2 {
				dx := (radius - ix)
				dy := (radius - iy)

				x := int32(int(p.X) - dx)
				if x < 0 {
					continue
				}
				y := int32(int(p.Y) - dy)
				if y < 0 {
					continue
				}
				if (dx*dx + dy*dy) <= radius*radius/4 {
					c = w.Get(x, y)
					if c == nil {
						continue
					}
					if c.CellType == cell.VACUUM_CELL {
						metVacuum = true
					}
					w.Set(x, y, cell.NewCell(p.CellType))

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

				x := int32(int(p.X) - dx)
				y := int32(int(p.Y) - dy)
				c = w.Get(x, y)
				if c == nil {
					continue
				}
				if c.CellType == cell.VACUUM_CELL {
					metVacuum = true
				}
				w.Set(x, y, cell.NewCell(p.CellType))
			}
		}
		return nil
	}
	switch p.BrushType {
	case common.CIRCLE_SMALL:
		return drawCircle(4), metVacuum
	case common.CIRCLE_BIG:
		return drawCircle(6), metVacuum
	case common.SQUARE_SMALL:
		return drawBox(4), metVacuum
	case common.SQUARE_BIG:
		return drawBox(6), metVacuum
	}
	return nil, metVacuum
}

func (w *ServerWorld) ImportGenerators(gen []byte) {
	var u64Generator []uint64
	for i := 0; i < len(gen); i += 8 {
		u64Generator = append(u64Generator, binary.BigEndian.Uint64(gen[i:i+8]))
	}
	for i := range u64Generator {
		w.generators = append(w.generators, common.Decode(u64Generator[i]).BrushPackage)
	}
}

func (w *ServerWorld) ImportCells(cells []byte) {
	var u16World []uint16
	for i := 0; i < len(cells); i += 2 {
		u16World = append(u16World, binary.BigEndian.Uint16(cells[i:i+2]))
	}
	for i := range u16World {
		w.cells[i] = cell.DecodeCell(u16World[i])
	}

	for i := range w.GetNumberChucks() {
		w.activeChunks.SortedInsert(i)
	}

}

func (w *ServerWorld) AddGenerator(brush common.BrushPackage) {
	w.generators = append(w.generators, brush)
}

func (w *ServerWorld) Loop() error {
	err := w.ApplyGenerators()
	if err != nil {
		return err
	}
	chunksToSend := w.GetActiveChunksAndNeiboroud()
	for _, iC := range chunksToSend {
		w.SimulateChunk(uint16(iC))
	}
	return nil
}
func (w *ServerWorld) ApplyGenerators() error {
	newGenerators := make([]common.BrushPackage, 0)
	for i := range w.generators {
		err, metVacuum := w.ApplyBrush(w.generators[i])
		if err != nil {
			return err
		}
		if !metVacuum {
			newGenerators = append(newGenerators, w.generators[i])
		}
	}
	w.generators = newGenerators
	return nil
}

func (w *ServerWorld) SimulateChunk(idChunk uint16) error {

	isFree := func(pos common.Vec2) bool {
		x, y := pos.Get()
		c := w.Get(x, y)

		return c != nil && (c.IsEmpty())
	}
	simulateCustomMovements := func(pos common.Vec2, maxSpeed int32, cell **cell.Cell, callbackBeforeMoving func(vec common.Vec2) bool, callbackAfterMoving func(x, y int32) error, groups []common.Vec2) bool {
		oldx, oldy := pos.Get()
		move := func(v common.Vec2) {
			pos.Add(v)
			(*cell).Touched()
			w.SetVec(pos, *(*cell))

			*cell = w.GetVec(pos)
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
			w.Set(x, y, cell.NewCell(cell.EMPTY_CELL))
			return nil
		}
		return simulateCustomMovements(pos, maxSpeed, c, isFree, afterMoving, groups)
	}

	simulateWaterMovements := func(pos common.Vec2, maxSpeed int32, c **cell.Cell, groups []common.Vec2) bool {
		afterMoving := func(x, y int32) error {
			w.Set(x, y, cell.NewCell(cell.EMPTY_CELL))
			return nil
		}
		beforeMoving := func(posToCheck common.Vec2) bool {
			x, y := posToCheck.Get()
			tcell := w.Get(x, y)
			if tcell == nil {
				return false
			}
			if tcell.CellType == cell.LAVA_CELL || tcell.CellType == cell.FIRE_CELL {
				smoke, isSmoke := NewCellByChance(cell.SMOKE_CELL, 10)
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
			w.Set(x, y, cell.NewCell(cell.EMPTY_CELL))
			return nil
		}
		beforeMoving := func(posToCheck common.Vec2) bool {
			tcell := w.GetVec(posToCheck)
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
			tcell := w.GetVec(pos)
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
			tcell := w.GetVec(pos)
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
			w.Set(x, y, cell.NewCell(cell.EMPTY_CELL))
			return nil
		}
		isFree := func(pos common.Vec2) bool {
			tcell := w.GetVec(pos)
			if tcell == nil {
				return false
			}
			if tcell.CellType == cell.WATER_CELL {
				smoke, _ := NewCellByChance(cell.SMOKE_CELL, 10)
				w.SetVec(pos, smoke)
				return false
			}
			if tcell.CellType == cell.WOOD_CELL || tcell.CellType == cell.LEAF_CELL {
				w.SetVec(pos, cell.NewCell(cell.FIRE_CELL))
				return false
			}
			return isFree(pos)
		}
		return simulateCustomMovements(pos, maxSpeed, c, isFree, afterMoving, groups)
	}
	return w.ForEachCell(idChunk, func(_x, _y uint16, center *cell.Cell) error {
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
				w.SetVec(pos, cell.NewCell(cell.EMPTY_CELL))
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
				w.SetVec(pos, cell.NewCell(cell.EMPTY_CELL))
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
				smoke, _ := NewCellByChance(cell.SMOKE_CELL, 20)
				w.SetVec(pos, smoke)
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
	w.Set(x, y, cell)
}

func (w *world) Set(_x, _y int32, cell cell.Cell) {
	x := uint16(_x)
	y := uint16(_y)
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

func (w *world) GetVec(pos common.Vec2) *cell.Cell {
	x, y := pos.Get()
	return w.Get(x, y)
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
