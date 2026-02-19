package common

import (
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name string
		p    Package
		want uint32
	}{
		{
			name: "brush packet",
			p: Package{
				Command: 0x00,
				Arg:     0x34,
				BrushPackage: BrushPackage{
					X: 0x56,
					Y: 0x78,
				},
			},
			want: 0x00345678,
		},
		{
			name: "command packet",
			p: Package{
				Command: 0x80,
				Arg:     0xAA,
				CommandPackage: CommandPackage{
					Ident: 0xBEEF,
				},
			},
			want: 0x80AABEEF,
		},
		{
			name: "brush max",
			p: Package{
				Command: 0x01,
				Arg:     0xFF,
				BrushPackage: BrushPackage{
					X: 0xFF,
					Y: 0xFF,
				},
			},
			want: 0x01FFFFFF,
		},
		{
			name: "command max",
			p: Package{
				Command: 0xFF,
				Arg:     0xFF,
				CommandPackage: CommandPackage{
					Ident: 0xFFFF,
				},
			},
			want: 0xFFFFFFFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Encode(tt.p)
			if got != tt.want {
				t.Fatalf("got 0x%08X want 0x%08X", got, tt.want)
			}
		})
	}
}
func TestDecode(t *testing.T) {
	tests := []struct {
		name string
		in   uint32
		want Package
	}{
		{
			name: "brush packet",
			in:   0x00345678,
			want: Package{
				Command: 0x00,
				Arg:     0x34,
				BrushPackage: BrushPackage{
					X: 0x56,
					Y: 0x78,
				},
			},
		},
		{
			name: "command packet",
			in:   0x80AABEEF,
			want: Package{
				Command: 0x80,
				Arg:     0xAA,
				CommandPackage: CommandPackage{
					Ident: 0xBEEF,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Decode(tt.in)

			if got.Command != tt.want.Command ||
				got.Arg != tt.want.Arg ||
				got.X != tt.want.X ||
				got.Y != tt.want.Y ||
				got.Ident != tt.want.Ident {
				t.Fatalf("got %+v want %+v", got, tt.want)
			}
		})
	}
}
