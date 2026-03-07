package world

import (
	"math/rand"
	"sand-mmo/cell"
)

// NewCellByChance generates a cell with type "typeCell" with a probability of chance/10.
// chance must be between 0 and 10.
func NewCellByChance(typeCell cell.CellType, chance int) (cell.Cell, error) {
	if rand.Intn(10) < chance {
		return cell.NewCell(typeCell)
	}
	return cell.NewCell(cell.EMPTY_CELL)
}
