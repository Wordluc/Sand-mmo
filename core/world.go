package core

import (
	"math"
	"sand-mmo/common"
)

type world struct {
	W            int
	H            int
	ChunkSize    int
	cells        []Cell
	activeChunks common.OrderList[int]
	generators   []common.BrushPackage
}

func newWorld(w, h, chunkSize int) world {
	world := world{}
	world.cells = make([]Cell, w*h)
	world.H = h
	world.W = w
	world.ChunkSize = chunkSize
	world.activeChunks = common.NewOrderList[int]()
	return world
}

func (w *world) ForEachCell(idChunk int, f func(x, y int, center *Cell) error) (err error) {

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

func (w *world) GetChunkId(x, y int) int {
	id := (y/w.ChunkSize)*common.W_CHUNKS_TOTAL + x/w.ChunkSize
	return id
}

func (w *world) GetNumberChucks() int {
	return w.W / w.ChunkSize * w.H / w.ChunkSize
}
