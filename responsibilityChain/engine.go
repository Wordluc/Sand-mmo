package responsibilityChain

import (
	"fmt"
	"net"
	sandmmo "sand-mmo"
	"sand-mmo/common"
)

type ResponsibilityChain struct {
	ps                []Handler
	i                 int
	world             *sandmmo.World
	tcpConn           net.Conn
	udpConn           *net.UDPConn
	callbackAddUdp    func(net.Addr, net.Addr)
	callbackRemoveUdp func(net.Addr)
}

func NewResponsibilityChainEngine(world *sandmmo.World, ps []Handler, tcpConn net.Conn, udpConn *net.UDPConn) (res ResponsibilityChain) {
	res.ps = ps
	res.world = world
	res.tcpConn = tcpConn
	res.udpConn = udpConn
	return res
}

func (pm *ResponsibilityChain) SetCallbackAddUdp(callback func(net.Addr, net.Addr)) {
	pm.callbackAddUdp = callback
}

func (pm *ResponsibilityChain) SetCallbackRemoveUdp(callback func(net.Addr)) {
	pm.callbackRemoveUdp = callback
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
	return err
}

type Handler struct {
	p       common.Package
	handler func(p common.Package, e *ResponsibilityChain) error
}

func (ph Handler) check(p common.Package) bool {
	if ph.p.Command != p.Command {
		return false
	}
	return true
}

func (ph Handler) run(p common.Package, e *ResponsibilityChain) error {
	return ph.handler(p, e)
}
