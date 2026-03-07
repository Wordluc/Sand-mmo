package responsibilityChain

import (
	"fmt"
	"sand-mmo/common"
	"sand-mmo/world"

	ws "github.com/gorilla/websocket"
)

type ResponsibilityChain struct {
	ps          []Handler
	i           int
	world       *world.ServerWorld
	tcpConn     *ws.Conn
	LastCommand common.Command
}

func NewResponsibilityChainEngine(world *world.ServerWorld, ps []Handler, tcpConn *ws.Conn) (res ResponsibilityChain) {
	res.ps = ps
	res.world = world
	res.tcpConn = tcpConn
	return res
}

func (pm *ResponsibilityChain) Run(p common.Package) error {
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
	pm.LastCommand = p.Command
	return err
}

type Handler struct {
	p       common.Command
	handler func(p common.Package, e *ResponsibilityChain) error
}

func (ph Handler) check(p common.Package) bool {
	if ph.p != p.Command {
		return false
	}
	return true
}

func (ph Handler) run(p common.Package, e *ResponsibilityChain) error {
	return ph.handler(p, e)
}
