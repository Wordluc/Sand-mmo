package sandmmo

import "testing"

func TestDecodeCell_Golden(t *testing.T) {
	tests := []struct {
		name string
		in   uint32
		want Cell
	}{
		{
			name: "mixed",
			in:   0xA123BCDE,
			want: Cell{
				Cell:   cellType(0xA),
				Life:   0x123,
				SpeedX: 0xBC,
				SpeedY: 0xDE,
			},
		},
		{
			name: "all max",
			in:   0xFFFFFFFF,
			want: Cell{
				Cell:   0xF,
				Life:   0xFFF,
				SpeedX: 0xFF,
				SpeedY: 0xFF,
			},
		},
		{
			name: "only cell",
			in:   0xF0000000,
			want: Cell{Cell: 0xF},
		},
		{
			name: "only life",
			in:   0x0ABC0000,
			want: Cell{Life: 0xABC},
		},
		{
			name: "only speedX",
			in:   0x0000AB00,
			want: Cell{SpeedX: 0xAB},
		},
		{
			name: "only speedY",
			in:   0x000000CD,
			want: Cell{SpeedY: 0xCD},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := DecodeCell(tt.in)

			if out.Cell != tt.want.Cell {
				t.Fatalf("cell = %X want %X", out.Cell, tt.want.Cell)
			}
			if out.Life != tt.want.Life {
				t.Fatalf("life = %X want %X", out.Life, tt.want.Life)
			}
			if out.SpeedX != tt.want.SpeedX {
				t.Fatalf("speedX = %X want %X", out.SpeedX, tt.want.SpeedX)
			}
			if out.SpeedY != tt.want.SpeedY {
				t.Fatalf("speedY = %X want %X", out.SpeedY, tt.want.SpeedY)
			}
			if en := uint32(EncodeCell(out)); en != tt.in {
				t.Fatalf("encode = %x should be %X", en, tt.in)
			}
		})
	}
}
