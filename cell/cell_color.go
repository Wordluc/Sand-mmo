package cell

import (
	"fmt"
	"math/rand"
	"sand-mmo/common"
)

var colorMap map[CellType][]common.Color = map[CellType][]common.Color{
	SAND_CELL: {
		common.Yellow,
		common.Gold,
		common.NewColor(240, 220, 80, 255),
		common.NewColor(210, 180, 40, 255),
	},
	WATER_CELL: {
		common.Blue,
		common.DarkBlue,
		common.NewColor(0, 60, 120, 255),
	},
	SMOKE_CELL: {
		common.LightGray,
		common.Gray,
		common.NewColor(100, 100, 100, 180),
	},
	EMPTY_CELL: {
		common.SkyBlue,
	},
	STONE_CELL: {
		common.LightGray,
		common.Gray,
		common.DarkGray,
		common.NewColor(60, 60, 60, 255),
	},
	FIRE_CELL: {
		common.Yellow,
		common.Gold,
		common.Orange,
		common.Red,
	},
	LAVA_CELL: {
		common.Orange,
		common.Red,
		common.Maroon,
	},
	LEAF_CELL: {
		common.NewColor(80, 200, 60, 255),
		common.Green,
		common.Lime,
		common.DarkGreen,
	},
	WOOD_CELL: {
		common.Beige,
		common.Brown,
		common.DarkBrown,
	},
	VACUUM_CELL: {
		common.Purple,
		common.Violet,
		common.DarkPurple,
		common.Black,
	},
}

func (cell *Cell) GenerateNewColor() (color common.Color, spriteType uint8) {
	var idx uint8
	colors, ok := colorMap[cell.CellType]
	if !ok {
		fmt.Println("Color not found")
		return common.Black, 0
	}
	idx = uint8(rand.Intn(len(colors)))
	cell.SpirteType = idx
	return colors[idx], idx
}

func (cell *Cell) GetColor() (color common.Color) {
	return colorMap[cell.CellType][cell.SpirteType]
}
