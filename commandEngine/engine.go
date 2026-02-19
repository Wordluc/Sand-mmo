package commandengine

import (
	"fmt"
	sandmmo "sand-mmo"
	"sand-mmo/common"
)

type PackageEngine struct {
	ps    []PackageHandler
	i     int
	world *sandmmo.World
}

func NewPackageEngine(world *sandmmo.World, ps []PackageHandler) (res PackageEngine) {
	res.ps = ps
	res.world = world
	return res
}

func (pm *PackageEngine) Run(p common.Package) error {
	if pm.i >= len(pm.ps) {
		return fmt.Errorf("Handler not found for: %x", p.Code)
	}
	if !pm.ps[pm.i].check(p) {
		pm.i += 1
		return pm.Run(p)
	}

	err := pm.ps[pm.i].run(p, pm.world)
	pm.i = 0
	return err
}

type PackageHandler struct {
	p       common.Package
	handler func(p common.Package, w *sandmmo.World) error
}

func (ph PackageHandler) check(p common.Package) bool {
	if ph.p.Command != p.Command {
		return false
	}
	if ph.p.CommandPackage.Ident != p.Ident {
		return false
	}
	return true
}

func (ph PackageHandler) run(p common.Package, w *sandmmo.World) error {
	return ph.handler(p, w)
}
