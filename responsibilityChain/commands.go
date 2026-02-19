package responsibilityChain

import "sand-mmo/common"

func GetInitCommand() (p common.Package) {
	//I need to find a way to comunition a port to start a udp socket to convey map information
	return common.Package{
		Command: common.INIT,
	}
}
func GetChunkCommand(chunkId uint32) (p common.Package) {
	return common.Package{
		Command: common.GET,
		CommandPackage: common.CommandPackage{
			Ident: common.CHUNK,
			Arg:   chunkId,
		},
	}
}
func GetDrawCommand(chunkId uint8, x uint16, y uint16) (p common.Package) {
	return common.Package{
		Command: common.DRAW_IN,
		BrushPackage: common.BrushPackage{
			X: x,
			Y: y,
		},
	}
}
