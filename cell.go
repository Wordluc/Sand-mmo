package sandmmo

type cellType uint8

type Cell struct {
	Cell   cellType
	Life   uint16
	SpeedX uint8
	SpeedY uint8
}

func DecodeCell(input uint32) Cell {
	c := Cell{}
	c.Cell = cellType((input & 0xF0000000) >> (4 * 7))
	c.Life = uint16((input & 0x0FFF0000) >> (4 * 4))
	c.SpeedX = uint8((input & 0x0000FF00) >> (4 * 2))
	c.SpeedY = uint8((input & 0x000000FF))
	return c
}

func EncodeCell(c Cell) uint32 {
	var output uint32

	output = output | (uint32(c.Cell))<<(4*7)
	output = output | (uint32(c.Life))<<(4*4)
	output = output | (uint32(c.SpeedX))<<(4*2)
	output = output | (uint32(c.SpeedY))

	return output
}
