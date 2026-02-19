package commandengine

import (
	"fmt"
	sandmmo "sand-mmo"
	"sand-mmo/common"
)

func GetHandlers() []PackageHandler {
	return []PackageHandler{
		{
			p: GetChunkCommand(0),
			handler: func(p common.Package, _ *sandmmo.World) error {
				fmt.Println("ReturnChunk")
				return nil
			},
		},
		{
			p: GetDrawCommand(0, 0, 0),
			handler: func(p common.Package, _ *sandmmo.World) error {
				fmt.Println("Draw")
				return nil
			},
		},
		{
			p: GetInitCommand(),
			handler: func(p common.Package, _ *sandmmo.World) error {
				fmt.Println("Init")
				return nil
			},
		},
	}
}
