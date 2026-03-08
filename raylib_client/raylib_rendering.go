package main

import (
	"fmt"
	"sand-mmo/cell"
	"sand-mmo/common"
	"sand-mmo/world"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func Draw(w world.ClientWorld) {
	var i, x, y uint16
	var color rl.Color
	for _, c := range w.GetCells() {
		x = i % w.W * common.SIZE_CELL
		y = i / w.W * common.SIZE_CELL
		switch c.CellType {
		case cell.SAND_CELL:
			color = rl.Yellow
		case cell.WATER_CELL:
			color = rl.Blue
		case cell.SMOKE_CELL:
			color = rl.LightGray
		case cell.EMPTY_CELL:
			color = rl.SkyBlue
		case cell.STONE_CELL:
			color = rl.Gray
		case cell.FIRE_CELL:
			color = rl.Orange
		case cell.LAVA_CELL:
			color = rl.Red
		case cell.LEAF_CELL:
			color = rl.Green
		case cell.WOOD_CELL:
			color = rl.Brown
		case cell.VACUUM_CELL:
			color = rl.DarkPurple
		}
		rl.DrawRectangle(int32(x), int32(y), common.SIZE_CELL, common.SIZE_CELL, color)
		rl.DrawText(fmt.Sprint(y/common.SIZE_CELL), 0, int32(y), common.SIZE_CELL, rl.Black)
		i++
	}
}
