package main

import (
	"fmt"
	"io"
	"net"
	sandmmo "sand-mmo"
	"sand-mmo/common"
	chain "sand-mmo/responsibilityChain"
	"time"
)

var world *sandmmo.World
var upds *[]net.Conn = &[]net.Conn{}

func main() {
	n, err := net.Listen("tcp", ":8000")
	if err != nil {
		panic(err)
	}
	t := sandmmo.NewWorld(sandmmo.W_WINDOWS, sandmmo.H_WINDOWS, sandmmo.CHUNK_SIZE)
	world = &t
	fmt.Println("Server setup ...")

	UpdateClientWorlds(world)
	for {
		conn, err := n.Accept()
		if err != nil {
			fmt.Printf("Error connecting %e", err)
			continue
		}
		go handlerConnection(conn)
	}

}
func callbackUdp(conn net.Conn) {
	*upds = append(*upds, conn)
}
func handlerConnection(conn net.Conn) {
	fmt.Printf("New connection %v\n", conn.RemoteAddr())
	defer conn.Close()
	engine := chain.NewResponsibilityChainEngine(world, chain.GetHandlers(), conn)
	engine.SetCallbackInitUdp(callbackUdp)
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
func UpdateClientWorlds(world *sandmmo.World) {
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			chunksToSend := world.GetTouchedChunks()
			for _, iC := range chunksToSend {
				chunk := world.GetChunkBytes(uint16(iC))
				for _, udp := range *upds {
					go udp.Write(chunk)
				}
			}
		}
	}()
}
