package wasm

import (
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
	Offset common.Vec2
	Size   common.Vec2
}

type WasmState struct {
	Mouse     Mouse
	Window    Window
	Brush     Brush
	CellType  core.CellType
	World     *core.ClientWorld
	WebSocket js.Value
	Document  js.Value
	Ctx2D     js.Value
}

func NewState() (res WasmState) {
	res.Window.Size = common.NewVec2(common.W_CHUNKS_TOTAL-common.W_CHUNKS_CLIENT, common.H_CHUNKS_TOTAL-common.H_CHUNKS_CLIENT)
	res.Brush.BrushShape = "circle"
	res.Brush.BrushSize = "small"
	res.CellType = core.SAND_CELL
	//res.Window.Pos.Set(0, common.H_CHUNKS_TOTAL-common.H_CHUNKS_CLIENT)
	res.Window.Pos.Set(0, 0)
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

func (w *Window) SetX(x int) {
	_, _y := w.Offset.Get()
	w.Offset.Set(x, _y)
}

func (w *Window) SetY(y int) {
	_x, _ := w.Offset.Get()
	w.Offset.Set(_x, y)
}

func (w *Window) GetChunkId() int {
	x, y := w.Pos.Get()
	return x + y*common.W_CHUNKS_TOTAL
}
