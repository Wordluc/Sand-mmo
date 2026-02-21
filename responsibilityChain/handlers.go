package responsibilityChain

import (
	"errors"
	"fmt"
	"net"
	sandmmo "sand-mmo"
	"sand-mmo/common"
)

func GetHandlers() []Handler {
	return []Handler{
		{
			p: GetChunkCommand(0),
			handler: func(p common.Package, e *ResponsibilityChain) error {
				bytes := e.world.GetChunkBytes(uint16(p.Arg))
				fmt.Printf("Sending.. %v\n", len(bytes))
				n, err := e.udpConn.Write(bytes)
				if err != nil {
					return err
				}
				if n <= 0 {
					return errors.New("Error sending cells")
				}
				fmt.Println("ReturnChunk")
				return err
			},
		},
		{
			p: GetDrawCommand(0, 0, 0),
			handler: func(p common.Package, e *ResponsibilityChain) error {
				fmt.Printf("Draw %v %v\n", p.X, p.Y)
				e.world.Set(p.X, p.Y, sandmmo.Cell{Cell: 1})
				return nil
			},
		},
		{
			p: GetInitCommand(0),
			handler: func(p common.Package, e *ResponsibilityChain) error {
				fmt.Println("Init ", p.Arg)
				add, _ := net.ResolveUDPAddr("udp", fmt.Sprint("127.0.0.1:", p.Arg))
				udpConn, err := net.DialUDP("udp", nil, add)
				if err != nil {
					return err
				}
				e.udpConn = udpConn
				e.callbackInitUdp(udpConn)
				for i := range e.world.GetNumberChucks() {
					udpConn.Write(e.world.GetChunkBytes(i))
				}
				return nil
			},
		},
		{
			p: GetENDCommand(),
			handler: func(p common.Package, e *ResponsibilityChain) error {
				fmt.Println("End ", e.udpConn.RemoteAddr())
				return nil
			},
		},
	}
}
