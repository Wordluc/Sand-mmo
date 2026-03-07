package main

import (
	"encoding/binary"
	"fmt"
	"net/url"
	"sand-mmo/cell"
	"sand-mmo/common"
	chain "sand-mmo/responsibilityChain"
	"sand-mmo/world"

	ru "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	ws "github.com/gorilla/websocket"
)

const W_BUTTONS_SIDE = 100
const W_GAME = common.W_WINDOWS * common.SIZE_CELL
const H_GAME = common.H_WINDOWS * common.SIZE_CELL

func main() {
	rl.InitWindow(W_GAME+W_BUTTONS_SIDE, H_GAME+common.SIZE_CELL, "")
	w := world.NewClientWorld(common.W_WINDOWS, common.H_WINDOWS, common.CHUNK_SIZE)

	conn, err := createWebSocket()
	if err != nil {
		panic(err)
	}

	go UpdateWorld(&w, conn)

	defer conn.Close()
	defer common.SendToWebSocketPackages(conn, chain.GetENDCommand())

	//Insert fps target
	rl.SetTargetFPS(30)
	var cellType cell.CellType = cell.SAND_CELL
	var brushType common.BrushType = common.CIRCLE_SMALL
	for {
		if rl.WindowShouldClose() {
			return
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.SkyBlue)
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 0, Width: 50, Height: 20}, "Water") {
			cellType = cell.WATER_CELL
		}
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 25, Width: 50, Height: 20}, "Sand") {
			cellType = cell.SAND_CELL
		}
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 50, Width: 50, Height: 20}, "Smoke") {
			cellType = cell.SMOKE_CELL
		}
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 75, Width: 50, Height: 20}, "Small Circle") {
			brushType = common.CIRCLE_SMALL
		}
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 100, Width: 50, Height: 20}, "Big Circle") {
			brushType = common.CIRCLE_BIG
		}
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 125, Width: 50, Height: 20}, "Small Square") {
			brushType = common.SQUARE_SMALL
		}
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 150, Width: 50, Height: 20}, "Big Square") {
			brushType = common.SQUARE_BIG
		}
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 175, Width: 50, Height: 20}, "Delete") {
			cellType = cell.EMPTY_CELL
		}
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 200, Width: 50, Height: 20}, "Stone") {
			cellType = cell.STONE_CELL
		}
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 225, Width: 50, Height: 20}, "Fire") {
			cellType = cell.FIRE_CELL
		}
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 250, Width: 50, Height: 20}, "Wood") {
			cellType = cell.WOOD_CELL
		}
		w.Draw()
		vec := rl.GetMousePosition()
		x := uint16(vec.X) / common.SIZE_CELL
		y := uint16(vec.Y) / common.SIZE_CELL
		//avoid to send same package twise
		chunkId := w.GetChunkId(x, y)
		if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
			if !(vec.X > W_GAME || vec.Y > H_GAME) {
				err := common.SendToWebSocketPackages(conn, chain.GetDrawCommand(x, y, cellType, brushType))
				if err != nil {
					fmt.Println(err.Error())
				}
			}
		}
		if rl.IsKeyDown(rl.KeyR) {
			err := common.SendToWebSocketPackages(conn, chain.GetGeneratorCommand(chain.GetDrawCommand(x, y, cellType, brushType))...)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
		rl.DrawText(fmt.Sprintf("x:%v\n y:%v\n c:%v", x, y, chunkId), W_GAME-30, 0, common.SIZE_CELL, rl.Black)
		rl.EndDrawing()
	}

}

func createWebSocket() (*ws.Conn, error) {
	u := url.URL{Scheme: "ws", Host: ":8000", Path: "/ws"}
	conn, _, err := ws.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	err = common.SendToWebSocketPackages(conn, chain.GetInitCommand(8000))
	if err != nil {
		return nil, err
	}
	println("WebSocket connected ", conn.RemoteAddr().String())

	return conn, nil
}

func UpdateWorld(world *world.ClientWorld, udp *ws.Conn) {
	for {
		//2->16bit
		//		var bytes []byte = make([]byte, 2*world.ChunkSize*world.ChunkSize+2)
		_, bytes, err := udp.ReadMessage()
		if err != nil {
			fmt.Println(err)
			continue
		}
		idChunk := binary.BigEndian.Uint16(bytes[0:2])
		world.SetCellsByte(bytes[2:], idChunk)
	}
}
