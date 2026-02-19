package sandmmo

type cellType = uint8

type Cell struct {
	Cell  cellType
	Life  uint16
	Extra uint16
}

func DecodeCell(input uint32) Cell {
	c := Cell{}
	c.Cell = cellType((input & 0xFF000000) >> (4 * 6))
	c.Life = uint16((input & 0x00FFF000) >> (4 * 3))
	c.Extra = uint16((input & 0x00000FFF))
	return c
}

func EncodeCell(c Cell) uint32 {
	var output uint32

	output = output | (uint32(c.Cell))<<(4*6)
	output = output | (uint32(c.Life))<<(4*3)
	output = output | (uint32(c.Extra))

	return output
}
