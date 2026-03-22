package handlers

import (
	"context"
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
				if e.IsLastCommand(common.ADD_GENERATOR) {
					e.world.AddGenerator(p.BrushPackage)
					return nil
				}
				x, y := e.world.GetGlobalXYChunk(e.client.AtChunkId)
				p.BrushPackage.X += uint16(x) * common.CHUNK_SIZE
				p.BrushPackage.Y += uint16(y) * common.CHUNK_SIZE
				err, _ := e.world.ApplyBrush(p.BrushPackage)
				return err
			},
		},
		{
			p: common.INIT,
			handler: func(p common.Package, e *CoreHandlers) error {
				fmt.Println("Init bidirectional connection: " + e.client.Addr)
				e.netCode.SendAllChunksTo(e.client)
				return nil
			},
		},
		{
			p: common.MOVE_AT,
			handler: func(p common.Package, e *CoreHandlers) error {
				e.client.AtChunkId = int(p.Arg1)
				e.netCode.SendAllChunksTo(e.client)
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
