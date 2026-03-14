package common

const SLEEP = 50
const FPS = 30
const W_WINDOWS = 300
const H_WINDOWS = 200
const SIZE_CELL = 3
const CHUNK_SIZE = 5

var GTouchedId uint8 = 0

func UntouchEverything() {
	GTouchedId -= 1
}
