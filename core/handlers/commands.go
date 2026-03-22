package handlers

import (
	"sand-mmo/cell"
	"sand-mmo/common"
)

func GetInitCommand(chunkId int) (p common.Package) {
	return common.Package{
		Command: common.INIT,
		CommandPackage: common.CommandPackage{
			Arg1: uint16(chunkId),
		},
	}
}
func GetENDCommand() (p common.Package) {
	return common.Package{
		Command: common.END,
	}
}

func GetGeneratorCommand(brush common.Package) (res []common.Package) {
	res = append(res, common.Package{
		Command: common.ADD_GENERATOR,
	})
	res = append(res, brush)
	return res
}

func GetMoveCommand(atChunkId uint16) (p common.Package) {
	return common.Package{
		Command: common.MOVE_AT,
		CommandPackage: common.CommandPackage{
			Arg1: atChunkId,
		},
	}
}

func GetChunkCommand(chunkId uint16) (p common.Package) {
	return common.Package{
		Command: common.GET,
		CommandPackage: common.CommandPackage{
			Arg1: chunkId,
		},
	}
}
func GetDrawCommand(x uint16, y uint16, cellType cell.CellType, brushType common.BrushType) (p common.Package) {
	return common.Package{
		Command: common.DRAW_IN,
		BrushPackage: common.BrushPackage{
			X:         x,
			Y:         y,
			CellType:  cellType,
			BrushType: brushType,
		},
	}
}
