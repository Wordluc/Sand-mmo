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

func (cell *Cell) NewColor(isGradientChoose bool) (color common.Color, spriteType uint8) {
	var idx uint8
	if colors, ok := colorMap[cell.CellType]; ok {
		if isGradientChoose {
			idx = cell.SpirteType
		} else {
			idx = uint8(rand.Intn(len(colors)))
		}
		return colors[idx], idx
	}
	fmt.Println("Color not found ")
	return common.Black, 0
}

func (cell *Cell) GetColor() (color common.Color) {
	color, _ = cell.NewColor(true)
	return color
}
