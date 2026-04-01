package main

import (
	"encoding/binary"
	"sand-mmo/common"
	"sand-mmo/core/handlers"
	"strings"
	"syscall/js"
	"wasm/utils"
	"wasm/wasm"
)

var ctx js.Value

var state *wasm.WasmState = new(wasm.NewState())
var bufferByte = utils.NewBuffer()

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

func setChunksIntoWorld(chunks []int) {
	xClient, yClient := state.Window.Pos.Get()
	var idChunk, fixedChunkId int
	if len(chunks) != 0 {
		for _, idChunk = range chunks {
			x, y := common.GetServerXYChunk(idChunk)
			x = x - xClient
			y = y - yClient
			fixedChunkId = x + y*common.W_CHUNKS_CLIENT
			if !state.World.IsChunkIdValid(fixedChunkId) {
				continue
			}

			state.World.SetDecodedCells(bufferByte.GetLast(idChunk), fixedChunkId)
		}
	}
}

func main() {
	state.InitCarosello()
	state.InitWorld()
	state.AddMouseEventListeners()
	state.AddKeyboardEventListeners()
	state.InitWebSocket(getAddr())

	state.WebSocket.Set("onmessage", js.FuncOf(func(this js.Value, args []js.Value) any {
		data := args[0].Get("data")

		buf := make([]byte, data.Get("byteLength").Int())
		js.CopyBytesToGo(buf, js.Global().Get("Uint8Array").New(data))
		gChunkId := int(binary.BigEndian.Uint16(buf[0:2]))

		bufferByte.Append(gChunkId, buf[2:])
		return nil
	}))

	js.Global().Set("changeBrushSize", js.FuncOf(func(this js.Value, args []js.Value) any {
		size := args[0].Get("target").Get("value").String()
		state.Brush.BrushSize = size
		return nil
	},
	))
	js.Global().Set("changeBrushShape", js.FuncOf(func(this js.Value, args []js.Value) any {
		shape := args[0].Get("target").Get("value").String()
		state.Brush.BrushShape = shape
		return nil
	},
	))
	js.Global().Set("goFrame", js.FuncOf(func(this js.Value, args []js.Value) any {
		x, y := state.Mouse.Get()
		go func() {
			x = x / wasm.SIZE_CELL
			y = y / wasm.SIZE_CELL
			if state.Brush.AddGenerator == 1 {
				wasm.Send(state.WebSocket, handlers.GetGeneratorCommand(handlers.GetDrawCommand(x, y, state.CellType, state.Brush.GetBrushType()))...)
				state.Brush.AddGenerator = -1
			}
			if state.Mouse.Pressed {
				wasm.Send(state.WebSocket, handlers.GetDrawCommand(x, y, state.CellType, state.Brush.GetBrushType()))
			}
		}()
		offset := state.Window.Pos.Copy()
		offset.Sub(state.Window.OldPos)

		if !offset.IsZero() {
			bufferByte.Clean()
			state.World.ShiftWorld(offset.Get())
			wasm.Send(state.WebSocket, handlers.GetMoveCommand(uint16(state.Window.GetChunkId())))
			wasm.Draw(state)
			state.Window.OldPos = state.Window.Pos
		}
		setChunksIntoWorld(bufferByte.GetChunks())

		wasm.Draw(state)

		return nil
	}))

	select {}
}
