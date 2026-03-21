package core

import (
	"encoding/binary"
	"sand-mmo/cell"
	"sand-mmo/common"
)

type ServerWorld struct {
	world
}

func NewServerWorld(w, h, chunkSize int) (res ServerWorld) {
	res.world = newWorld(w, h, chunkSize)
	return res
}

func (w *ServerWorld) ApplyBrush(p common.BrushPackage) (err error, metVacuum bool) {
	var c *cell.Cell
	drawCircle := func(radius int) error {
		for i_y := range radius * 2 {
			for ix := range radius * 2 {
				dx := (radius - ix)
				dy := (radius - i_y)

				x := int(p.X) - dx
				if x < 0 {
					continue
				}
				y := int(p.Y) - dy
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
		for i_y := range size * 2 {
			for ix := range size * 2 {
				dx := (size - ix)
				dy := (size - i_y)

				x := int(p.X) - dx
				y := int(p.Y) - dy
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
		w.SimulateChunk(iC)
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

func (w *ServerWorld) GetActiveChunksAndNeiboroud() (res []int) {
	l := common.NewOrderList[int]()
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
			l.SortedInsert(n)
		}
	}
	return l.GetReversSort()
}

func (w *ServerWorld) GetChunksToSend() []int {
	return w.activeChunks.Get()
}

func (w *world) SetVec(pos common.Vec2, cell cell.Cell) {
	x, y := pos.Get()
	w.Set(x, y, cell)
}

func (w *world) Set(x, y int, cell cell.Cell) {
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

func (w *world) Get(x, y int) *cell.Cell {
	if x < 0 {
		return nil
	}
	if y < 0 {
		return nil
	}
	if x >= w.W {
		return nil
	}
	if y >= w.H {
		return nil
	}
	return &w.cells[x+(y*w.W)]
}
