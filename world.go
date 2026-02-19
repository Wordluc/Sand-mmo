package sandmmo

type World struct {
	cells        []Cell
	activeChunks []struct {
		id    uint16
		cells []Cell
	}
}

func NewWorld(w, h uint16) World {
	world := World{}
	world.cells = make([]Cell, w*h)
	return world
}

func (w *World) Set(x, y uint16, cell Cell) {
	indexCell := x * y
	w.cells[indexCell] = cell
}

func (w *World) Get(x, y uint16) Cell {
	indexCell := x * y
	return w.cells[indexCell]
}
