package common

type Command = uint16

// <16 brush
// <=16 commands
const (
	DRAW_IN         = iota
	GET     Command = 16 + iota
	INIT
	INITGOD
	ADD_GENERATOR
	MOVE_AT
	END
)

type args = uint16

type Package struct {
	Code    uint64
	Command uint16
	BrushPackage
	CommandPackage
}

// 16bit command | 16bit ident | 32bit args
type CommandPackage struct {
	Arg1 uint16
	Arg2 uint32
}
type BrushType = uint8

const (
	CIRCLE_SMALL BrushType = iota
	CIRCLE_BIG
	SQUARE_SMALL
	SQUARE_BIG
)

// 16bit command | 12b x | 12b y | 8b typeBrush | 8b typeMaterial | 8b extra
type BrushPackage struct {
	X         uint16
	Y         uint16
	BrushType BrushType
	CellType  uint8
	Extra     uint8
}

func Decode(input uint64) Package {
	c := Package{}
	c.Code = input
	c.Command = uint16((input & 0xFFFF000000000000) >> (4 * 12))
	if c.Command < 16 {
		return decodeBrush(input, c)
	} else {
		return decodeCommand(input, c)
	}
}

func Encode(c Package) uint64 {
	output := uint64(c.Command) << 48
	if c.Command < 16 {
		return encodeBrush(c, output)
	} else {
		return encodeCommand(c, output)
	}
}

func decodeCommand(input uint64, p Package) Package {
	p.Arg1 = uint16((input & 0x0000FFFF00000000) >> 32)
	p.Arg2 = uint32((input & 0x00000000FFFFFFFF))
	return p
}

func encodeCommand(c Package, output uint64) uint64 {
	output |= (uint64(c.Arg1) & 0xFFFF) << 32
	output |= (uint64(c.Arg2) & 0xFFFFFFFF)
	return output
}

func decodeBrush(input uint64, p Package) Package {
	p.X = uint16((input & 0x0000FFF000000000) >> 36)
	p.Y = uint16((input & 0x0000000FFF000000) >> 24)
	p.BrushType = uint8((input & 0x0000000000FF0000) >> 16)
	p.CellType = uint8((input & 0x000000000000FF00) >> 8)
	p.Extra = uint8((input & 0x00000000000000FF) >> 0)
	return p
}

func encodeBrush(c Package, output uint64) uint64 {
	output |= (uint64(c.X) & 0x0FFF) << 36
	output |= (uint64(c.Y) & 0x0FFF) << 24
	output |= (uint64(c.BrushType) & 0x00FF) << 16
	output |= (uint64(c.CellType) & 0x00FF) << 8
	output |= (uint64(c.Extra) & 0x00FF) << 0
	return output
}
