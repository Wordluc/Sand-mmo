package core

import (
	"encoding/binary"
	"math"
	"sand-mmo/cell"
	"sand-mmo/common"
)

type world struct {
	W            int
	H            int
	ChunkSize    int
	cells        []cell.Cell
	activeChunks common.OrderList[int]
	generators   []common.BrushPackage
}

func newWorld(w, h, chunkSize int) world {
	world := world{}
	world.cells = make([]cell.Cell, w*h)
	world.H = h
	world.W = w
	world.ChunkSize = chunkSize
	world.activeChunks = common.NewOrderList[int]()
	return world
}

func (w *world) CleanAllMap() {
	w.cells = make([]cell.Cell, w.W*w.H)
}
func (w *world) SetCellsByte(bytes []byte, idChunk int) {
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

func (w *world) GetWorldBytes() []byte {
	var decoded []byte
	for _, c := range w.cells {
		decoded = binary.BigEndian.AppendUint16(decoded, cell.EncodeCell(c))
	}
	return decoded
}

func (w *world) ForEachCell(idChunk int, f func(x, y int, center *cell.Cell) error) (err error) {

	chunkPerRow := w.W / w.ChunkSize
	chunkY := idChunk / chunkPerRow
	chunkX := idChunk % chunkPerRow
	x := chunkX*w.ChunkSize + w.ChunkSize - 1
	y := chunkY*w.ChunkSize + w.ChunkSize - 1
	for {
		if err = f(x, y, w.Get(x, y)); err != nil {
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

func (w *world) GetGeneratorsBytes() []byte {
	var decoded []byte = make([]byte, 8*len(w.generators))
	var mockPackage common.Package
	for i, c := range w.generators {
		mockPackage.BrushPackage = c
		binary.BigEndian.PutUint64(decoded[i*8:], common.Encode(mockPackage))
	}
	return decoded
}

func (w *world) GetChunkId(x, y int) int {
	id := (y/w.ChunkSize)*common.W_CHUNKS_TOTAL + x/w.ChunkSize
	return id
}

func (w *world) GetGlobalXYChunk(idChunk int) (x, y int) {
	y = idChunk / common.W_CHUNKS_TOTAL
	x = idChunk % common.W_CHUNKS_TOTAL
	return x, y
}

func (w *world) GetNumberChucks() int {
	return w.W / w.ChunkSize * w.H / w.ChunkSize
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
			decoded[i] = cell.EncodeCell(c)
			i++
		}
		iCell += (w.W)
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
