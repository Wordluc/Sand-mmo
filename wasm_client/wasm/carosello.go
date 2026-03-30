package wasm

import (
	"sand-mmo/cell"
	"syscall/js"
)

type ButtonDef struct {
	text     string
	cellType cell.CellType
}

func (b ButtonDef) label() string {
	return b.text
}

type buttonCarosello interface {
	label() string
}
type callbackCarosello = func(button js.Value, idxButton, idxData int, isSelect bool)
type Carosello[t buttonCarosello] struct {
	data         []t
	idxLogic     int
	idxSelect    int
	totalButtons int
	totalData    int
	buttons      js.Value
	callback     callbackCarosello
}

func newCarosello[t buttonCarosello](data []t, callback callbackCarosello) *Carosello[t] {
	doc := js.Global().Get("document")
	var c Carosello[t]
	c.buttons = doc.Call("getElementsByClassName", "carosello-buttons")
	c.totalButtons = c.buttons.Get("length").Int()
	c.data = data
	c.totalData = len(data)
	c.callback = callback
	return &c
}

func (c *Carosello[t]) loop() {
	drawButton := func(item js.Value, i, idxData int) {
		if i == c.idxSelect {
			item.Get("style").Set("borderColor", "red")
			c.callback(item, i, idxData, true)
		} else {
			item.Get("style").Set("borderColor", "black")
		}
		item.Set("textContent", c.data[idxData].label())
	}
	drawAllButtons := func() {
		for i := 0; i < c.totalButtons; i++ {
			idxData := ((c.idxLogic + i) % c.totalData)
			item := c.buttons.Call("item", i)
			drawButton(item, i, idxData)
		}
	}
	for i := 0; i < c.totalButtons; i++ {
		item := c.buttons.Call("item", i)
		idxData := ((c.idxLogic + i) % c.totalData)
		drawButton(item, i, idxData)
		ti := i
		item.Set("onclick", js.FuncOf(func(this js.Value, args []js.Value) any {
			//TODO:to fix
			if c.idxLogic < 0 {
				c.idxLogic += c.totalData
			}
			c.idxSelect = ti
			drawAllButtons()
			return nil
		}))
	}
}

func (c *Carosello[t]) move(this js.Value, args []js.Value) interface{} {
	by := args[0].Int()
	c.idxSelect += by

	if c.idxSelect < 0 {
		c.idxSelect = (c.idxSelect % c.totalButtons)
		c.idxLogic = (c.idxLogic - c.totalButtons) % c.totalData
	} else if c.idxSelect >= c.totalButtons {
		c.idxSelect = (c.idxSelect % c.totalButtons)
		c.idxLogic = (c.idxLogic + c.totalButtons) % c.totalData
	}

	if c.idxSelect < 0 {
		c.idxSelect += c.totalButtons
	}
	if c.idxLogic < 0 {
		c.idxLogic += c.totalData
	}
	c.loop()
	return nil
}

var buttons = []ButtonDef{
	{text: "Void", cellType: cell.VACUUM_CELL},
	{text: "Water", cellType: cell.WATER_CELL},
	{text: "Sand", cellType: cell.SAND_CELL},
	{text: "Wood", cellType: cell.WOOD_CELL},
	{text: "Leaf", cellType: cell.LEAF_CELL},
	{text: "Stone", cellType: cell.STONE_CELL},
	{text: "Smoke", cellType: cell.SMOKE_CELL},
	{text: "Fire", cellType: cell.FIRE_CELL},
	{text: "Lava", cellType: cell.LAVA_CELL},
}

func (state *WasmState) InitCarosello() {
	var c *Carosello[ButtonDef]
	c = newCarosello(buttons,
		func(button js.Value, idxButton, idxData int, isSelect bool) {
			state.CellType = c.data[idxData].cellType
		})

	js.Global().Set("move_carosello", js.FuncOf(c.move))
	c.loop()
}
