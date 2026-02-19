package responsibilityChain

import (
	"fmt"
	"net"
	"sand-mmo/common"
	"strings"
)

func GetHandlers() []Handler {
	return []Handler{
		{
			p: GetChunkCommand(0),
			handler: func(p common.Package, _ *ResponsibilityChain) error {
				fmt.Println("ReturnChunk")
				return nil
			},
		},
		{
			p: GetDrawCommand(0, 0, 0),
			handler: func(p common.Package, _ *ResponsibilityChain) error {
				fmt.Println("Draw")
				return nil
			},
		},
		{
			p: GetInitCommand(),
			handler: func(p common.Package, e *ResponsibilityChain) error {
				fmt.Println("Init")
				fullAdd := e.tcpConn.RemoteAddr()
				add := strings.Split(fullAdd.String(), ":")
				udpConn, err := net.Dial("udp", add[0]+":8001")
				if err != nil {
					return err
				}
				common.SendToTcpSocket(1234, udpConn)
				e.udpConn = udpConn
				return nil
			},
		},
	}
}
