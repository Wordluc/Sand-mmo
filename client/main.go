package main

import (
	"net"
	sandmmo "sand-mmo"
	"sand-mmo/common"
	chain "sand-mmo/responsibilityChain"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const W_WINDOWS = 100
const H_WINDOWS = 100

func main() {
	rl.InitWindow(W_WINDOWS, H_WINDOWS, "")
	w := sandmmo.NewWorld(W_WINDOWS, H_WINDOWS)
	socket, err := net.Dial("tcp", ":8000")
	if err != nil {
		panic(err)
	}
	var mousePosition struct {
		X uint16
		Y uint16
	}

	p := common.Encode(chain.GetInitCommand())
	if err != nil {
		panic(err)
	}

	common.SendToTcpSocket(p, socket)
	go UpdateWorld(&w)
	for {
		if rl.WindowShouldClose() {
			return
		}
		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			vec := rl.GetMousePosition()
			mousePosition.X = uint16(vec.X)
			mousePosition.Y = uint16(vec.Y)

			p := common.Encode(chain.GetDrawCommand(0, mousePosition.X, mousePosition.Y))
			common.SendToTcpSocket(p, socket)

		}
		rl.BeginDrawing()
		rl.EndDrawing()
	}

}

func UpdateWorld(world *sandmmo.World) {
	udpConn, err := net.ListenPacket("udp", "127.0.0.1:8001")
	if err != nil {
		panic(err)
	}
	defer udpConn.Close()
	for {
		//4->32bit
		var bytes [4*sandmmo.SizeChunk ^ 2]byte
		udpConn.ReadFrom(bytes[:])
	}
}
