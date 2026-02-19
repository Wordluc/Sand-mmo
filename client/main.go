package main

import (
	"net"
	commandengine "sand-mmo/commandEngine"
	"sand-mmo/common"

	"github.com/gen2brain/raylib-go/raylib"
)

const W_WINDOWS = 100
const H_WINDOWS = 100

func main() {
	rl.InitWindow(W_WINDOWS, H_WINDOWS, "")
	socket, err := net.Dial("tcp", ":8000")
	if err != nil {
		panic(err)
	}
	var mousePosition struct {
		X uint16
		Y uint16
	}

	p := common.Encode(commandengine.GetInitCommand())
	common.SendToSocket(p, socket)

	for {
		if rl.WindowShouldClose() {
			return
		}
		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			vec := rl.GetMousePosition()
			mousePosition.X = uint16(vec.X)
			mousePosition.Y = uint16(vec.Y)

			x := uint8(mousePosition.X)
			y := uint8(mousePosition.Y)
			p := common.Encode(commandengine.GetDrawCommand(0, x, y))
			common.SendToSocket(p, socket)

		}
		rl.BeginDrawing()
		rl.EndDrawing()
	}

}
