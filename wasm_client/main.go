package main

import (
	"encoding/binary"
	"sand-mmo/common"
	chain "sand-mmo/responsibilityChain"
	"sand-mmo/world"
	"syscall/js"
	"time"
)

var ws js.Value
var ctx js.Value

func Draw(w world.ClientWorld) {
	// 4 bytes per pixel (RGBA)
	buf := make([]byte, common.W_WINDOWS*common.H_WINDOWS*common.SIZE_CELL*common.SIZE_CELL*4)

	var i uint16
	for _, c := range w.GetCells() {
		x := int(i%w.W) * common.SIZE_CELL
		y := int(i/w.W) * common.SIZE_CELL
		i++

		// fill SIZE_CELL x SIZE_CELL pixels
		for dy := range common.SIZE_CELL {
			for dx := range common.SIZE_CELL {
				px := ((y+dy)*common.W_WINDOWS*common.SIZE_CELL + (x + dx)) * 4 //?
				buf[px], buf[px+1], buf[px+2], buf[px+3] = w.GetColor(c.CellType).Get()
			}
		}
	}

	// copy entire buffer to JS in ONE call
	dst := js.Global().Get("Uint8ClampedArray").New(len(buf))
	js.CopyBytesToJS(dst, buf)
	imageData := js.Global().Get("ImageData").New(dst,
		common.W_WINDOWS*common.SIZE_CELL,
		common.H_WINDOWS*common.SIZE_CELL,
	)
	ctx.Call("putImageData", imageData, 0, 0)
}
func main() {
	doc := js.Global().Get("document")
	div := doc.Call("getElementById", "GAME_WINDOW")
	div.Set("width", common.SIZE_CELL*common.W_WINDOWS)
	div.Set("height", common.SIZE_CELL*common.H_WINDOWS)
	ctx = div.Call("getContext", "2d")

	w := world.NewClientWorld(common.W_WINDOWS, common.H_WINDOWS, common.CHUNK_SIZE)
	ws = js.Global().Get("WebSocket").New("ws://localhost:8000/ws")
	ws.Set("binaryType", "arraybuffer")

	ws.Set("onopen", js.FuncOf(func(this js.Value, args []js.Value) any {
		send(common.Encode(chain.GetInitCommand(8000)))
		return nil
	}))
	js.Global().Get("window").Call("addEventListener", "beforeunload", js.FuncOf(func(this js.Value, args []js.Value) any {
		send(common.Encode(chain.GetENDCommand()))
		ws.Call("close")
		return nil
	}))

	ws.Set("onmessage", js.FuncOf(func(this js.Value, args []js.Value) any {
		data := args[0].Get("data")

		buf := make([]byte, data.Get("byteLength").Int())
		js.CopyBytesToGo(buf, js.Global().Get("Uint8Array").New(data))

		chunkId := binary.BigEndian.Uint16(buf[0:2])
		cells := js.Global().Get("Uint8Array").New(len(buf) - 2)
		js.CopyBytesToJS(cells, buf[2:])
		w.SetCellsByte(buf[2:], chunkId)
		return nil
	}))

	ws.Set("onclose", js.FuncOf(func(this js.Value, args []js.Value) any {
		println("WebSocket closed")
		return nil
	}))

	ws.Set("onerror", js.FuncOf(func(this js.Value, args []js.Value) any {
		println("WebSocket error")
		return nil
	}))
	sleepFor := 1000.0 / 30.0
	go func() {
		for {
			time.Sleep(time.Duration(sleepFor) * time.Millisecond)
			Draw(w)
		}
	}()
	select {}
}

func send(encoded uint64) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, encoded)
	dst := js.Global().Get("Uint8Array").New(8)
	js.CopyBytesToJS(dst, buf)
	ws.Call("send", dst)
}
