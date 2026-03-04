package cell

import (
	"fmt"
	"sand-mmo/common"
)

func NewCell(celltype CellType) (res Cell, err error) {
	res.CellType = celltype
	res.touchedId = common.GTouchedId - 1
	res.forceTouched = true
	res.Velocity = new(common.NewVec2(0, 0))
	switch celltype {
	case SAND_CELL:
		res.initialLifeSec = 30
	case SMOKE_CELL:
		res.initialLifeSec = 30
	case FIRE_CELL:
		res.initialLifeSec = 10
	case WATER_CELL:
		res.initialLifeSec = 10
	case EMPTY_CELL:
	case WOOD_CELL:
	case STONE_CELL:
	default:
		return res, fmt.Errorf("CellType not found %v", celltype)
	}
	res.RemainingLife = float32(res.initialLifeSec)
	return res, nil

}
