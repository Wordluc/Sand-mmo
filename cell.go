package sandmmo

type CellType = uint8

const (
	NULL_CELL CellType = iota
	SAND_CELL
	WATER_CELL
	SMOKE_CELL
	DELETE_CELL
	STONE_CELL
	WOOD_CELL
	FIRE_CELL
)

type Cell struct {
	CellType       CellType
	InitialLifeSec uint16
	RemainingLife  float32
	Extra          uint16
	touchedId      uint8
	forceTouched   bool
}

func NewCell(cellType CellType, initialLife uint16) (res Cell) {
	res.CellType = cellType
	res.InitialLifeSec = initialLife
	res.RemainingLife = float32(initialLife)
	res.touchedId = GTouchedId - 1
	res.forceTouched = true
	return res

}

func (c *Cell) DecreaseLife() bool {
	c.RemainingLife -= 1
	return c.RemainingLife <= 0
}

func (c Cell) IsEmpty() bool {
	return c.CellType == NULL_CELL
}

func (c *Cell) IsTouched() bool {
	return c.touchedId == GTouchedId
}

func (c *Cell) Touched() {
	c.touchedId = GTouchedId
}

func DecodeCell(input uint32) Cell {
	c := Cell{}
	c.CellType = CellType((input & 0xFF000000) >> (4 * 6))
	c.InitialLifeSec = uint16((input & 0x00FFF000) >> (4 * 3))
	c.Extra = uint16((input & 0x00000FFF))
	c.touchedId = GTouchedId
	return c
}

func EncodeCell(c Cell) uint32 {
	var output uint32

	output = output | (uint32(c.CellType))<<(4*6)
	output = output | (uint32(c.InitialLifeSec))<<(4*3)
	output = output | (uint32(c.Extra))

	return output
}
