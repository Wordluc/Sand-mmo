package world

import (
	"math/rand"
	"sand-mmo/cell"
)

// NewCellByChance generates a cell with type "typeCell" with a probability of chance/10.
// chance must be between 0 and 10.
func NewCellByChance(typeCell cell.CellType, chance int) (cell.Cell, bool) {
	if rand.Intn(10) < chance {
		c, _ := cell.NewCell(typeCell)
		return c, true
	}
	c, _ := cell.NewCell(cell.EMPTY_CELL)
	return c, false
}
