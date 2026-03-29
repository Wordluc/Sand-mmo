package main

import (
	"encoding/binary"
	"fmt"
	"sand-mmo/cell"
	"sand-mmo/common"
	"sand-mmo/core"
	"sand-mmo/core/handlers"
	"strings"
	"sync"
	"syscall/js"
	"time"

	"wasm/utils"
)

var ws js.Value
var ctx js.Value

var frameBuf []byte
var jsDst js.Value
var jsImageData js.Value
var canvasW, canvasH int
var bufferByte utils.Buffer = utils.NewBuffer()

const SIZE_CELL = 3

func initDrawBuffers() {
	size := int(common.W_CELLS_CLIENT) * int(common.H_CELLS_CLIENT) * int(SIZE_CELL) * int(SIZE_CELL) * 4
	frameBuf = make([]byte, size)
	jsDst = js.Global().Get("Uint8ClampedArray").New(size)
	canvasW = int(common.W_CELLS_CLIENT) * int(SIZE_CELL)
	canvasH = int(common.H_CELLS_CLIENT) * int(SIZE_CELL)
	jsImageData = js.Global().Get("ImageData").New(jsDst, canvasW, canvasH)
}

func DrawAll(w core.ClientWorld) {
	var dx, dy, px int
	var color common.Color
	for chunkId := range w.GetNumberChucks() {
		w.ForEachCell(chunkId, func(x, y int, center *cell.Cell) error {
			x = x * SIZE_CELL
			y = y * SIZE_CELL
			color = center.GetColor()
			for dy = range SIZE_CELL {
				for dx = range SIZE_CELL {
					px = ((y+dy)*common.W_CELLS_CLIENT*SIZE_CELL + (x + dx)) * 4
					frameBuf[px], frameBuf[px+1], frameBuf[px+2], frameBuf[px+3] = color.Get()
				}
			}
			return nil
		})
	}

	js.CopyBytesToJS(jsDst, frameBuf)

	ctx.Call("putImageData", jsImageData, 0, 0)

}

func Draw(w core.ClientWorld, chunksId []int) {
	var dx, dy, px int
	var color common.Color
	for _, chunkId := range chunksId {
		w.ForEachCell(chunkId, func(x, y int, center *cell.Cell) error {
			x = x * SIZE_CELL
			y = y * SIZE_CELL
			color = center.GetColor()
			for dy = range SIZE_CELL {
				for dx = range SIZE_CELL {
					px = ((y+dy)*common.W_CELLS_CLIENT*SIZE_CELL + (x + dx)) * 4
					frameBuf[px], frameBuf[px+1], frameBuf[px+2], frameBuf[px+3] = color.Get()
				}
			}
			return nil
		})
	}

	js.CopyBytesToJS(jsDst, frameBuf)

	ctx.Call("putImageData", jsImageData, 0, 0)

}

var mouse common.Vec2
var pressed bool
var m sync.Mutex
var brushSize string = "small"
var brushShape string = "circle"
var cellType cell.CellType = cell.SAND_CELL
var addGenerator int
var xClient int
var yClient = common.H_CHUNKS_TOTAL - common.H_CHUNKS_CLIENT

var oldXClient int
var oldYClient int = yClient

// Button definitions
type ButtonDef struct {
	label    string
	cellType cell.CellType
}

var buttons = []ButtonDef{
	{label: "Vacuum Cleaner", cellType: cell.VACUUM_CELL},
	{label: "Water", cellType: cell.WATER_CELL},
	{label: "Sand", cellType: cell.SAND_CELL},
	{label: "Wood", cellType: cell.WOOD_CELL},
	{label: "Leaf", cellType: cell.LEAF_CELL},
	{label: "Stone", cellType: cell.STONE_CELL},
	{label: "Smoke", cellType: cell.SMOKE_CELL},
	{label: "Fire", cellType: cell.FIRE_CELL},
	{label: "Lava", cellType: cell.LAVA_CELL},
}

// Render buttons using HTML DOM
func renderButtons(buttons []ButtonDef, cellType *cell.CellType) {

	doc := js.Global().Get("document")
	elemContainer := doc.Call("getElementById", "buttons-elements")

	// Labels
	elemLabel := doc.Call("createElement", "p")
	elemLabel.Set("textContent", "Elements")
	elemContainer.Call("appendChild", elemLabel)

	brushLabel := doc.Call("createElement", "p")
	brushLabel.Set("textContent", "Brush")

	for _, btn := range buttons {
		b := btn
		el := doc.Call("createElement", "button")
		el.Set("textContent", b.label)

		cls := "snes-button"
		el.Set("className", cls)
		elemContainer.Call("appendChild", el)

		el.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			*cellType = b.cellType
			return nil
		}))
	}
}

var move = false

func registryMouseMovement(document js.Value, w *core.ClientWorld) {
	rectDiv := document.Call("getElementById", "GAME_WINDOW").Call("getBoundingClientRect")
	xStart, yStart := rectDiv.Get("x").Int(), rectDiv.Get("y").Int()
	const t = 100
	moveA := throttle(t, func() {
		xClient--
		if xClient <= 0 {
			xClient = 0
		}
		move = true
	})
	moveD := throttle(t, func() {
		xClient++
		if xClient >= common.W_CHUNKS_TOTAL-common.W_CHUNKS_CLIENT {
			xClient = common.W_CHUNKS_TOTAL - common.W_CHUNKS_CLIENT
		}
		move = true
	})
	moveW := throttle(t, func() {
		yClient -= 1
		if yClient <= 0 {
			yClient = 0
		}
		move = true

	})
	moveS := throttle(t, func() {

		yClient += 1
		if yClient >= common.H_CHUNKS_TOTAL-common.H_CHUNKS_CLIENT {
			yClient = common.H_CHUNKS_TOTAL - common.H_CHUNKS_CLIENT
		}
		move = true
	})

	document.Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) any {
		m.Lock()
		defer m.Unlock()
		if args[0].Get("key").String() == "r" && addGenerator == 0 {
			addGenerator = 1
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
	document.Call("addEventListener", "keyup", js.FuncOf(func(this js.Value, args []js.Value) any {
		m.Lock()
		if args[0].Get("key").String() == "r" {
			addGenerator = 0
		}
		m.Unlock()
		return nil
	}))
	document.Call("addEventListener", "mousemove", js.FuncOf(func(this js.Value, args []js.Value) any {
		m.Lock()
		mouse.Set(args[0].Get("clientX").Int()-xStart, args[0].Get("clientY").Int()-yStart)
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
	div.Set("width", SIZE_CELL*common.W_CELLS_CLIENT)
	div.Set("height", SIZE_CELL*common.H_CELLS_CLIENT)
	ctx = div.Call("getContext", "2d")
	var idChunk int
	renderButtons(buttons, &cellType)
	w := core.NewClientWorld()
	registryMouseMovement(doc, &w)

	loc := js.Global().Get("location")
	host, _, _ := strings.Cut(loc.Get("host").String(), ":")
	protocol := "ws"
	if loc.Get("protocol").String() == "https:" {
		protocol = "wss"
	}

	wsURL := protocol + "://" + host + ":8000" + "/ws"
	//wsURL := protocol + "://" + "www.wordluc.it" + ":8000" + "/ws"
	ws = js.Global().Get("WebSocket").New(wsURL)

	ws.Set("binaryType", "arraybuffer")

	ws.Set("onopen", js.FuncOf(func(this js.Value, args []js.Value) any {
		send(handlers.GetInitCommand(xClient + yClient*common.W_CHUNKS_TOTAL))
		return nil
	}))
	js.Global().Get("window").Call("addEventListener", "beforeunload", js.FuncOf(func(this js.Value, args []js.Value) any {
		send(handlers.GetENDCommand())
		ws.Call("close")
		return nil
	}))

	ws.Set("onmessage", js.FuncOf(func(this js.Value, args []js.Value) any {
		data := args[0].Get("data")

		buf := make([]byte, data.Get("byteLength").Int())
		js.CopyBytesToGo(buf, js.Global().Get("Uint8Array").New(data))
		gChunkId := int(binary.BigEndian.Uint16(buf[0:2]))

		bufferByte.Append(gChunkId, buf[2:])
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
	getBrushType := func(brushSize, brushShape string) common.BrushType {
		switch brushShape {
		case "circle":
			if brushSize == "small" {
				return common.CIRCLE_SMALL
			}
			return common.CIRCLE_BIG
		case "square":
			if brushSize == "small" {
				return common.SQUARE_SMALL
			}
			return common.SQUARE_BIG
		}
		return common.CIRCLE_SMALL
	}
	js.Global().Set("changeBrushSize", js.FuncOf(func(this js.Value, args []js.Value) any {
		size := args[0].Get("target").Get("value").String()
		brushSize = size
		return nil
	},
	))
	js.Global().Set("changeBrushShape", js.FuncOf(func(this js.Value, args []js.Value) any {
		shape := args[0].Get("target").Get("value").String()
		brushShape = shape
		return nil
	},
	))
	js.Global().Set("goFrame", js.FuncOf(func(this js.Value, args []js.Value) any {
		x, y := mouse.Get()
		go func() {
			if x < 0 || y < 0 {
				return
			}
			if x >= common.W_CELLS_CLIENT*SIZE_CELL {
				return
			}
			if y >= common.H_CELLS_CLIENT*SIZE_CELL {
				return
			}
			x = x / SIZE_CELL
			y = y / SIZE_CELL
			if addGenerator == 1 {
				send(handlers.GetGeneratorCommand(handlers.GetDrawCommand(uint16(x), uint16(y), cellType, getBrushType(brushSize, brushShape)))...)
				addGenerator = -1
			}
			if pressed {
				send(handlers.GetDrawCommand(uint16(x), uint16(y), cellType, getBrushType(brushSize, brushShape)))
			}
		}()

		if move {
			bufferByte.Clean()
			w.ShiftWorld(xClient-oldXClient, yClient-oldYClient)
			send(handlers.GetMoveCommand(uint16(xClient + yClient*common.W_CHUNKS_TOTAL)))
			move = false
			DrawAll(w)
			oldXClient = xClient
			oldYClient = yClient
			fmt.Printf("x: %v/%v\n", xClient, common.W_CHUNKS_TOTAL-common.W_CHUNKS_CLIENT)
			fmt.Printf("y: %v/%v\n", yClient, common.H_CHUNKS_TOTAL-common.H_CHUNKS_CLIENT)
			//return nil
		}
		chunks := bufferByte.GetChunks()
		var toDraw = []int{}
		if len(chunks) != 0 {
			for _, idChunk = range chunks {
				x, y := common.GetServerXYChunk(idChunk)
				x = x - xClient
				y = y - yClient
				if x < 0 || x >= common.W_CHUNKS_CLIENT {
					continue
				}
				if y < 0 || y >= common.H_CHUNKS_CLIENT {
					continue
				}

				w.SetDecodedCells(bufferByte.GetLast(idChunk), x+y*common.W_CHUNKS_CLIENT)
				toDraw = append(toDraw, x+y*common.W_CHUNKS_CLIENT)
			}
		}
		Draw(w, toDraw)

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

func throttle(t time.Duration, fn func()) func() {
	var mu sync.Mutex
	var running bool

	return func() {
		mu.Lock()
		defer mu.Unlock()
		if running {
			return
		}
		running = true
		go func() {
			time.Sleep(t * time.Millisecond)
			mu.Lock()
			running = false
			mu.Unlock()
		}()
		fn()
	}
}

func sendRaw(bytes []byte) {
	dst := js.Global().Get("Uint8Array").New(8)
	js.CopyBytesToJS(dst, bytes)
	ws.Call("send", dst)
}
