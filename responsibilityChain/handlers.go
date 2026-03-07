package responsibilityChain

import (
	"fmt"
	"io"
	"sand-mmo/cell"
	"sand-mmo/common"

	ws "github.com/gorilla/websocket"
)

func GetHandlers() []Handler {
	return []Handler{
		{
			p: GetChunkCommand(0),
			handler: func(p common.Package, e *ResponsibilityChain) error {
				bytes := e.world.GetChunkBytesToSend(uint16(p.Arg))
				//			fmt.Printf("Sending.. %v\n", len(bytes))
				err := e.tcpConn.WriteMessage(0, bytes)
				if err != nil {
					return err
				}
				//		fmt.Println("ReturnChunk")
				return err
			},
		},
		{
			p: GetDrawCommand(0, 0, cell.EMPTY_CELL, 0),
			handler: func(p common.Package, e *ResponsibilityChain) error {
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
								cell, err := cell.NewCell(p.CellType)
								if err != nil {
									return err
								}
								e.world.Set(uint16(x), uint16(y), cell)

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
							cell, err := cell.NewCell(p.CellType)
							if err != nil {
								return err
							}
							e.world.Set(uint16(x), uint16(y), cell)
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
			},
		},
		{
			p: GetInitCommand(0),
			handler: func(p common.Package, e *ResponsibilityChain) error {
				fmt.Println("Init bidirectional connection ", e.tcpConn.RemoteAddr())
				for i := range e.world.GetNumberChucks() {
					e.tcpConn.WriteMessage(ws.TextMessage, e.world.GetChunkBytesToSend(i))
				}
				return nil
			},
		},
		{
			p: GetENDCommand(),
			handler: func(p common.Package, e *ResponsibilityChain) error {
				fmt.Println("End ", e.tcpConn.RemoteAddr())
				return io.ErrUnexpectedEOF
			},
		},
	}
}
