package commandengine

import "sand-mmo/common"

func GetInitCommand() (p common.Package) {
	return common.Package{
		Command: common.INIT,
	}
}
func GetChunkCommand(chunkId uint8) (p common.Package) {
	return common.Package{
		Command: common.GET,
		CommandPackage: common.CommandPackage{
			Ident: common.CHUNK,
		},
		Arg: chunkId,
	}
}
func GetDrawCommand(chunkId uint8, x uint8, y uint8) (p common.Package) {
	return common.Package{
		Command: common.DRAW_IN,
		Arg:     chunkId,
		BrushPackage: common.BrushPackage{
			X: x,
			Y: y,
		},
	}
}
