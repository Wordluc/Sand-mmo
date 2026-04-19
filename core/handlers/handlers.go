package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sand-mmo/common"

	ws "github.com/coder/websocket"
)

func GetHandlers() []handler {
	return []handler{
		{
			p: common.GET,
			handler: func(p common.Package, e *CoreHandlers) error {
				bytes := e.world.GetChunkBytesToSend(int(p.Arg1))
				err := e.client.Conn.Write(context.Background(), ws.MessageBinary, bytes)
				if err != nil {
					return err
				}
				return err
			},
		},
		{
			p: common.DRAW_IN,
			handler: func(p common.Package, e *CoreHandlers) error {
				x, y := common.GetServerXYChunk(e.client.AtChunkId)
				if !e.world.IsCoordinateValid(x, y) {
					return errors.New("Coordinate not valid")
				}
				p.BrushPackage.X += uint16(x) * common.CHUNK_SIZE
				p.BrushPackage.Y += uint16(y) * common.CHUNK_SIZE
				if e.IsLastCommand(common.ADD_GENERATOR) {
					e.world.AddGenerator(p.BrushPackage)
					return nil
				}
				err, _ := e.world.ApplyBrush(p.BrushPackage)
				return err
			},
		},
		{
			p: common.INIT,
			handler: func(p common.Package, e *CoreHandlers) error {
				if !e.world.IsChunkIdValid(int(p.Arg1)) {
					return errors.New("INIT: ChunkId not valid")
				}
				fmt.Println("Init bidirectional connection: "+e.client.Addr, " at ", p.Arg1)
				e.client.AtChunkId = int(p.Arg1)
				e.netCode.SendInitialChunks(e.client)
				return nil
			},
		},
		{
			p: common.INITGOD,
			handler: func(p common.Package, e *CoreHandlers) error {
				fmt.Println("Init god bidirectional connection: " + e.client.Addr)
				e.client.IsGod = true
				e.netCode.SendInitialChunksForGod(e.client)
				return nil
			},
		},
		{
			p: common.MOVE_AT,
			handler: func(p common.Package, e *CoreHandlers) error {
				if !e.world.IsChunkIdValid(int(p.Arg1)) {
					return errors.New("MOVE_AT: ChunkId not valid")
				}
				oldX, oldY := common.GetServerXYChunk(e.client.AtChunkId)
				e.client.AtChunkId = int(p.Arg1)
				newX, newY := common.GetServerXYChunk(e.client.AtChunkId)
				var chunksToSend []int
				if newX > oldX {
					chunksToSend = make([]int, 0, common.H_CELLS_CLIENT)
					for y := newY; y < newY+common.H_CHUNKS_CLIENT; y++ {
						chunksToSend = append(chunksToSend, newX+common.W_CHUNKS_CLIENT+y*common.W_CHUNKS_TOTAL-1) //retrieve last column of the client's view
					}
				} else if newX < oldX {
					chunksToSend = make([]int, 0, common.H_CELLS_CLIENT)
					for y := newY; y < newY+common.H_CHUNKS_CLIENT; y++ {
						chunksToSend = append(chunksToSend, newX+y*common.W_CHUNKS_TOTAL) //retrieve first column of the client's view
					}
				}
				if newY > oldY {
					chunksToSend = make([]int, 0, common.W_CELLS_CLIENT)
					for x := newX; x < newX+common.W_CHUNKS_CLIENT; x++ {
						chunksToSend = append(chunksToSend, x+(newY+common.H_CHUNKS_CLIENT-1)*common.W_CHUNKS_TOTAL) //retrieve last row of the client's view
					}
				} else if newY < oldY {
					chunksToSend = make([]int, 0, common.W_CELLS_CLIENT)
					for x := newX; x < newX+common.W_CHUNKS_CLIENT; x++ {
						chunksToSend = append(chunksToSend, x+newY*common.W_CHUNKS_TOTAL) //retrieve first row of the client's view
					}
				}
				//left border
				for y := newY; y < newY+common.H_CHUNKS_CLIENT; y++ {
					if (newX - 1) < 0 {
						break
					}
					chunksToSend = append(chunksToSend, (newX-1)+y*common.W_CHUNKS_TOTAL)
				}
				//right border
				for y := newY; y < newY+common.H_CHUNKS_CLIENT; y++ {
					if (newX + common.W_CHUNKS_CLIENT) > common.W_CHUNKS_TOTAL {
						break
					}
					chunksToSend = append(chunksToSend, (newX+common.W_CHUNKS_CLIENT)+y*common.W_CHUNKS_TOTAL)
				}
				//up border
				for x := newX; x < newX+common.W_CHUNKS_CLIENT; x++ {
					if (newY - 1) < 0 {
						break
					}
					chunksToSend = append(chunksToSend, x+(newY-1)*common.W_CHUNKS_TOTAL)
				}
				//down border
				for x := newX; x < newX+common.W_CHUNKS_CLIENT; x++ {
					if (newY + common.H_CHUNKS_CLIENT) > common.H_CHUNKS_TOTAL {
						break
					}
					chunksToSend = append(chunksToSend, x+(newY+common.H_CHUNKS_CLIENT)*common.W_CHUNKS_TOTAL)
				}

				var chunks map[int][]byte = make(map[int][]byte, len(chunksToSend))
				maxIdChunk := e.world.GetNumberChucks()
				for _, iC := range chunksToSend {
					if iC >= maxIdChunk {
						println("OutOfBound world")
						continue
					}
					chunks[iC] = e.world.GetChunkBytesToSend(iC)
				}
				e.netCode.SendChunksTo(chunks, e.client)
				return nil
			},
		},
		{
			p: common.ADD_GENERATOR,
			handler: func(p common.Package, e *CoreHandlers) error {
				//Used to set "LastCommand"
				return nil
			},
		},
		{
			p: common.END,
			handler: func(p common.Package, e *CoreHandlers) error {
				return io.ErrUnexpectedEOF
			},
		},
	}
}
