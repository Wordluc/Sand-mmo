package wasm

import "syscall/js"

func (state *WasmState) AddMouseEventListeners() {
	rectDiv := state.Document.Call("getElementById", GAME_WINDOW).Call("getBoundingClientRect")
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
	state.Document.Call("addEventListener", "touchmove", js.FuncOf(func(this js.Value, args []js.Value) any {
		touch := args[0].Get("touches").Call("item", 0)
		state.Mouse.Set(touch.Get("clientX").Int()-xStart, touch.Get("clientY").Int()-yStart)
		return nil
	}))
	state.Document.Call("addEventListener", "touchstart", js.FuncOf(func(this js.Value, args []js.Value) any {
		touch := args[0].Get("touches").Call("item", 0)
		state.Mouse.Set(touch.Get("clientX").Int()-xStart, touch.Get("clientY").Int()-yStart)
		state.Mouse.Pressed = true
		return nil
	}))
	releaseFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		state.Mouse.Pressed = false
		return nil
	})
	state.Document.Call("addEventListener", "touchend", releaseFn)
	state.Document.Call("addEventListener", "touchcancel", releaseFn)
}
