package responsibilityChain

import (
	"fmt"
	"io"
	"sand-mmo/common"

	ws "github.com/gorilla/websocket"
)

func GetHandlers() []Handler {
	return []Handler{
		{
			p: common.GET,
			handler: func(p common.Package, e *ResponsibilityChain) error {
				bytes := e.world.GetChunkBytesToSend(uint16(p.Arg))
				err := e.webSocket.WriteMessage(ws.BinaryMessage, bytes)
				if err != nil {
					return err
				}
				return err
			},
		},
		{
			p: common.DRAW_IN,
			handler: func(p common.Package, e *ResponsibilityChain) error {
				if e.LastCommand == common.ADD_GENERATOR {
					e.world.AddGenerator(p.BrushPackage)
					return nil
				}
				return e.world.ApplyBrush(p.BrushPackage)
			},
		},
		{
			p: common.INIT,
			handler: func(p common.Package, e *ResponsibilityChain) error {
				fmt.Println("Init bidirectional connection ", e.webSocket.RemoteAddr())
				for i := range e.world.GetNumberChucks() {
					e.webSocket.WriteMessage(ws.BinaryMessage, e.world.GetChunkBytesToSend(i))
				}
				return nil
			},
		},
		{
			p: common.ADD_GENERATOR,
			handler: func(p common.Package, e *ResponsibilityChain) error {
				return nil
			},
		},
		{
			p: common.END,
			handler: func(p common.Package, e *ResponsibilityChain) error {
				fmt.Println("End ", e.webSocket.RemoteAddr())
				return io.ErrUnexpectedEOF
			},
		},
	}
}
