package core

import (
	"math/rand"
	"sand-mmo/cell"
	"sand-mmo/common"
)

// NewCellByChance generates a cell with type "typeCell" with a probability of chance/10.
// chance must be between 0 and 100.
func NewCellByChance(typeCell cell.CellType, chance int) (cell.Cell, bool) {
	if rand.Intn(100) < chance {
		c := cell.NewCell(typeCell)
		return c, true
	}
	c := cell.NewCell(cell.EMPTY_CELL)
	return c, false
}

func (w *ServerWorld) isFree(pos common.Vec2) bool {
	x, y := pos.Get()
	c := w.Get(x, y)
	return c != nil && c.IsEmpty()
}

func (w *ServerWorld) isFlammable(c *cell.Cell) bool {
	return c.CellType == cell.WOOD_CELL || c.CellType == cell.LEAF_CELL
}

func (w *ServerWorld) setEmptyCell(pos common.Vec2) error {
	w.SetVec(pos, cell.NewCell(cell.EMPTY_CELL))
	return nil
}
