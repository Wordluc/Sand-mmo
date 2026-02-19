package common

type command = uint8

const (
	INIT command = iota
	DRAW_IN
	GET command = 16 + iota
)

type args = uint16

const (
	CHUNK args = iota
)

type Package struct {
	Code           uint32
	Command        command
	Arg            uint8
	BrushPackage   //command = 0xxx
	CommandPackage //command = 1xxx
}

// 8bit command | 8bit Arg | 16 bit Indet
type CommandPackage struct {
	Ident uint16
}

// 8bit command | 8bit Arg | 8bit x coordinate | 8bit y coordinate
// X < 255 Y < 255
// Can use Arg to specify the chunk to draw in
type BrushPackage struct {
	X uint8
	Y uint8
}

func Decode(input uint32) Package {
	c := Package{}
	c.Code = input
	c.Command = uint8((input & 0xFF000000) >> (4 * 6))
	c.Arg = uint8((input & 0x00FF0000) >> (4 * 4))
	if c.Command < 16 {
		return decodeBrush(input, c)
	} else {
		return decodeCommand(input, c)
	}
}

func Encode(c Package) uint32 {

	output := uint32(c.Command) << (4 * 6)
	output = output | (uint32(c.Arg))<<(4*4)
	if c.Command < 16 {
		return encodeBrush(c, output)
	} else {
		return encodeCommand(c, output)
	}
}

func decodeCommand(input uint32, p Package) Package {
	p.Ident = uint16((input & 0x0000FFFF))
	return p
}

func encodeCommand(c Package, output uint32) uint32 {
	output = output | (uint32(c.Ident))
	return output
}

func decodeBrush(input uint32, p Package) Package {
	p.X = uint8((input & 0x0000FF00) >> (4 * 2))
	p.Y = uint8((input & 0x000000FF))
	return p
}

func encodeBrush(c Package, output uint32) uint32 {
	output = output | (uint32(c.X))<<(4*2)
	output = output | (uint32(c.Y))

	return output
}
