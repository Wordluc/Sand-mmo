package world

import (
	"encoding/binary"
	"sand-mmo/cell"
	"sand-mmo/common"
)

type world struct {
	W            uint16
	H            uint16
	ChunkSize    uint16
	cells        []cell.Cell
	activeChunks common.OrderList[uint16]
	generators   []common.BrushPackage
}

func newWorld(w, h, chunkSize uint16) world {
	world := world{}
	world.cells = make([]cell.Cell, w*h)
	world.H = h
	world.W = w
	world.ChunkSize = chunkSize
	world.activeChunks = common.NewOrderList[uint16]()
	return world
}

func (w *world) SetCellsByte(bytes []byte, idChunk uint16) {
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

// For test
func (w *world) importCell(cells []uint16) {
	w.cells = []cell.Cell{}
	for i := range cells {
		w.cells = append(w.cells, cell.DecodeCell(cells[i]))
	}
}

func (w *world) GetChunkId(x, y uint16) uint16 {
	chunkPerRow := w.W / w.ChunkSize
	id := (y/w.ChunkSize)*chunkPerRow + x/w.ChunkSize
	return id
}
func (w *world) GetNumberChucks() uint16 {
	return w.W / w.ChunkSize * w.H / w.ChunkSize
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

func (w *world) GetChunkBytes(idChunk uint16) []uint16 {
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
func (w *world) GetChunkBytesToSend(idChunk uint16) []byte {
	chunk := w.GetChunkBytes(idChunk)
	var bytes []byte
	bytes = binary.BigEndian.AppendUint16(bytes, idChunk)
	for i := range chunk {
		bytes = binary.BigEndian.AppendUint16(bytes, chunk[i])
	}
	return bytes
}
