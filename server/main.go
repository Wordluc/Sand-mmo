package main

import (
	"fmt"
	"net"
	sandmmo "sand-mmo"
	commandengine "sand-mmo/commandEngine"
	"sand-mmo/common"
)

var world *sandmmo.World

func main() {
	n, err := net.Listen("tcp", ":8000")
	if err != nil {
		panic(err)
	}
	t := sandmmo.NewWorld(100, 100)
	world = &t
	fmt.Println("Server setup ...")
	for {
		conn, err := n.Accept()
		if err != nil {
			fmt.Printf("Error connecting %e", err)
			continue
		}
		go handlerConnection(conn)
	}

}

func handlerConnection(conn net.Conn) {
	fmt.Printf("New connection %v\n", conn.RemoteAddr())
	defer conn.Close()
	engine := commandengine.NewPackageEngine(world, commandengine.GetHandlers())
	for {
		r, err := common.ReadFromSocket(conn)
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Printf("Closing connection %v\n", conn.RemoteAddr())
				return
			}
			fmt.Printf("Error receiving %s\n", err.Error())
			continue
		}
		err = engine.Run(r)
		if err != nil {
			fmt.Print(err.Error(), "\n\n")
			continue
		}
	}
}
