package world

import (
	"sand-mmo/cell"
	"sand-mmo/common"
)

type ClientWorld struct {
	world
}

func NewClientWorld(w, h, chunkSize uint16) (res ClientWorld) {
	res.world = newWorld(w, h, chunkSize)
	return res
}

func (w *ClientWorld) GetCells() []cell.Cell {
	return w.cells
}

func (w *ClientWorld) GetColor(cellType cell.CellType) (color common.Color) {
	switch cellType {
	case cell.SAND_CELL:
		color = common.Yellow
	case cell.WATER_CELL:
		color = common.Blue
	case cell.SMOKE_CELL:
		color = common.LightGray
	case cell.EMPTY_CELL:
		color = common.SkyBlue
	case cell.STONE_CELL:
		color = common.Gray
	case cell.FIRE_CELL:
		color = common.Orange
	case cell.LAVA_CELL:
		color = common.Red
	case cell.LEAF_CELL:
		color = common.Green
	case cell.WOOD_CELL:
		color = common.Brown
	case cell.VACUUM_CELL:
		color = common.DarkPurple
	}
	return color
}
