package main

import (
	"fmt"
	"io"
	"net"
	sandmmo "sand-mmo"
	"sand-mmo/common"
	chain "sand-mmo/responsibilityChain"
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
	engine := chain.NewResponsibilityChainEngine(world, chain.GetHandlers(), conn)
	for {
		r, err := common.ReadFromTcpSocket(conn)
		if err != nil {
			if err == io.EOF {
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
