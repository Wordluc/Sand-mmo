package main

import (
	"encoding/binary"
	"sand-mmo/common"
	"sand-mmo/core/handlers"
	"syscall/js"
	"wasm/utils"
	"wasm/wasm"
)

var ctx js.Value

var state *wasm.WasmState = new(wasm.NewState())
var bufferByte = utils.NewBuffer()

func main() {
	var idChunk int
	state.InitCarosello()
	state.InitWorld()
	state.AddMouseEventListeners()
	state.AddKeyboardEventListeners()
	state.InitWebSocket()

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
				wasm.Send(state.WebSocket, handlers.GetGeneratorCommand(handlers.GetDrawCommand(uint16(x), uint16(y), state.CellType, state.Brush.GetBrushType()))...)
				state.Brush.AddGenerator = -1
			}
			if state.Mouse.Pressed {
				wasm.Send(state.WebSocket, handlers.GetDrawCommand(uint16(x), uint16(y), state.CellType, state.Brush.GetBrushType()))
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
			//return nil
		}
		chunks := bufferByte.GetChunks()
		var toDraw = []int{}
		xClient, yClient := state.Window.Pos.Get()

		if len(chunks) != 0 {
			for _, idChunk = range chunks {
				x, y := common.GetServerXYChunk(idChunk)
				x = x - xClient
				y = y - yClient
				if x < 0 || x >= common.W_CHUNKS_CLIENT {
					continue
				}
				if y < 0 || y >= common.H_CHUNKS_CLIENT {
					continue
				}

				state.World.SetDecodedCells(bufferByte.GetLast(idChunk), x+y*common.W_CHUNKS_CLIENT)
				toDraw = append(toDraw, x+y*common.W_CHUNKS_CLIENT)
			}
		}
		wasm.Draw(state)

		return nil
	}))

	select {}
}
