package wasm

import (
	"encoding/binary"
	"sand-mmo/common"
	"sand-mmo/core/handlers"
	"strings"
	"syscall/js"
)

func (state *WasmState) InitWebSocket() {
	loc := js.Global().Get("location")
	host, _, _ := strings.Cut(loc.Get("host").String(), ":")
	protocol := "ws"
	if loc.Get("protocol").String() == "https:" {
		protocol = "wss"
	}

	wsURL := protocol + "://" + host + ":8000" + "/ws"
	//wsURL := protocol + "://" + "www.wordluc.it" + ":8000" + "/ws"
	ws := js.Global().Get("WebSocket").New(wsURL)

	ws.Set("binaryType", "arraybuffer")
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

func Send(ws js.Value, ps ...common.Package) {
	for i := range ps {
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, common.Encode(ps[i]))
		dst := js.Global().Get("Uint8Array").New(8)
		js.CopyBytesToJS(dst, buf)
		ws.Call("send", dst)
	}
}

func SendRaw(ws js.Value, bytes []byte) {
	dst := js.Global().Get("Uint8Array").New(8)
	js.CopyBytesToJS(dst, bytes)
	ws.Call("send", dst)
}
