package common

const SLEEP = 50
const FPS = 30
const W_WINDOWS = 240
const H_WINDOWS = 120
const SIZE_CELL = 5
const CHUNK_SIZE = 5

var GTouchedId uint8 = 0

func UntouchEverything() {
	GTouchedId -= 1
}
