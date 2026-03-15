package cell

import (
	"math/rand"
	"sand-mmo/common"
)

func (cell *Cell) NewColor(isGradientChoose bool) (color common.Color, spriteType uint8) {
	// Helper: pick a random index from the palette
	// Use cell position or a stored field for determinism if needed
	type palette = []common.Color

	pick := func(p palette) (common.Color, uint8) {
		var idx uint8
		if isGradientChoose {
			idx = cell.SpirteType
		} else {
			idx = uint8(rand.Intn(len(p)))
		}
		return p[idx], idx
	}

	switch cell.CellType {

	case SAND_CELL:
		p := palette{
			common.Yellow,                      // 0
			common.Gold,                        // 1
			common.NewColor(240, 220, 80, 255), // 2 - pale yellow
			common.NewColor(210, 180, 40, 255), // 3 - dark sand
		}
		return pick(p)

	case WATER_CELL:
		p := palette{
			common.Blue,                      // 1 - blue
			common.DarkBlue,                  // 2 - dark blue
			common.NewColor(0, 60, 120, 255), // 3 - deep water
		}
		return pick(p)

	case SMOKE_CELL:
		p := palette{
			common.LightGray,                    // 1
			common.Gray,                         // 2
			common.NewColor(100, 100, 100, 180), // 3 - translucent dark
		}
		return pick(p)

	case EMPTY_CELL:
		p := palette{
			common.SkyBlue, // 0
		}
		return pick(p)

	case STONE_CELL:
		p := palette{
			common.LightGray,                 // 0
			common.Gray,                      // 1
			common.DarkGray,                  // 2
			common.NewColor(60, 60, 60, 255), // 3 - near black stone
		}
		return pick(p)

	case FIRE_CELL:
		p := palette{
			common.Yellow, // 0 - hot core
			common.Gold,   // 1
			common.Orange, // 2
			common.Red,    // 3 - outer flame
		}
		return pick(p)

	case LAVA_CELL:
		p := palette{
			common.Orange, // 0 - bright lava
			common.Red,    // 1
			common.Maroon, // 2
		}
		return pick(p)

	case LEAF_CELL:
		p := palette{
			common.NewColor(80, 200, 60, 255), // 0 - bright leaf
			common.Green,                      // 1
			common.Lime,                       // 2
			common.DarkGreen,                  // 3
		}
		return pick(p)

	case WOOD_CELL:
		p := palette{
			common.Beige,     // 0 - light wood
			common.Brown,     // 1
			common.DarkBrown, // 2
		}
		return pick(p)

	case VACUUM_CELL:
		p := palette{
			common.Purple,     // 0
			common.Violet,     // 1
			common.DarkPurple, // 2
			common.Black,      // 3
		}
		return pick(p)

	default:
		return common.Magenta, 0
	}
}

func (cell *Cell) GetColor() (color common.Color) {
	color, _ = cell.NewColor(true)
	return color
}
