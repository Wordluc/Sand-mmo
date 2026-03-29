package wasm

import (
	"sand-mmo/cell"
	"sand-mmo/common"
	"sand-mmo/core"
	"syscall/js"
)

type Mouse struct {
	common.Vec2
	Pressed bool
}

type Brush struct {
	BrushShape   string
	BrushSize    string
	AddGenerator int
}

type Window struct {
	Pos    common.Vec2
	OldPos common.Vec2
	Size   common.Vec2
}

type WasmState struct {
	Mouse     Mouse
	Window    Window
	Brush     Brush
	CellType  cell.CellType
	World     *core.ClientWorld
	WebSocket js.Value
	Document  js.Value
	Ctx2D     js.Value
}

func NewState() (res WasmState) {
	res.Window.Size = common.NewVec2(common.W_CHUNKS_TOTAL-common.W_CHUNKS_CLIENT, common.H_CHUNKS_TOTAL-common.H_CHUNKS_CLIENT)
	res.Brush.BrushShape = "circle"
	res.Brush.BrushSize = "small"
	res.CellType = cell.SAND_CELL
	return res
}
func (b Brush) GetBrushType() common.BrushType {
	switch b.BrushShape {
	case "circle":
		if b.BrushSize == "small" {
			return common.CIRCLE_SMALL
		}
		return common.CIRCLE_BIG
	case "square":
		if b.BrushSize == "small" {
			return common.SQUARE_SMALL
		}
		return common.SQUARE_BIG
	}
	return common.CIRCLE_SMALL
}

func (w *Window) AddX(x int) {
	w.Pos.AddX(x)
	x, y := w.Pos.Get()
	width, _ := w.Size.Get()
	if x > width {
		w.Pos.Set(width, y)
	}
	if x < 0 {
		w.Pos.Set(0, y)
	}
}

func (w *Window) AddY(y int) {
	w.Pos.AddY(y)
	x, y := w.Pos.Get()
	_, height := w.Size.Get()
	if y > height {
		w.Pos.Set(x, height)
	}
	if y < 0 {
		w.Pos.Set(x, 0)
	}
}

func (w *Window) GetChunkId() int {
	x, y := w.Pos.Get()
	return x + y*common.W_CHUNKS_TOTAL
}
