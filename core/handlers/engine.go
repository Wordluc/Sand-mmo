package handlers

import (
	"fmt"
	"sand-mmo/common"
	"sand-mmo/core"
)

type CoreHandlers struct {
	ps          []handler
	i           int
	world       *core.ServerWorld
	client      *core.Client
	lastCommand common.Command
	netCode     *core.NetCode
}

func NewCoreHandlers(world *core.ServerWorld, ps []handler, client *core.Client, netCode *core.NetCode) (res CoreHandlers) {
	res.ps = ps
	res.world = world
	res.client = client
	res.netCode = netCode
	return res
}

func (pm *CoreHandlers) IsLastCommand(command common.Command) bool {
	return pm.lastCommand == command
}

func (pm *CoreHandlers) Run(p common.Package) error {
	defer func() {
		pm.i = 0
	}()
	if pm.i >= len(pm.ps) {
		return fmt.Errorf("Handler not found for: %x", p.Code)
	}
	if !pm.ps[pm.i].check(p) {
		pm.i += 1
		return pm.Run(p)
	}

	err := pm.ps[pm.i].run(p, pm)
	pm.lastCommand = p.Command
	return err
}

type handler struct {
	p       common.Command
	handler func(p common.Package, e *CoreHandlers) error
}

func (ph handler) check(p common.Package) bool {
	if ph.p != p.Command {
		return false
	}
	return true
}

func (ph handler) run(p common.Package, e *CoreHandlers) error {
	return ph.handler(p, e)
}
