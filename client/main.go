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
	w := sandmmo.NewWorld(sandmmo.W_WINDOWS, sandmmo.H_WINDOWS, 25)
	socket, err := net.Dial("tcp", ":8000")
	if err != nil {
		panic(err)
	}

	UpdateWorld(&w, socket)
	p := chain.GetChunkCommand(0)
	common.SendToTcpSocket(p, socket)
	for {
		if rl.WindowShouldClose() {
			return
		}
		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			vec := rl.GetMousePosition()
			x := uint16(vec.X) / sandmmo.SIZE_CELL
			y := uint16(vec.Y) / sandmmo.SIZE_CELL
			chunkId := w.GetChuck(x, y)
			common.SendToTcpSocket(chain.GetDrawCommand(uint8(chunkId), x, y), socket)
			common.SendToTcpSocket(chain.GetChunkCommand(uint32(chunkId)), socket)
		}
		rl.BeginDrawing()
		w.Draw()
		rl.EndDrawing()
	}

}

func UpdateWorld(world *sandmmo.World, tcp net.Conn) {
	var port uint32 = 8005
	add, _ := net.ResolveUDPAddr("udp", fmt.Sprint("127.0.0.1:", port))
	udpConn, err := net.ListenUDP("udp", add)
	if err != nil {
		panic(err)
	}
	common.SendToTcpSocket(chain.GetInitCommand(port), tcp)
	println("Udp connection ", udpConn.LocalAddr().String())
	go func() {
		for {
			//4->32bit
			var bytes []byte = make([]byte, 4*world.ChunkSize*world.ChunkSize+2)
			n, _, err := udpConn.ReadFromUDP(bytes)
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
