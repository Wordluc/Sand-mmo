package cell

import (
	"math/rand"
	"sand-mmo/common"
)

type CellType = uint8

const (
	EMPTY_CELL CellType = iota
	SAND_CELL
	WATER_CELL
	SMOKE_CELL
	STONE_CELL
	WOOD_CELL
	FIRE_CELL
	LAVA_CELL
	LEAF_CELL
	VACUUM_CELL
)

type Cell struct {
	CellType CellType
	Extra    uint8

	initialLifeSec uint16
	RemainingLife  float32
	touchedId      uint8
	forceTouched   bool
}

func (c *Cell) DecreaseLife() bool {
	c.RemainingLife -= float32(rand.Intn(3))
	return c.RemainingLife <= 0
}

func (c Cell) IsEmpty() bool {
	return c.CellType == EMPTY_CELL
}

func (c *Cell) IsTouched() bool {
	return c.touchedId == common.GTouchedId
}

func (c *Cell) Touched() {
	c.touchedId = common.GTouchedId
}

func (c *Cell) IsNew() bool {
	defer func() {
		c.forceTouched = false
	}()
	return c.forceTouched
}

func DecodeCell(input uint16) Cell {
	c := Cell{}
	c.CellType = CellType((input & 0xFF00) >> (4 * 2))
	c.Extra = uint8((input & 0x00FF))
	c.touchedId = common.GTouchedId
	return c
}

func EncodeCell(c Cell) uint16 {
	var output uint16

	output = output | (uint16(c.CellType))<<(4*2)
	output = output | (uint16(c.Extra))

	return output
}
