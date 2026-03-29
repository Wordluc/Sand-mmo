package wasm

import "syscall/js"

func (state *WasmState) AddMouseEventListeners() {
	rectDiv := state.Document.Call("getElementById", "GAME_WINDOW").Call("getBoundingClientRect")
	xStart, yStart := rectDiv.Get("x").Int(), rectDiv.Get("y").Int()

	state.Document.Call("addEventListener", "mousemove", js.FuncOf(func(this js.Value, args []js.Value) any {
		state.Mouse.Set(args[0].Get("clientX").Int()-xStart, args[0].Get("clientY").Int()-yStart)
		return nil
	}))
	state.Document.Call("addEventListener", "mousedown", js.FuncOf(func(this js.Value, args []js.Value) any {
		state.Mouse.Pressed = true
		return nil
	}))
	state.Document.Call("addEventListener", "mouseup", js.FuncOf(func(this js.Value, args []js.Value) any {
		state.Mouse.Pressed = false
		return nil
	}))
}
