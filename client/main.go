package main

import (
	"encoding/binary"
	"fmt"
	"net"
	sandmmo "sand-mmo"
	"sand-mmo/common"
	chain "sand-mmo/responsibilityChain"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	rl.InitWindow(sandmmo.W_WINDOWS*sandmmo.SIZE_CELL, sandmmo.H_WINDOWS*sandmmo.SIZE_CELL, "")
	w := sandmmo.NewWorld(sandmmo.W_WINDOWS, sandmmo.H_WINDOWS, sandmmo.CHUNK_SIZE)
	socket, err := net.Dial("tcp", ":8000")
	if err != nil {
		panic(err)
	}
	udp := createUdpConnection(socket)
	UpdateWorld(&w, udp)
	defer common.SendToTcpSocket(chain.GetENDCommand(), socket)
	//Insert fps target
	rl.SetTargetFPS(30)
	for {
		if rl.WindowShouldClose() {
			return
		}
		//avoid to send same package twise
		if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
			vec := rl.GetMousePosition()
			x := uint16(vec.X) / sandmmo.SIZE_CELL
			y := uint16(vec.Y) / sandmmo.SIZE_CELL
			chunkId := w.GetChuck(x, y)
			common.SendToTcpSocket(chain.GetDrawCommand(uint8(chunkId), x, y), socket)
		}
		rl.BeginDrawing()
		w.Draw()
		rl.EndDrawing()
	}

}
func GetFreePort() (port uint32, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return uint32(l.Addr().(*net.TCPAddr).Port), nil
		}
	}
	return
}

func createUdpConnection(tcp net.Conn) *net.UDPConn {

	port, _ := GetFreePort()
	add, _ := net.ResolveUDPAddr("udp", fmt.Sprint("127.0.0.1:", port))
	udpConn, err := net.ListenUDP("udp", add)
	if err != nil {
		panic(err)
	}
	common.SendToTcpSocket(chain.GetInitCommand(port), tcp)
	println("Udp connection ", udpConn.LocalAddr().String())

	return udpConn
}

func UpdateWorld(world *sandmmo.World, udp *net.UDPConn) {
	go func() {
		for {
			//4->32bit
			var bytes []byte = make([]byte, 4*world.ChunkSize*world.ChunkSize+2)
			n, _, err := udp.ReadFromUDP(bytes)
			if n <= 0 {
				continue
			}
			if err != nil {
				continue
			}

			port := binary.BigEndian.Uint16(bytes[0:3])
			world.SetCellsByte(bytes[2:], port)
		}
	}()
}
