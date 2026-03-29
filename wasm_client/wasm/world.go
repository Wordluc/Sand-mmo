package wasm

import (
	"sand-mmo/common"
	"sand-mmo/core"
	"syscall/js"
)

func (state *WasmState) InitWorld() {
	state.World = new(core.NewClientWorld())
	doc := js.Global().Get("document")
	div := doc.Call("getElementById", "GAME_WINDOW")
	div.Set("width", SIZE_CELL*common.W_CELLS_CLIENT)
	div.Set("height", SIZE_CELL*common.H_CELLS_CLIENT)
	state.Document = doc
	state.Ctx2D = div.Call("getContext", "2d")
}
