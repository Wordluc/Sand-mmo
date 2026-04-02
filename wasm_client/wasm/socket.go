package wasm

import (
	"encoding/binary"
	"sand-mmo/common"
	"sand-mmo/core/handlers"
	"syscall/js"
)

func (state *WasmState) InitWebSocket(addr string) {
	ws := js.Global().Get("WebSocket").New(addr)
	js.Global().Call("setup_websocket", ws)

	state.WebSocket = ws
	state.WebSocket.Set("onopen", js.FuncOf(func(this js.Value, args []js.Value) any {
		Send(state.WebSocket, handlers.GetInitCommand(state.Window.GetChunkId()))
		return nil
	}))
	state.WebSocket.Set("onclose", js.FuncOf(func(this js.Value, args []js.Value) any {
		Send(state.WebSocket, handlers.GetENDCommand())
		println("WebSocket closed")
		return nil
	}))
	state.WebSocket.Set("onerror", js.FuncOf(func(this js.Value, args []js.Value) any {
		Send(state.WebSocket, handlers.GetENDCommand())
		println("WebSocket error")
		return nil
	}))
	js.Global().Get("window").Call("addEventListener", "beforeunload", js.FuncOf(func(this js.Value, args []js.Value) any {
		Send(state.WebSocket, handlers.GetENDCommand())
		state.WebSocket.Call("close")
		return nil
	}))
}

var sendBuf = make([]byte, 8)
var sendDst = js.Global().Get("Uint8Array").New(8)

func Send(ws js.Value, ps ...common.Package) {
	for i := range ps {
		binary.BigEndian.PutUint64(sendBuf, common.Encode(ps[i]))
		js.CopyBytesToJS(sendDst, sendBuf)
		ws.Call("send", sendDst)
	}
}
