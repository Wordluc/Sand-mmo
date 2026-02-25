package common

import (
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name string
		p    Package
		want uint64
	}{
		{
			name: "brush packet",
			p: Package{
				Command: 0x1,
				BrushPackage: BrushPackage{
					X:         0x123,
					Y:         0x456,
					TypeBrush: 0xAA,
					CellType:  0xBB,
					Extra:     0xCC,
				},
			},
			want: 0x0001123456AABBCC,
		},
		{
			name: "command packet",
			p: Package{
				Command: 0x0080,
				CommandPackage: CommandPackage{
					Ident: 0xBEEF,
					Arg:   0xDEADBEEF,
				},
			},
			want: 0x0080BEEFDEADBEEF,
		},
		{
			name: "brush max",
			p: Package{
				Command: 0x000F,
				BrushPackage: BrushPackage{
					X:         0xFFF,
					Y:         0xFFF,
					TypeBrush: 0xFF,
					CellType:  0xFF,
					Extra:     0xFF,
				},
			},
			want: 0x000FFFFFFFFFFFFF,
		},
		{
			name: "command max",
			p: Package{
				Command: 0xFFFF,
				CommandPackage: CommandPackage{
					Ident: 0xFFFF,
					Arg:   0xFFFFFFFF,
				},
			},
			want: 0xFFFFFFFFFFFFFFFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Encode(tt.p)
			if got != tt.want {
				t.Fatalf("got 0x%016X want 0x%016X", got, tt.want)
			}
		})
	}
}
func TestDecode(t *testing.T) {
	tests := []struct {
		name string
		in   uint64
		want Package
	}{
		{
			name: "brush packet",
			in:   0x0001123456AABBCC,
			want: Package{
				Command: 0x0001,
				BrushPackage: BrushPackage{
					X:         0x123,
					Y:         0x456,
					TypeBrush: 0xAA,
					CellType:  0xBB,
					Extra:     0xCC,
				},
			},
		},
		{
			name: "command packet",
			in:   0x0080BEEFDEADBEEF,
			want: Package{
				Command: 0x0080,
				CommandPackage: CommandPackage{
					Ident: 0xBEEF,
					Arg:   0xDEADBEEF,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Decode(tt.in)

			if got.Command != tt.want.Command ||
				got.X != tt.want.X ||
				got.Y != tt.want.Y ||
				got.TypeBrush != tt.want.TypeBrush ||
				got.CellType != tt.want.CellType ||
				got.Extra != tt.want.Extra ||
				got.Ident != tt.want.Ident ||
				got.Arg != tt.want.Arg {
				t.Fatalf("got %+v want %+v", got, tt.want)
			}
		})
	}
}
