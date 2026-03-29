package wasm

import "syscall/js"

func (state *WasmState) AddKeyboardEventListeners() {
	const t = 100
	moveA := throttle(t, func() {
		state.Window.AddX(-1)
	})
	moveD := throttle(t, func() {
		state.Window.AddX(1)
	})
	moveW := throttle(t, func() {
		state.Window.AddY(-1)

	})
	moveS := throttle(t, func() {
		state.Window.AddY(1)
	})

	state.Document.Call("addEventListener", "keyup", js.FuncOf(func(this js.Value, args []js.Value) any {
		if args[0].Get("key").String() == "r" {
			state.Brush.AddGenerator = 0
		}
		return nil
	}))
	state.Document.Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) any {
		if args[0].Get("key").String() == "r" && state.Brush.AddGenerator == 0 {
			state.Brush.AddGenerator = 1
		}
		if args[0].Get("key").String() == "a" {
			moveA()
		}
		if args[0].Get("key").String() == "d" {
			moveD()
		}
		if args[0].Get("key").String() == "w" {
			moveW()
		}
		if args[0].Get("key").String() == "s" {
			moveS()
		}
		return nil
	}))
}
