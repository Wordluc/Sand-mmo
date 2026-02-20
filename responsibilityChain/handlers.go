package responsibilityChain

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sand-mmo/common"
)

func GetHandlers() []Handler {
	return []Handler{
		{
			p: GetChunkCommand(0),
			handler: func(p common.Package, e *ResponsibilityChain) error {
				chunk := e.world.GetChunk(uint16(p.Arg))
				var bytes []byte
				bytes = binary.BigEndian.AppendUint16(bytes, uint16(p.Arg))
				for i := range chunk {
					bytes = binary.BigEndian.AppendUint32(bytes, chunk[i])
				}
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
			handler: func(p common.Package, _ *ResponsibilityChain) error {
				fmt.Println("Draw")
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
				return nil
			},
		},
	}
}
