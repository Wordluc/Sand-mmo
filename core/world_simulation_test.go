package core

import (
	"sand-mmo/common"
	"testing"
)

func BenchmarkWorldSimulation(b *testing.B) {
	w := newServerWorld_test(240, 120, 5)
	w.ApplyBrush(common.BrushPackage{X: 10, Y: 40, BrushType: common.CIRCLE_BIG, CellType: LAVA_CELL})
	w.AddGenerator(common.BrushPackage{X: 50, Y: 100, BrushType: common.CIRCLE_BIG, CellType: WATER_CELL})
	for b.Loop() {
		w.Loop()
	}
}
