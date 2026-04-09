package core

import (
	"math/rand"
	"sand-mmo/common"
)

// NewCellByChance generates a cell with type "typeCell" with a probability of chance/10.
// chance must be between 0 and 100.
func NewCellByChance(typeCell CellType, chance int) (Cell, bool) {
	if rand.Intn(100) < chance {
		c := NewCell(typeCell)
		return c, true
	}
	c := NewCell(EMPTY_CELL)
	return c, false
}

func (w *ServerWorld) isFree(pos common.Vec2) bool {
	x, y := pos.Get()
	c := w.Get(x, y)
	return c != nil && c.IsEmpty()
}

func (w *ServerWorld) isFlammable(c *Cell) bool {
	return c.CellType == WOOD_CELL || c.CellType == LEAF_CELL
}

func (w *ServerWorld) isBurning(c *Cell) bool {
	return c.CellType == FIRE_CELL || c.CellType == LAVA_CELL
}

func (w *ServerWorld) setEmptyCell(pos common.Vec2) error {
	w.SetVec(pos, NewCell(EMPTY_CELL))
	return nil
}
