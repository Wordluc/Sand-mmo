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
			p: GetDrawCommand(0, 0, 0, sandmmo.NULL_CELL),
			handler: func(p common.Package, e *ResponsibilityChain) error {
				//	fmt.Printf("Draw %v %v\n", p.X, p.Y)
				//TODO: to change, create a factory of cell
				e.world.Set(p.X, p.Y, sandmmo.Cell{CellType: p.CellType, Life: 50})
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
