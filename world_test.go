package sandmmo

import (
	"slices"
	"testing"
)

func TestWorld_GetChunk_FirstChunk(t *testing.T) {

	worldCells := []Cell{
		NewCell(1, 10), NewCell(2, 20), NewCell(3, 30), NewCell(4, 40), NewCell(5, 50), NewCell(6, 60), NewCell(7, 70), NewCell(8, 80),
		NewCell(9, 90), NewCell(10, 100), NewCell(11, 110), NewCell(12, 120), NewCell(13, 130), NewCell(14, 140), NewCell(15, 150), NewCell(16, 160),

		NewCell(17, 170), NewCell(18, 180), NewCell(19, 190), NewCell(20, 200), NewCell(21, 210), NewCell(22, 220), NewCell(23, 230), NewCell(24, 240),
		NewCell(25, 250), NewCell(26, 260), NewCell(27, 270), NewCell(28, 280), NewCell(29, 290), NewCell(30, 300), NewCell(31, 310), NewCell(32, 320),
	}

	var encoded []uint32
	for _, c := range worldCells {
		encoded = append(encoded, EncodeCell(c))
	}

	// 8x4 world
	w := NewWorld(8, 4, 2)
	w.importCell(encoded)

	caseTest := []struct {
		idChunk int
		cell    []Cell
	}{
		{
			idChunk: 1,
			cell: []Cell{
				worldCells[2], worldCells[3],
				worldCells[10], worldCells[11],
			},
		},
		{
			idChunk: 3,
			cell: []Cell{
				worldCells[6], worldCells[7],
				worldCells[14], worldCells[15],
			},
		},
		{
			idChunk: 5,
			cell: []Cell{
				worldCells[18], worldCells[19],
				worldCells[26], worldCells[27],
			},
		},
		{
			idChunk: 6,
			cell: []Cell{
				worldCells[20], worldCells[21],
				worldCells[28], worldCells[29],
			},
		},
	}

	for _, c := range caseTest {
		var want []uint32
		for i := range c.cell {

			want = append(want, EncodeCell(c.cell[i]))
		}
		got := w.GetChunk(uint16(c.idChunk))

		if !slices.Equal(got, want) {
			t.Fatalf("GetChunk(%v) failed\n got: %v\nwant: %v", c.idChunk, got, want)
		}

		for i, v := range got {
			if DecodeCell(v) != c.cell[i] {
				t.Fatalf("decode mismatch at %d", i)
			}
		}
	}

}
