package cell

import "testing"

func TestDecodeCell_Golden(t *testing.T) {
	tests := []struct {
		name string
		in   uint16
		want Cell
	}{
		{
			name: "mixed",
			in:   0xA123,
			want: Cell{
				CellType:   CellType(0xA1),
				SpirteType: 0x23,
			},
		},
		{
			name: "all max",
			in:   0xFFFF,
			want: Cell{
				CellType:   0xFF,
				SpirteType: 0xFF,
			},
		},
		{
			name: "only cell",
			in:   0xF000,
			want: Cell{CellType: 0xF0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := DecodeCell(tt.in)

			if out.CellType != tt.want.CellType {
				t.Fatalf("cell = %X want %X", out.CellType, tt.want.CellType)
			}
			if out.initialLifeSec != tt.want.initialLifeSec {
				t.Fatalf("life = %X want %X", out.initialLifeSec, tt.want.initialLifeSec)
			}
			if en := EncodeCell(out); en != tt.in {
				t.Fatalf("encode = %x should be %X", en, tt.in)
			}
		})
	}
}
