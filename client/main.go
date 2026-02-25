package main

import (
	"encoding/binary"
	"fmt"
	"net"
	sandmmo "sand-mmo"
	"sand-mmo/common"
	chain "sand-mmo/responsibilityChain"

	ru "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const W_BUTTONS_SIDE = 100
const W_GAME = sandmmo.W_WINDOWS * sandmmo.SIZE_CELL

func main() {
	rl.InitWindow(W_GAME+W_BUTTONS_SIDE, sandmmo.H_WINDOWS*sandmmo.SIZE_CELL, "")
	w := sandmmo.NewWorld(sandmmo.W_WINDOWS, sandmmo.H_WINDOWS, sandmmo.CHUNK_SIZE)
	socket, err := net.Dial("tcp", ":8000")
	if err != nil {
		panic(err)
	}
	udp := createDialUdp(socket)
	UpdateWorld(&w, udp)
	defer udp.Close()
	defer common.SendToTcpSocket(chain.GetENDCommand(), socket)
	//Insert fps target
	rl.SetTargetFPS(30)
	var cellType sandmmo.CellType = sandmmo.SAND_CELL
	for {
		if rl.WindowShouldClose() {
			return
		}
		//avoid to send same package twise
		if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
			vec := rl.GetMousePosition()
			x := uint16(vec.X) / sandmmo.SIZE_CELL
			y := uint16(vec.Y) / sandmmo.SIZE_CELL
			chunkId := w.GetChunkId(x, y)
			common.SendToTcpSocket(chain.GetDrawCommand(uint8(chunkId), x, y, cellType), socket)
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.SkyBlue)
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 0, Width: 50, Height: 20}, "Water") {
			cellType = sandmmo.WATER_CELL
		}
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 25, Width: 50, Height: 20}, "Sand") {
			cellType = sandmmo.SAND_CELL
		}
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 25 * 2, Width: 50, Height: 20}, "Smoke") {
			cellType = sandmmo.SMOKE_CELL
		}
		w.Draw()
		rl.EndDrawing()
	}

}

func createDialUdp(tcp net.Conn) *net.UDPConn {
	addTo, _ := net.ResolveUDPAddr("udp", fmt.Sprint(tcp.RemoteAddr().(*net.TCPAddr).IP, ":", 8000))
	udpConn, err := net.DialUDP("udp", nil, addTo)
	if err != nil {
		panic(err)
	}
	common.SendToTcpSocket(chain.GetInitCommand(8000), tcp)
	udpConn.Write([]byte("ping"))
	println("Udp connection ", udpConn.LocalAddr().String())

	return udpConn
}

func UpdateWorld(world *sandmmo.World, udp *net.UDPConn) {
	go func() {
		for {
			//4->32bit
			var bytes []byte = make([]byte, 4*world.ChunkSize*world.ChunkSize+2)
			n, _, err := udp.ReadFrom(bytes)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if n <= 0 {
				continue
			}
			idChunk := binary.BigEndian.Uint16(bytes[0:2])
			world.SetCellsByte(bytes[2:], idChunk)
		}
	}()
}
