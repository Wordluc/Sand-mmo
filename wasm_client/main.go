package main

import (
	"encoding/binary"
	"sand-mmo/common"
	"sand-mmo/core/handlers"
	"strings"
	"syscall/js"
	"wasm/wasm"
)

var ctx js.Value

var state *wasm.WasmState = new(wasm.NewState())

func getAddr() string {
	loc := js.Global().Get("location")
	host, _, _ := strings.Cut(loc.Get("host").String(), ":")
	protocol := "ws"
	if loc.Get("protocol").String() == "https:" {
		protocol = "wss"
	}

	//wsURL := protocol + "://" + "www.wordluc.it" + ":8000" + "/ws"
	return protocol + "://" + host + ":8000" + "/ws"
}

func loadChunksIntoWorld() {
	xClient, yClient := state.Window.Pos.Get()

	jsBuffer := js.Global().Call("get_all_chunks_binary")
	length := jsBuffer.Get("byteLength").Int()
	if length == 0 {
		return
	}

	raw := make([]byte, length)
	js.CopyBytesToGo(raw, jsBuffer)

	chunkSize := 2 + common.CHUNK_SIZE*common.CHUNK_SIZE*2
	var idChunk uint16
	for offset := 0; offset < len(raw); offset += chunkSize {
		idChunk = binary.BigEndian.Uint16(raw[offset : offset+2])
		cellData := raw[offset+2 : offset+chunkSize]

		x, y := common.GetServerXYChunk(int(idChunk))
		x -= xClient
		y -= yClient

		if x < 0 || x >= common.W_CHUNKS_CLIENT {
			continue
		}
		if y < 0 || y >= common.H_CHUNKS_CLIENT {
			continue
		}

		state.World.SetDecodedCells(cellData, x+y*common.W_CHUNKS_CLIENT)
	}
}

func main() {
	state.InitCarosello()
	state.InitWorld()
	state.AddMouseEventListeners()
	state.AddKeyboardEventListeners()
	state.InitWebSocket(getAddr())

	js.Global().Set("setGenerator", js.FuncOf(func(this js.Value, args []js.Value) any {
		state.Brush.AddGenerator = 1
		return nil
	}))

	js.Global().Set("changeBrushSize", js.FuncOf(func(this js.Value, args []js.Value) any {
		size := args[0].Get("target").Get("value").String()
		state.Brush.BrushSize = size
		return nil
	}))
	js.Global().Set("changeBrushShape", js.FuncOf(func(this js.Value, args []js.Value) any {
		shape := args[0].Get("target").Get("value").String()
		state.Brush.BrushShape = shape
		return nil
	}))
	js.Global().Set("moveView", js.FuncOf(func(this js.Value, args []js.Value) any {
		x := args[0].Int()
		y := args[1].Int()
		state.Window.Offset.Set(x, y)
		return nil
	}))
	move := wasm.Throttle(100, func() {
		offset := state.Window.Offset
		if !offset.IsZero() {

			newPos := state.Window.Pos.Copy()
			newPos.Add(offset)
			x, y := newPos.Get()
			if !(x < 0 || x+common.W_CHUNKS_CLIENT > common.W_CHUNKS_TOTAL || y < 0 || y+common.H_CHUNKS_CLIENT > common.H_CHUNKS_TOTAL) {
				state.Window.Pos = newPos
				js.Global().Call("clear_all_queued_chunks")
				state.World.ShiftWorld(offset.Get())

				wasm.Send(state.WebSocket, handlers.GetMoveCommand(uint16(state.Window.GetChunkId())))
			}
			state.Window.Offset.Set(0, 0)
		}
	})
	var x, y int
	js.Global().Set("goFrame", js.FuncOf(func(this js.Value, args []js.Value) any {
		x, y = state.Mouse.Get()
		x = x / wasm.SIZE_CELL
		y = y / wasm.SIZE_CELL
		if state.Mouse.Pressed {
			if state.Brush.AddGenerator == 1 {
				wasm.Send(state.WebSocket, handlers.GetGeneratorCommand(handlers.GetDrawCommand(x, y, state.CellType, state.Brush.GetBrushType()))...)
				state.Brush.AddGenerator = -1
			} else {
				wasm.Send(state.WebSocket, handlers.GetDrawCommand(x, y, state.CellType, state.Brush.GetBrushType()))
			}
		}
		loadChunksIntoWorld()
		move()

		wasm.Draw(state)

		return nil
	}))

	select {}
}
