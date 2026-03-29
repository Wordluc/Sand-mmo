package wasm

import (
	"sand-mmo/cell"
	"sand-mmo/common"
	"syscall/js"
	"wasm/utils"
)

var frameBuf []byte
var jsDst js.Value
var jsImageData js.Value
var canvasW, canvasH int
var bufferByte utils.Buffer = utils.NewBuffer()

func init() {
	size := int(common.W_CELLS_CLIENT) * int(common.H_CELLS_CLIENT) * int(SIZE_CELL) * int(SIZE_CELL) * 4
	frameBuf = make([]byte, size)
	jsDst = js.Global().Get("Uint8ClampedArray").New(size)
	canvasW = int(common.W_CELLS_CLIENT) * int(SIZE_CELL)
	canvasH = int(common.H_CELLS_CLIENT) * int(SIZE_CELL)
	jsImageData = js.Global().Get("ImageData").New(jsDst, canvasW, canvasH)
}

func Draw(state *WasmState) {
	var dx, dy, px int
	var color common.Color
	for chunkId := range state.World.GetNumberChucks() {
		state.World.ForEachCell(chunkId, func(x, y int, center *cell.Cell) error {
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

	state.Ctx2D.Call("putImageData", jsImageData, 0, 0)
}
