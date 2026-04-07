package common

const SLEEP = 50
const FPS = 30
const SIZE_CELL = 5
const CHUNK_SIZE = 5

const W_CELLS_TOTAL = 1200
const H_CELLS_TOTAL = 750

const W_CELLS_CLIENT = 240
const H_CELLS_CLIENT = 120

const H_CHUNKS_CLIENT = H_CELLS_CLIENT / CHUNK_SIZE
const W_CHUNKS_CLIENT = W_CELLS_CLIENT / CHUNK_SIZE

const H_CHUNKS_TOTAL = H_CELLS_TOTAL / CHUNK_SIZE
const W_CHUNKS_TOTAL = W_CELLS_TOTAL / CHUNK_SIZE

var GTouchedId uint8 = 0

func UntouchEverything() {
	GTouchedId -= 1
}

//const W_WINDOWS = 240
//const H_WINDOWS = 120

func GetServerXYChunk(idChunk int) (x, y int) {
	y = idChunk / W_CHUNKS_TOTAL
	x = idChunk % W_CHUNKS_TOTAL
	return x, y
}
