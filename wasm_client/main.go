package main

import (
	"encoding/binary"
	"sand-mmo/cell"
	"sand-mmo/common"
	chain "sand-mmo/responsibilityChain"
	"sand-mmo/world"
	"strings"
	"sync"
	"syscall/js"

	"wasm/utils"
)

var ws js.Value
var ctx js.Value

var frameBuf []byte
var jsDst js.Value
var jsImageData js.Value
var canvasW, canvasH int

func initDrawBuffers() {
	size := int(common.W_WINDOWS) * int(common.H_WINDOWS) * int(common.SIZE_CELL) * int(common.SIZE_CELL) * 4
	frameBuf = make([]byte, size)
	jsDst = js.Global().Get("Uint8ClampedArray").New(size)
	jsImageData = js.Global().Get("ImageData")
	canvasW = int(common.W_WINDOWS) * int(common.SIZE_CELL)
	canvasH = int(common.H_WINDOWS) * int(common.SIZE_CELL)
}

func Draw(w world.ClientWorld) {
	// 4 bytes per pixel (RGBA)

	var i uint16
	for _, c := range w.GetCells() {
		x := int(i%w.W) * common.SIZE_CELL
		y := int(i/w.W) * common.SIZE_CELL
		i++

		// fill SIZE_CELL x SIZE_CELL pixels
		for dy := range common.SIZE_CELL {
			for dx := range common.SIZE_CELL {
				px := ((y+dy)*common.W_WINDOWS*common.SIZE_CELL + (x + dx)) * 4 //?
				frameBuf[px], frameBuf[px+1], frameBuf[px+2], frameBuf[px+3] = w.GetColor(c.CellType).Get()
			}
		}
	}

	// copy entire buffer to JS in ONE call
	js.CopyBytesToJS(jsDst, frameBuf)
	imageData := jsImageData.New(jsDst, canvasW, canvasH)

	ctx.Call("putImageData", imageData, 0, 0)

}

var mouse common.Vec2
var pressed bool
var m sync.Mutex
var brushType common.BrushType
var cellType cell.CellType
var addGenerator int

// Button definitions
type ButtonDef struct {
	label     string
	isBrush   bool
	cellType  cell.CellType
	brushType common.BrushType
}

var buttons = []ButtonDef{
	{label: "Water", isBrush: false, cellType: cell.WATER_CELL},
	{label: "Sand", isBrush: false, cellType: cell.SAND_CELL},
	{label: "Smoke", isBrush: false, cellType: cell.SMOKE_CELL},
	{label: "Small Circle", isBrush: true, brushType: common.CIRCLE_SMALL},
	{label: "Big Circle", isBrush: true, brushType: common.CIRCLE_BIG},
	{label: "Small Square", isBrush: true, brushType: common.SQUARE_SMALL},
	{label: "Big Square", isBrush: true, brushType: common.SQUARE_BIG},
	{label: "Buco nero", isBrush: false, cellType: cell.VACUUM_CELL},
	{label: "Stone", isBrush: false, cellType: cell.STONE_CELL},
	{label: "Fire", isBrush: false, cellType: cell.FIRE_CELL},
	{label: "Wood", isBrush: false, cellType: cell.WOOD_CELL},
	{label: "Lava", isBrush: false, cellType: cell.LAVA_CELL},
	{label: "Foglia", isBrush: false, cellType: cell.LEAF_CELL},
}

// Render buttons using HTML DOM
func renderButtons(buttons []ButtonDef, cellType *cell.CellType, brushType *common.BrushType) {
	js.Global().Get("document").Call("getElementById", "buttons")

	for _, btn := range buttons {
		b := btn // capture
		el := js.Global().Get("document").Call("createElement", "button")
		el.Set("textContent", b.label)

		el.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if b.isBrush {
				*brushType = b.brushType
			} else {
				*cellType = b.cellType
			}
			return nil
		}))

		js.Global().Get("document").Call("getElementById", "buttons").Call("appendChild", el)
	}
}

func registryMouseMovement(document js.Value) {

	document.Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) any {
		m.Lock()
		event := args[0]
		if event.Get("key").String() == "r" && addGenerator == 0 {
			addGenerator = 1
		}
		m.Unlock()
		return nil
	}))
	document.Call("addEventListener", "keyup", js.FuncOf(func(this js.Value, args []js.Value) any {
		m.Lock()
		event := args[0]
		if event.Get("key").String() == "r" {
			addGenerator = 0
		}
		m.Unlock()
		return nil
	}))
	document.Call("addEventListener", "mousemove", js.FuncOf(func(this js.Value, args []js.Value) any {
		m.Lock()
		event := args[0]
		mouse.Set(int32(event.Get("clientX").Int()), int32(event.Get("clientY").Int()))
		m.Unlock()
		return nil
	}))
	document.Call("addEventListener", "mousedown", js.FuncOf(func(this js.Value, args []js.Value) any {
		m.Lock()
		pressed = true
		m.Unlock()
		return nil
	}))
	document.Call("addEventListener", "mouseup", js.FuncOf(func(this js.Value, args []js.Value) any {
		m.Lock()
		pressed = false
		m.Unlock()
		return nil
	}))

}
func main() {
	initDrawBuffers()
	doc := js.Global().Get("document")
	div := doc.Call("getElementById", "GAME_WINDOW")
	div.Set("width", common.SIZE_CELL*common.W_WINDOWS)
	div.Set("height", common.SIZE_CELL*common.H_WINDOWS)
	ctx = div.Call("getContext", "2d")
	registryMouseMovement(doc)
	renderButtons(buttons, &cellType, &brushType)
	w := world.NewClientWorld(common.W_WINDOWS, common.H_WINDOWS, common.CHUNK_SIZE)

	loc := js.Global().Get("location")
	host, _, _ := strings.Cut(loc.Get("host").String(), ":")
	protocol := "ws"
	if loc.Get("protocol").String() == "https:" {
		protocol = "wss"
	}

	wsURL := protocol + "://" + host + ":8000" + "/ws"
	ws = js.Global().Get("WebSocket").New(wsURL)

	ws.Set("binaryType", "arraybuffer")

	var bufferByte utils.Buffer = utils.NewBuffer()

	ws.Set("onopen", js.FuncOf(func(this js.Value, args []js.Value) any {
		send(chain.GetInitCommand())
		return nil
	}))
	js.Global().Get("window").Call("addEventListener", "beforeunload", js.FuncOf(func(this js.Value, args []js.Value) any {
		send(chain.GetENDCommand())
		ws.Call("close")
		return nil
	}))

	ws.Set("onmessage", js.FuncOf(func(this js.Value, args []js.Value) any {
		data := args[0].Get("data")

		buf := make([]byte, data.Get("byteLength").Int())
		js.CopyBytesToGo(buf, js.Global().Get("Uint8Array").New(data))
		chunkId := binary.BigEndian.Uint16(buf[0:2])
		bufferByte.Append(chunkId, buf[2:])
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
	js.Global().Set("goFrame", js.FuncOf(func(this js.Value, args []js.Value) any {
		x, y := mouse.Get()
		x = x / common.SIZE_CELL
		y = y / common.SIZE_CELL

		m.Lock()
		isPressed := pressed
		m.Unlock()

		if addGenerator == 1 {
			send(chain.GetGeneratorCommand(chain.GetDrawCommand(uint16(x), uint16(y), cellType, brushType))...)
			addGenerator = -1
		}
		if isPressed {
			send(chain.GetDrawCommand(uint16(x), uint16(y), cellType, brushType))
		}

		for _, idChunk := range bufferByte.GetChunks() {
			w.SetCellsByte(bufferByte.GetLast(idChunk), idChunk)
		}
		Draw(w)
		return nil
	}))

	select {}
}

func send(ps ...common.Package) {
	for i := range ps {
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, common.Encode(ps[i]))
		dst := js.Global().Get("Uint8Array").New(8)
		js.CopyBytesToJS(dst, buf)
		ws.Call("send", dst)
	}
}
func sendRaw(bytes []byte) {
	dst := js.Global().Get("Uint8Array").New(8)
	js.CopyBytesToJS(dst, bytes)
	ws.Call("send", dst)
}
