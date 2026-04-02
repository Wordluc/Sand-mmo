package main

import (
	"encoding/binary"
	"sand-mmo/common"
	"sand-mmo/core/handlers"
	"strings"
	"sync"
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
			//DROD chunks that arent in the view anymore
			if x < 0 || x >= common.W_CHUNKS_CLIENT {
				continue
			}
			//DROD chunks that arent in the view anymore
			if y < 0 || y >= common.H_CHUNKS_CLIENT {
				continue
			}

			fixedChunkId = x + y*common.W_CHUNKS_CLIENT

			state.World.SetDecodedCells(bufferByte.GetLast(idChunk), fixedChunkId)
		}
	}
}

var socketMessagePool = sync.Pool{
	New: func() any {
		return js.Value{}
	},
}

func main() {
	state.InitCarosello()
	state.InitWorld()
	state.AddMouseEventListeners()
	state.AddKeyboardEventListeners()
	state.InitWebSocket(getAddr())
	state.WebSocket.Set("onmessage", js.FuncOf(func(this js.Value, args []js.Value) any {
		message := socketMessagePool.Get().(js.Value)
		message = args[0].Get("data")

		buf := make([]byte, message.Get("byteLength").Int())
		js.CopyBytesToGo(buf, js.Global().Get("Uint8Array").New(message))

		gChunkId := int(binary.BigEndian.Uint16(buf[0:2]))
		bufferByte.Append(gChunkId, buf[2:])

		socketMessagePool.Put(message)
		return nil
	}))

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
	var offset common.Vec2
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
		setChunksIntoWorld(bufferByte.GetChunks())

		offset = state.Window.Offset
		if !offset.IsZero() {

			newPos := state.Window.Pos.Copy()
			newPos.Add(offset)
			x, y := newPos.Get()
			if !(x < 0 || x+common.W_CHUNKS_CLIENT > common.W_CHUNKS_TOTAL || y < 0 || y+common.H_CHUNKS_CLIENT > common.H_CHUNKS_TOTAL) {
				state.Window.Pos = newPos

				bufferByte.Clean()
				state.World.ShiftWorld(offset.Get())

				wasm.Send(state.WebSocket, handlers.GetMoveCommand(uint16(state.Window.GetChunkId())))
			}
			state.Window.Offset.Set(0, 0)
		}
		wasm.Draw(state)

		return nil
	}))

	select {}
}
