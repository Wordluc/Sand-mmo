package core

import (
	"encoding/binary"
	"slices"
	"testing"
)

func newCell_for_test(a, b int) Cell {
	return NewCell(0)

}
func TestWorld_GetChunk_FirstChunk(t *testing.T) {

	worldCells := []Cell{
		newCell_for_test(1, 10), newCell_for_test(2, 20), newCell_for_test(3, 30), newCell_for_test(4, 40), newCell_for_test(5, 50), newCell_for_test(6, 60), newCell_for_test(7, 70), newCell_for_test(8, 80),
		newCell_for_test(9, 90), newCell_for_test(10, 100), newCell_for_test(11, 110), newCell_for_test(12, 120), newCell_for_test(13, 130), newCell_for_test(14, 140), newCell_for_test(15, 150), newCell_for_test(16, 160),

		newCell_for_test(17, 170), newCell_for_test(18, 180), newCell_for_test(19, 190), newCell_for_test(20, 200), newCell_for_test(21, 210), newCell_for_test(22, 220), newCell_for_test(23, 230), newCell_for_test(24, 240),
		newCell_for_test(25, 250), newCell_for_test(26, 260), newCell_for_test(27, 270), newCell_for_test(28, 280), newCell_for_test(29, 290), newCell_for_test(30, 300), newCell_for_test(31, 310), newCell_for_test(32, 320),
	}

	var encoded []byte
	for _, c := range worldCells {
		encoded = binary.BigEndian.AppendUint16(encoded, EncodeCell(c))
	}

	// 8x4 world
	w := newServerWorld_test(8, 4, 2)
	w.ImportCells(encoded)

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
		var want []uint16
		for i := range c.cell {

			want = append(want, EncodeCell(c.cell[i]))
		}
		got := w.GetChunkBytes(c.idChunk)

		if !slices.Equal(got, want) {
			t.Fatalf("GetChunk(%v) failed\n got: %v\nwant: %v", c.idChunk, got, want)
		}

		for i, v := range got {
			if DecodeCell(v) != c.cell[i] {
				t.Fatalf("decode mismatch %+v != %+v", DecodeCell(v), c.cell[i])
			}
		}
	}

}
