package wasm

import (
	"sand-mmo/core"
	"strings"
	"syscall/js"
)

func IsMobile() bool {
	userAgent := js.Global().Get("navigator").Get("userAgent").String()
	userAgent = strings.ToLower(userAgent)
	mobileKeywords := []string{"android", "iphone", "mobile"}
	for _, keyword := range mobileKeywords {
		if strings.Contains(userAgent, keyword) {
			return true
		}
	}
	return false
}
func (state *WasmState) InitWorld() {
	if IsMobile() {
		SIZE_CELL = 2
	}
	state.World = new(core.NewClientWorld())
	doc := js.Global().Get("document")
	div := doc.Call("getElementById", GAME_WINDOW)
	div.Set("width", SIZE_CELL*state.World.W)
	div.Set("height", SIZE_CELL*state.World.H)
	state.Document = doc
	state.Ctx2D = div.Call("getContext", "2d")
	initDrawMemory(state)
}
