package common

const SLEEP = 50
const FPS = 30
const SIZE_CELL = 5
const CHUNK_SIZE = 5

const H_CHUNKS_CLIENT = 24
const W_CHUNKS_CLIENT = 48

const H_CHUNKS_TOTAL = 150
const W_CHUNKS_TOTAL = 240

const W_CELLS_TOTAL = W_CHUNKS_TOTAL * CHUNK_SIZE
const H_CELLS_TOTAL = H_CHUNKS_TOTAL * CHUNK_SIZE

const W_CELLS_CLIENT = W_CHUNKS_CLIENT * CHUNK_SIZE
const H_CELLS_CLIENT = H_CHUNKS_CLIENT * CHUNK_SIZE

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
