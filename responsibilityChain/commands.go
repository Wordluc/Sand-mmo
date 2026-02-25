package responsibilityChain

import (
	sandmmo "sand-mmo"
	"sand-mmo/common"
)

func GetInitCommand(port uint32) (p common.Package) {
	return common.Package{
		Command: common.INIT,
		CommandPackage: common.CommandPackage{
			Arg: port,
		},
	}
}
func GetENDCommand() (p common.Package) {
	return common.Package{
		Command: common.END,
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
func GetDrawCommand(chunkId uint8, x uint16, y uint16, cellType sandmmo.CellType) (p common.Package) {
	return common.Package{
		Command: common.DRAW_IN,
		BrushPackage: common.BrushPackage{
			X:        x,
			Y:        y,
			CellType: cellType,
		},
	}
}
