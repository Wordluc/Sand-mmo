package responsibilityChain

import (
	"errors"
	"fmt"
	sandmmo "sand-mmo"
	"sand-mmo/common"
)

func GetHandlers() []Handler {
	return []Handler{
		{
			p: GetChunkCommand(0),
			handler: func(p common.Package, e *ResponsibilityChain) error {
				bytes := e.world.GetChunkBytesToSend(uint16(p.Arg))
				//			fmt.Printf("Sending.. %v\n", len(bytes))
				n, err := e.udpConn.Write(bytes)
				if err != nil {
					return err
				}
				if n <= 0 {
					return errors.New("Error sending cells")
				}
				//		fmt.Println("ReturnChunk")
				return err
			},
		},
		{
			p: GetDrawCommand(0, 0, sandmmo.NULL_CELL, 0),
			handler: func(p common.Package, e *ResponsibilityChain) error {
				//	fmt.Printf("Draw %v %v\n", p.X, p.Y)
				//TODO: to change, create a factory of cell
				drawCircle := func(radius int) {
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
								e.world.Set(uint16(x), uint16(y), sandmmo.NewCell(p.CellType, 10))
							}
						}
					}
				}
				drawBox := func(size int) {
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
							cell := sandmmo.NewCell(p.CellType, 10)
							e.world.Set(uint16(x), uint16(y), cell)
						}
					}
				}
				size := common.GetSizeFromBrushType(p.BrushType)
				switch p.BrushType {
				case common.CIRCLE_SMALL, common.CIRCLE_BIG:
					drawCircle(size)
				case common.SQUARE_SMALL, common.SQUARE_BIG:
					drawBox(size)

				}
				return nil
			},
		},
		{
			p: GetInitCommand(0),
			handler: func(p common.Package, e *ResponsibilityChain) error {
				t := make([]byte, 4)
				_, addr, _ := e.udpConn.ReadFromUDP(t)
				e.callbackAddUdp(e.tcpConn.RemoteAddr(), addr)
				fmt.Println("Init udp connection with", addr)

				for i := range e.world.GetNumberChucks() {
					e.udpConn.WriteToUDP(e.world.GetChunkBytesToSend(i), addr)
				}
				return nil
			},
		},
		{
			p: GetENDCommand(),
			handler: func(p common.Package, e *ResponsibilityChain) error {
				fmt.Println("End ", e.tcpConn.RemoteAddr())
				e.callbackRemoveUdp(e.tcpConn.RemoteAddr())
				return nil
			},
		},
	}
}
