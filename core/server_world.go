package core

import (
	"encoding/binary"
	"sand-mmo/common"
)

type ServerWorld struct {
	world
}

func newServerWorld_test(w, h, chunkSize int) (res ServerWorld) {
	res.world = newWorld(w, h, chunkSize)
	return res
}

func NewServerWorld() (res ServerWorld) {
	res.world = newWorld(common.W_CELLS_TOTAL, common.H_CELLS_TOTAL, common.CHUNK_SIZE)
	return res
}

func (w *ServerWorld) ApplyBrush(p common.BrushPackage) (err error, metVoid bool) {
	var c *Cell
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
					if c.CellType == VOID_CELL {
						metVoid = true
					}
					w.Set(x, y, NewCell(p.CellType))

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
				if c.CellType == VOID_CELL {
					metVoid = true
				}
				w.Set(x, y, NewCell(p.CellType))
			}
		}
		return nil
	}
	switch p.BrushType {
	case common.CIRCLE_SMALL:
		return drawCircle(4), metVoid
	case common.CIRCLE_BIG:
		return drawCircle(6), metVoid
	case common.SQUARE_SMALL:
		return drawBox(4), metVoid
	case common.SQUARE_BIG:
		return drawBox(6), metVoid
	}
	return nil, metVoid
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
	if len(u16World) != len(w.cells) {
		println("Error loading world")
		return
	}
	for i := range u16World {
		w.cells[i] = DecodeCell(u16World[i])
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
		err, metVoid := w.ApplyBrush(w.generators[i])
		if err != nil {
			return err
		}
		if !metVoid {
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

func (w *world) IsCoordinateValid(x, y int) bool {
	if x >= w.W || x < 0 {
		return false
	}
	if y >= w.H || y < 0 {
		return false
	}
	return true
}

func (w *world) IsChunkIdValid(chunkId int) bool {
	if chunkId < 0 || chunkId >= w.GetNumberChucks() {
		return false
	}
	return true
}

func (w *world) SetVec(pos common.Vec2, cell Cell) {
	x, y := pos.Get()
	w.Set(x, y, cell)
}

func (w *world) Set(x, y int, cell Cell) {
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

func (w *world) GetVec(pos common.Vec2) *Cell {
	x, y := pos.Get()
	return w.Get(x, y)
}

func (w *world) Get(x, y int) *Cell {
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

func (w *ServerWorld) GetGeneratorsBytes() []byte {
	var decoded []byte = make([]byte, 8*len(w.generators))
	var mockPackage common.Package
	for i, c := range w.generators {
		mockPackage.BrushPackage = c
		binary.BigEndian.PutUint64(decoded[i*8:], common.Encode(mockPackage))
	}
	return decoded
}

func (w *ServerWorld) GetWorldBytes() []byte {
	var decoded []byte
	for _, c := range w.cells {
		decoded = binary.BigEndian.AppendUint16(decoded, EncodeCell(c))
	}
	return decoded
}

func (w *world) GetChunkBytesToSend(idChunk int) []byte {
	chunk := w.GetChunkBytes(idChunk)
	var bytes []byte = make([]byte, 2+len(chunk)*2)
	binary.BigEndian.PutUint16(bytes[0:], uint16(idChunk))
	for i := range chunk {
		binary.BigEndian.PutUint16(bytes[i*2+2:], chunk[i])
	}
	return bytes
}

func (w *world) GetChunkBytes(idChunk int) []uint16 {
	var decoded []uint16 = make([]uint16, w.ChunkSize*w.ChunkSize)

	chunkPerRow := w.W / w.ChunkSize

	chunkY := idChunk / chunkPerRow
	chunkX := idChunk % chunkPerRow
	iCell := chunkY*(w.W*w.ChunkSize) + chunkX*w.ChunkSize
	var i uint16
	for range uint16(w.ChunkSize) {
		for _, c := range w.cells[iCell : iCell+w.ChunkSize] {
			decoded[i] = EncodeCell(c)
			i++
		}
		iCell += (w.W)
	}

	return decoded
}

func (w *world) GetActiveChunks() []int {
	return w.activeChunks.Get()
}
