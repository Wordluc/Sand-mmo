package wasm

import "syscall/js"

func (state *WasmState) AddMouseEventListeners() {
	canvas := state.Document.Call("getElementById", GAME_WINDOW)
	rectDiv := canvas.Call("getBoundingClientRect")
	xStart, yStart := rectDiv.Get("x").Int(), rectDiv.Get("y").Int()
	width := canvas.Get("offsetWidth").Int()
	height := canvas.Get("offsetHeight").Int()

	canvas.Call("addEventListener", "mousemove", js.FuncOf(func(this js.Value, args []js.Value) any {
		state.Mouse.Set(args[0].Get("clientX").Int()-xStart, args[0].Get("clientY").Int()-yStart)
		return nil
	}))
	canvas.Call("addEventListener", "mousedown", js.FuncOf(func(this js.Value, args []js.Value) any {
		state.Mouse.Pressed = true
		return nil
	}))
	canvas.Call("addEventListener", "mouseleave", js.FuncOf(func(this js.Value, args []js.Value) any {
		state.Mouse.Pressed = false
		return nil
	}))
	canvas.Call("addEventListener", "mouseup", js.FuncOf(func(this js.Value, args []js.Value) any {
		state.Mouse.Pressed = false
		return nil
	}))
	canvas.Call("addEventListener", "touchmove", js.FuncOf(func(this js.Value, args []js.Value) any {
		touch := args[0].Get("touches").Index(0)
		x := touch.Get("clientX").Int() - xStart
		y := touch.Get("clientY").Int() - yStart

		if x < 0 || y < 0 || x > width || y > height {
			state.Mouse.Pressed = false
		}

		state.Mouse.Set(x, y)
		return nil
	}))
	canvas.Call("addEventListener", "touchstart", js.FuncOf(func(this js.Value, args []js.Value) any {
		touch := args[0].Get("touches").Call("item", 0)
		state.Mouse.Set(touch.Get("clientX").Int()-xStart, touch.Get("clientY").Int()-yStart)
		state.Mouse.Pressed = true
		return nil
	}))
	releaseFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		state.Mouse.Pressed = false
		return nil
	})
	canvas.Call("addEventListener", "touchend", releaseFn)
	canvas.Call("addEventListener", "touchcancel", releaseFn)
}
