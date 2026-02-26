package common

const SLEEP = 50
const W_WINDOWS = 120
const H_WINDOWS = 70
const SIZE_CELL = 10
const CHUNK_SIZE = 10

func GetSizeFromBrushType(t BrushType) int {
	switch t {
	case CIRCLE_BIG, SQUARE_BIG:
		return 6
	case CIRCLE_SMALL, SQUARE_SMALL:
		return 2
	}
	return 0
}
