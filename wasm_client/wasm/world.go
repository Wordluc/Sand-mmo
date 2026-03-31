package wasm

import (
	"sand-mmo/core"
	"syscall/js"
)

func (state *WasmState) InitWorld() {
	state.World = new(core.NewClientWorld())
	doc := js.Global().Get("document")
	div := doc.Call("getElementById", "GAME_WINDOW")
	div.Set("width", SIZE_CELL*state.World.W)
	div.Set("height", SIZE_CELL*state.World.H)
	state.Document = doc
	state.Ctx2D = div.Call("getContext", "2d")
}
