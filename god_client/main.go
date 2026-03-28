package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"sand-mmo/cell"
	"sand-mmo/common"
	"sand-mmo/core"
	"sand-mmo/core/handlers"

	ws "github.com/coder/websocket"
	ru "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const W_BUTTONS_SIDE = 100
const SIZE_CELL = 1
const W_GAME = common.W_CELLS_TOTAL * SIZE_CELL
const H_GAME = common.H_CELLS_TOTAL * SIZE_CELL

var moved = false

func main() {
	rl.InitWindow(W_GAME+W_BUTTONS_SIDE, H_GAME+common.SIZE_CELL, "")
	w := core.NewCustomWorld(common.W_CELLS_TOTAL, common.H_CELLS_TOTAL, common.CHUNK_SIZE)
	addr := os.Args[1]
	port := os.Args[2]
	println("Attempting connecting ", addr, ":", port)
	conn, err := createWebSocket(addr, port)
	if err != nil {
		panic(err)
	}

	go UpdateWorld(&w, conn)

	defer conn.CloseNow()
	defer common.SendToWebSocketPackages(conn, handlers.GetENDCommand())

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
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 175, Width: 50, Height: 20}, "Vacuum Cleaner") {
			cellType = cell.VACUUM_CELL
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
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 275, Width: 50, Height: 20}, "Lava") {
			cellType = cell.LAVA_CELL
		}
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 300, Width: 50, Height: 20}, "Leaf") {
			cellType = cell.LEAF_CELL
		}

		Draw(w)
		vec := rl.GetMousePosition()
		x := uint16(vec.X) / SIZE_CELL
		y := uint16(vec.Y) / SIZE_CELL
		//avoid to send same package twise
		chunkId := w.GetChunkId(int(x), int(y))
		if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
			if !(vec.X > W_GAME || vec.Y > H_GAME) {
				err := common.SendToWebSocketPackages(conn, handlers.GetDrawCommand(x, y, cellType, brushType))
				if err != nil {
					fmt.Println(err.Error())
				}
			}
		}
		if rl.IsKeyDown(rl.KeyR) {
			err := common.SendToWebSocketPackages(conn, handlers.GetGeneratorCommand(handlers.GetDrawCommand(x, y, cellType, brushType))...)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
		if moved {
			common.SendToWebSocketPackages(conn, handlers.GetGeneratorCommand(handlers.GetDrawCommand(uint16(x), uint16(y), cellType, brushType))...)
		}
		rl.DrawText(fmt.Sprintf("x:%v\n y:%v\n c:%v", x, y, chunkId), W_GAME-30, 0, common.SIZE_CELL, rl.Black)
		rl.EndDrawing()
	}

}

func createWebSocket(addr, port string) (*ws.Conn, error) {
	conn, _, err := ws.Dial(context.Background(), fmt.Sprintf("ws://%v:%v/ws", addr, port), nil)
	if err != nil {
		return nil, err
	}
	err = common.SendToWebSocketPackages(conn, handlers.GetInitGODCommand())
	if err != nil {
		return nil, err
	}
	println("WebSocket connected ")

	return conn, nil
}

func UpdateWorld(world *core.ClientWorld, webSocket *ws.Conn) {
	for {
		_, bytes, err := webSocket.Read(context.Background())
		if err != nil {
			fmt.Println(err)
			continue
		}

		idChunk := binary.BigEndian.Uint16(bytes[0:2])
		world.SetDecodedCells(bytes[2:], int(idChunk))

	}
}

func Draw(w core.ClientWorld) {
	for _, chunkId := range w.GetChunksToDraw() {
		w.ForEachCell(chunkId, func(x, y int, center *cell.Cell) error {
			x = x * SIZE_CELL
			y = y * SIZE_CELL
			rl.DrawRectangle(int32(x), int32(y), SIZE_CELL, SIZE_CELL, rl.NewColor(center.GetColor().Get()))
			return nil
		})
	}
}
