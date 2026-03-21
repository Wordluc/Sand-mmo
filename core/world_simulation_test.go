package core

import (
	"sand-mmo/cell"
	"sand-mmo/common"
	"testing"
)

func BenchmarkWorldSimulation(b *testing.B) {
	w := NewServerWorld(240, 120, 5)
	w.ApplyBrush(common.BrushPackage{X: 10, Y: 40, BrushType: common.CIRCLE_BIG, CellType: cell.LAVA_CELL})
	w.AddGenerator(common.BrushPackage{X: 50, Y: 100, BrushType: common.CIRCLE_BIG, CellType: cell.WATER_CELL})
	for b.Loop() {
		w.Loop()
	}
}
