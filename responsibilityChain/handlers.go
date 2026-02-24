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
			p: GetDrawCommand(0, 0, 0),
			handler: func(p common.Package, e *ResponsibilityChain) error {
				//	fmt.Printf("Draw %v %v\n", p.X, p.Y)
				e.world.Set(p.X, p.Y, sandmmo.Cell{Cell: 1})
				return nil
			},
		},
		{
			p: GetInitCommand(0),
			handler: func(p common.Package, e *ResponsibilityChain) error {
				ip := e.tcpConn.RemoteAddr().(*net.TCPAddr).IP
				fmt.Println("Init ", e.tcpConn.RemoteAddr())
				addrTo := &net.UDPAddr{IP: ip, Port: int(p.Arg)}
				e.callbackInitUdp(addrTo)
				for i := range e.world.GetNumberChucks() {
					e.udpConn.WriteTo(e.world.GetChunkBytesToSend(i), addrTo)
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
