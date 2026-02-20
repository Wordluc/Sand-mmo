package sandmmo

import (
	"slices"
	"testing"
)

func TestWorld_GetChunk_FirstChunk(t *testing.T) {

	worldCells := []Cell{
		{1, 10, 1}, {2, 20, 2}, {3, 30, 3}, {4, 40, 4}, {5, 50, 5}, {6, 60, 6}, {7, 70, 7}, {8, 80, 8},
		{9, 90, 9}, {10, 100, 10}, {11, 110, 11}, {12, 120, 12}, {13, 130, 13}, {14, 140, 14}, {15, 150, 15}, {16, 160, 16},

		{17, 170, 17}, {18, 180, 18}, {19, 190, 19}, {20, 200, 20}, {21, 210, 21}, {22, 220, 22}, {23, 230, 23}, {24, 240, 24},
		{25, 250, 25}, {26, 260, 26}, {27, 270, 27}, {28, 280, 28}, {29, 290, 29}, {30, 300, 30}, {31, 310, 31}, {32, 320, 32},
	}

	var encoded []uint32
	for _, c := range worldCells {
		encoded = append(encoded, EncodeCell(c))
	}

	// 8x4 world
	w := NewWorld(8, 4)
	w.ImportCell(encoded)

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
