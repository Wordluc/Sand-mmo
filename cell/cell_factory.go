package cell

import (
	"sand-mmo/common"
)

func NewCell(celltype CellType) (res Cell) {
	res.CellType = celltype
	res.touchedId = common.GTouchedId - 1
	res.forceTouched = true
	switch celltype {
	case SMOKE_CELL:
		res.initialLifeSec = 30
	case FIRE_CELL:
		res.initialLifeSec = 10
	case VACUUM_CELL:
		res.initialLifeSec = 10
	}
	res.RemainingLife = float32(res.initialLifeSec)
	return res

}
