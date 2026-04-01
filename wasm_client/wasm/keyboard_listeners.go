package wasm

import "syscall/js"

func (state *WasmState) AddKeyboardEventListeners() {

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
			state.Window.SetX(-1)
		}
		if args[0].Get("key").String() == "d" {
			state.Window.SetX(1)
		}
		if args[0].Get("key").String() == "w" {
			state.Window.SetY(-1)
		}
		if args[0].Get("key").String() == "s" {
			state.Window.SetY(1)
		}
		return nil
	}))
}
