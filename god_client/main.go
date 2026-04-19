package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sand-mmo/common"
	"sand-mmo/core"
	"sand-mmo/core/handlers"
	"time"

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
	var cellType core.CellType = core.SAND_CELL
	var brushType common.BrushType = common.CIRCLE_SMALL
	var currentPlayers int
	var responseStr common.ResponseMetadati
	var poolingMetadati common.Scheduler
	poolingMetadati = common.NewScheduler(1000, "PollingMetadati", func() {
		res, err := http.Get(fmt.Sprintf("http://%v:%v/metadati", addr, port))
		if err != nil {
			println(err.Error())
			poolingMetadati.Stop()
			return
		}
		resp, err := io.ReadAll(res.Body)
		if err != nil {
			println(err.Error())
			poolingMetadati.Stop()
			return
		}
		json.Unmarshal(resp, &responseStr)
		currentPlayers = responseStr.NClients
	})
	poolingMetadati.Start()
	defer poolingMetadati.Stop()
	for {
		if rl.WindowShouldClose() {
			return
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.SkyBlue)
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 0, Width: 50, Height: 20}, "Water") {
			cellType = core.WATER_CELL
		}
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 25, Width: 50, Height: 20}, "Sand") {
			cellType = core.SAND_CELL
		}
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 50, Width: 50, Height: 20}, "Smoke") {
			cellType = core.SMOKE_CELL
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
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 175, Width: 50, Height: 20}, "Void") {
			cellType = core.VOID_CELL
		}
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 200, Width: 50, Height: 20}, "Stone") {
			cellType = core.STONE_CELL
		}
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 225, Width: 50, Height: 20}, "Fire") {
			cellType = core.FIRE_CELL
		}
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 250, Width: 50, Height: 20}, "Wood") {
			cellType = core.WOOD_CELL
		}
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 275, Width: 50, Height: 20}, "Lava") {
			cellType = core.LAVA_CELL
		}
		if ru.Button(rl.Rectangle{X: W_GAME + 5, Y: 300, Width: 50, Height: 20}, "Leaf") {
			cellType = core.LEAF_CELL
		}

		Draw(w)
		vec := rl.GetMousePosition()
		x := int(vec.X) / SIZE_CELL
		y := int(vec.Y) / SIZE_CELL
		//avoid to send same package twise
		if rl.IsMouseButtonDown(rl.MouseButtonLeft) && w.IsCoordinateValid(x, y) {
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
			common.SendToWebSocketPackages(conn, handlers.GetGeneratorCommand(handlers.GetDrawCommand(x, y, cellType, brushType))...)
		}
		rl.DrawText(fmt.Sprintf("CurrentPlayer: %v", currentPlayers), W_GAME-40, 350, common.SIZE_CELL, rl.Black)
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
	conn.SetReadLimit(int64(common.CHUNK_BYTES_SIZE) * common.W_CHUNKS_TOTAL * common.H_CHUNKS_TOTAL)
	println("WebSocket connected ")

	return conn, nil
}

func UpdateWorld(world *core.ClientWorld, webSocket *ws.Conn) {
	var offset = 0
	var idChunk uint16
	var x, y int

	for {
		ctx, _ := context.WithTimeout(context.Background(), 500*time.Second)
		_, bytes, err := webSocket.Read(ctx)
		offset = 0
		if err != nil {
			fmt.Println(err)
			continue
		}
		for offset < len(bytes) {
			idChunk = binary.BigEndian.Uint16(bytes[offset : offset+2])
			x, y = common.GetServerXYChunk(int(idChunk))
			world.SetDecodedCells(bytes[offset+2:offset+common.CHUNK_BYTES_SIZE], x, y)
			offset += common.CHUNK_BYTES_SIZE
		}

	}
}

func Draw(w core.ClientWorld) {
	for _, chunkId := range w.PopActiveChunks() {
		w.ForEachCell(chunkId, func(x, y int, center *core.Cell) error {
			x = x * SIZE_CELL
			y = y * SIZE_CELL
			rl.DrawRectangle(int32(x), int32(y), SIZE_CELL, SIZE_CELL, rl.NewColor(center.GetColor().Get()))
			return nil
		})
	}
}
