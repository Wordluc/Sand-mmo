package cell

import "testing"

func TestDecodeCell_Golden(t *testing.T) {
	tests := []struct {
		name string
		in   uint32
		want Cell
	}{
		{
			name: "mixed",
			in:   0xA123BC00,
			want: Cell{
				CellType:       CellType(0xA1),
				InitialLifeSec: 0x23B,
				Extra:          0xC00,
			},
		},
		{
			name: "all max",
			in:   0xFFFFFF00,
			want: Cell{
				CellType:       0xFF,
				InitialLifeSec: 0xFFF,
				Extra:          0xF00,
			},
		},
		{
			name: "only cell",
			in:   0xF0000000,
			want: Cell{CellType: 0xF0},
		},
		{
			name: "only life",
			in:   0x0ABC0000,
			want: Cell{CellType: 0x0A, InitialLifeSec: 0xBC0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := DecodeCell(tt.in)

			if out.CellType != tt.want.CellType {
				t.Fatalf("cell = %X want %X", out.CellType, tt.want.CellType)
			}
			if out.InitialLifeSec != tt.want.InitialLifeSec {
				t.Fatalf("life = %X want %X", out.InitialLifeSec, tt.want.InitialLifeSec)
			}
			if en := uint32(EncodeCell(out)); en != tt.in {
				t.Fatalf("encode = %x should be %X", en, tt.in)
			}
		})
	}
}
