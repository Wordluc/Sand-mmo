package sandmmo

import (
	"slices"
	"testing"
)

func TestWorld_GetChunk_FirstChunk(t *testing.T) {

	worldCells := []Cell{
		{1, 10, 1, false}, {2, 20, 2, false}, {3, 30, 3, false}, {4, 40, 4, false}, {5, 50, 5, false}, {6, 60, 6, false}, {7, 70, 7, false}, {8, 80, 8, false},
		{9, 90, 9, false}, {10, 100, 10, false}, {11, 110, 11, false}, {12, 120, 12, false}, {13, 130, 13, false}, {14, 140, 14, false}, {15, 150, 15, false}, {16, 160, 16, false},

		{17, 170, 17, false}, {18, 180, 18, false}, {19, 190, 19, false}, {20, 200, 20, false}, {21, 210, 21, false}, {22, 220, 22, false}, {23, 230, 23, false}, {24, 240, 24, false},
		{25, 250, 25, false}, {26, 260, 26, false}, {27, 270, 27, false}, {28, 280, 28, false}, {29, 290, 29, false}, {30, 300, 30, false}, {31, 310, 31, false}, {32, 320, 32, false},
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
