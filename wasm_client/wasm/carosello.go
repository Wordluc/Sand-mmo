package wasm

import (
	"syscall/js"
)

type callbackCarosello = func(button js.Value, idxButton, idxData int, isSelect bool)
type Carosello struct {
	data         []string
	idxLogic     int
	idxSelect    int
	totalButtons int
	totalData    int
	buttons      js.Value
	callback     callbackCarosello
}

func newCarosello(data []string, callback callbackCarosello) *Carosello {
	doc := js.Global().Get("document")
	var c Carosello
	c.buttons = doc.Call("getElementsByClassName", "carosello-buttons")
	c.totalButtons = c.buttons.Get("length").Int()
	c.data = data
	c.totalData = len(data)
	c.callback = callback
	return &c
}

func (c *Carosello) loop() {
	var i = 0
	drawButton := func(item js.Value, idxData int) {
		if i == c.idxSelect {
			item.Get("style").Set("borderColor", "red")
			c.callback(item, i, idxData, true)
		} else {
			item.Get("style").Set("borderColor", "black")
		}
	}
	drawAllButtons := func() {
		for i = 0; i < c.totalButtons; i++ {
			idxData := ((c.idxLogic + i) % c.totalData)
			item := c.buttons.Call("item", i)
			drawButton(item, idxData)
		}
	}
	for i = 0; i < c.totalButtons; i++ {
		item := c.buttons.Call("item", i)
		idxData := ((c.idxLogic + i) % c.totalData)
		drawButton(item, idxData)
		ti := i
		item.Set("onclick", js.FuncOf(func(this js.Value, args []js.Value) any {
			println(c.idxSelect)
			//TODO:to fix
			c.idxLogic = (c.idxSelect - ti) % c.totalData
			c.idxSelect = ti
			drawAllButtons()
			return nil
		}))
		item.Set("textContent", c.data[idxData])
	}
}

func (c *Carosello) move(this js.Value, args []js.Value) interface{} {
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

func (WasmState) InitCarosello() {
	var c *Carosello
	c = newCarosello([]string{"Prova1", "Prova2", "Prova3", "Prova4", "Prova5"},
		func(button js.Value, idxButton, idxData int, isSelect bool) {
			println(c.data[idxData])

		})

	js.Global().Set("move_carosello", js.FuncOf(c.move))
	c.loop()
}
