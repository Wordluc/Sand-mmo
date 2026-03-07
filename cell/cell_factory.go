package cell

import (
	"fmt"
	"sand-mmo/common"
)

func NewCell(celltype CellType) (res Cell, err error) {
	res.CellType = celltype
	res.touchedId = common.GTouchedId - 1
	res.forceTouched = true
	switch celltype {
	case SMOKE_CELL:
		res.initialLifeSec = 30
	case FIRE_CELL:
		res.initialLifeSec = 10
	case SAND_CELL:
	case LAVA_CELL:
	case WATER_CELL:
	case EMPTY_CELL:
	case WOOD_CELL:
	case STONE_CELL:
	default:
		return res, fmt.Errorf("CellType not found %v", celltype)
	}
	res.RemainingLife = float32(res.initialLifeSec)
	return res, nil

}
