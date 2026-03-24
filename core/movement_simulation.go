package core

import (
	"sand-mmo/cell"
	"sand-mmo/common"
)

var (
	sand_movement = []common.Vec2{
		common.NewVec2(0, 1),
		common.NewVec2(1, 1),
		common.NewVec2(-1, 1),
	}
	lave_movement = []common.Vec2{
		common.NewVec2(0, 1),
		common.NewVec2(1, 1),
		common.NewVec2(-1, 1),
		common.NewVec2(-1, 0),
		common.NewVec2(1, 0),
	}
	leaf_movement = []common.Vec2{
		common.NewVec2(0, 1),
		common.NewVec2(1, 1),
		common.NewVec2(-1, 1),
	}
	water_movement = []common.Vec2{
		common.NewVec2(0, 1),
		common.NewVec2(1, 1),
		common.NewVec2(-1, 1),
		common.NewVec2(-1, 0),
		common.NewVec2(1, 0),
	}
	vacumm_movement = []common.Vec2{
		common.NewVec2(0, 1),
		common.NewVec2(0, -1),
		common.NewVec2(1, 0),
		common.NewVec2(-1, 0),
	}
	smoke_movement = []common.Vec2{
		common.NewVec2(0, -1),
		common.NewVec2(1, -1),
		common.NewVec2(-1, -1),
		common.NewVec2(-1, 0),
		common.NewVec2(1, 0),
	}
	fire_movement = []common.Vec2{
		common.NewVec2(0, 1),
		common.NewVec2(0, -1),
		common.NewVec2(1, 0),
		common.NewVec2(-1, 0),
	}
)

func (w *ServerWorld) SimulateChunk(idChunk int) error {
	var pos common.Vec2
	return w.ForEachCell(idChunk, func(x, y int, center *cell.Cell) error {
		if center == nil {
			return nil
		}
		if center.IsNew() {
			w.activeChunks.SortedInsert(idChunk)
		}
		if center.IsEmpty() || center.IsTouched() {
			return nil
		}

		pos = common.NewVec2(x, y)

		switch center.CellType {

		case cell.SAND_CELL:
			w.simulateSimpleMovements(pos, 2, &center, sand_movement)

		case cell.LAVA_CELL:
			w.simulateLavaMovements(idChunk, pos, 1, &center, lave_movement)

		case cell.LEAF_CELL:
			w.simulateLeafMovements(pos, 1, &center, leaf_movement)

		case cell.WATER_CELL:
			w.simulateWaterMovements(pos, 2, &center, water_movement)

		case cell.VACUUM_CELL:
			if center.RemainingLife <= 0 {
				w.SetVec(pos, cell.NewCell(cell.EMPTY_CELL))
				return nil
			}
			center.DecreaseLife()
			if !w.simulateVacuumMovements(pos, 1, &center, vacumm_movement) {
				center.Touched()
				w.activeChunks.SortedInsert(idChunk)
			}

		case cell.SMOKE_CELL:
			if center.RemainingLife <= 0 {
				w.SetVec(pos, cell.NewCell(cell.EMPTY_CELL))
				return nil
			}
			center.DecreaseLife()
			if !w.simulateSimpleMovements(pos, 2, &center, smoke_movement) {
				center.Touched()
				w.activeChunks.SortedInsert(idChunk)
			}

		case cell.FIRE_CELL:
			moved := w.simulateFireMovements(idChunk, pos, 2, &center, fire_movement)
			if center.RemainingLife <= 0 {
				smoke, _ := NewCellByChance(cell.SMOKE_CELL, 20)
				w.SetVec(pos, smoke)
				return nil
			}
			if !moved {
				center.Touched()
				w.activeChunks.SortedInsert(idChunk)
				center.DecreaseLife()
			}
		}

		return nil
	})
}

func (w *ServerWorld) isFree(pos common.Vec2) bool {
	x, y := pos.Get()
	c := w.Get(x, y)
	return c != nil && c.IsEmpty()
}

func (w *ServerWorld) setEmptyCell(pos common.Vec2) error {
	w.SetVec(pos, cell.NewCell(cell.EMPTY_CELL))
	return nil
}

func (w *ServerWorld) simulateCustomMovements(
	pos common.Vec2,
	maxSpeed int,
	c **cell.Cell,
	callbackInNewPosition func(common.Vec2) bool,
	callbackForOldPosition func(common.Vec2) error,
	groups []common.Vec2,
) bool {
	var (
		formal_group       common.Vec2
		currentSpeed       int
		performed_position common.Vec2
	)
	oldPos := pos.Copy()

	for _, formal_group = range groups {
		for currentSpeed = maxSpeed; currentSpeed > 0; currentSpeed-- {
			performed_position = formal_group.Copy()
			performed_position.MultConst(currentSpeed)
			performed_position.Add(pos)
			if callbackInNewPosition(performed_position) {
				(*c).Touched()
				w.SetVec(performed_position, *(*c))
				*c = w.GetVec(performed_position)
				callbackForOldPosition(oldPos)
				return true
			}
		}
	}
	return false
}

func (w *ServerWorld) simulateSimpleMovements(
	pos common.Vec2,
	maxSpeed int,
	c **cell.Cell,
	groups []common.Vec2,
) bool {
	return w.simulateCustomMovements(pos, maxSpeed, c, w.isFree, w.setEmptyCell, groups)
}

func (w *ServerWorld) simulateWaterMovements(
	pos common.Vec2,
	maxSpeed int,
	c **cell.Cell,
	groups []common.Vec2,
) bool {
	put_out := func(posToCheck common.Vec2) bool {
		x, y := posToCheck.Get()
		tcell := w.Get(x, y)
		if tcell == nil {
			return false
		}
		if tcell.CellType == cell.LAVA_CELL || tcell.CellType == cell.FIRE_CELL {
			smoke, isSmoke := NewCellByChance(cell.SMOKE_CELL, 10)
			if !isSmoke {
				tcell.RemainingLife = 0
				return false
			}
			w.SetVec(pos, smoke)
			return false
		}
		return w.isFree(posToCheck)
	}
	return w.simulateCustomMovements(pos, maxSpeed, c, put_out, w.setEmptyCell, groups)
}

func (w *ServerWorld) simulateLeafMovements(
	pos common.Vec2,
	maxSpeed int,
	c **cell.Cell,
	groups []common.Vec2,
) bool {
	move_light_leaf := func(posToCheck common.Vec2) bool {
		tcell := w.GetVec(posToCheck)
		if tcell == nil {
			return false
		}
		if tcell.CellType == cell.LAVA_CELL || tcell.CellType == cell.FIRE_CELL {
			w.SetVec(pos, cell.NewCell(cell.FIRE_CELL))
			return false
		}
		return w.isFree(posToCheck)
	}
	return w.simulateCustomMovements(pos, maxSpeed, c, move_light_leaf, w.setEmptyCell, groups)
}

func (w *ServerWorld) simulateVacuumMovements(
	pos common.Vec2,
	maxSpeed int,
	c **cell.Cell,
	groups []common.Vec2,
) bool {
	nothing := func(_ common.Vec2) error { return nil }
	delete_cell := func(posToCheck common.Vec2) bool {
		tcell := w.GetVec(posToCheck)
		if tcell == nil {
			return false
		}
		if !w.isFree(posToCheck) && tcell.CellType != cell.VACUUM_CELL {
			w.SetVec(posToCheck, cell.NewCell(cell.EMPTY_CELL))
		}
		return false
	}
	return w.simulateCustomMovements(pos, maxSpeed, c, delete_cell, nothing, groups)
}

func (w *ServerWorld) simulateFireMovements(
	idChunk int,
	pos common.Vec2,
	maxSpeed int,
	c **cell.Cell,
	groups []common.Vec2,
) bool {
	change_color_cell_fire := func(pos common.Vec2) error {
		prev := w.GetVec(pos)
		prev.GenerateNewColor()
		prev.Touched()
		w.activeChunks.SortedInsert(idChunk)
		prev.DecreaseLife()
		return nil
	}
	light_fire := func(posToCheck common.Vec2) bool {
		tcell := w.GetVec(posToCheck)
		if tcell == nil {
			return false
		}
		if (tcell.CellType == cell.WOOD_CELL || tcell.CellType == cell.LEAF_CELL) && (*c).RemainingLife != 0 {
			(*c).RemainingLife = 3
			return true
		}
		return false
	}
	return w.simulateCustomMovements(pos, maxSpeed, c, light_fire, change_color_cell_fire, groups)
}

func (w *ServerWorld) simulateLavaMovements(
	idChunk int,
	pos common.Vec2,
	maxSpeed int,
	c **cell.Cell,
	groups []common.Vec2,
) bool {
	afterMoving := func(pos common.Vec2) error {
		w.GetVec(pos).Touched()
		w.activeChunks.SortedInsert(idChunk)
		w.SetVec(pos, cell.NewCell(cell.EMPTY_CELL))
		return nil
	}
	light_flammable_create_smoke := func(posToCheck common.Vec2) bool {
		tcell := w.GetVec(posToCheck)
		if tcell == nil {
			return false
		}
		if tcell.CellType == cell.WATER_CELL {
			smoke, _ := NewCellByChance(cell.SMOKE_CELL, 10)
			w.SetVec(posToCheck, smoke)
			return false
		}
		if tcell.CellType == cell.WOOD_CELL || tcell.CellType == cell.LEAF_CELL {
			w.SetVec(posToCheck, cell.NewCell(cell.FIRE_CELL))
			return false
		}
		return w.isFree(posToCheck)
	}
	return w.simulateCustomMovements(pos, maxSpeed, c, light_flammable_create_smoke, afterMoving, groups)
}
