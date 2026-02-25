package main

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	sandmmo "sand-mmo"
	"sand-mmo/common"
	chain "sand-mmo/responsibilityChain"
	"time"
)

var world *sandmmo.World
var addrsUdp map[net.Addr]net.Addr = map[net.Addr]net.Addr{}
var udp *net.UDPConn

func main() {
	n, err := net.Listen("tcp", ":8000")
	if err != nil {
		panic(err)
	}
	addrUdp, _ := net.ResolveUDPAddr("udp", ":8000")
	udp, err = net.ListenUDP("udp", addrUdp)
	if err != nil {
		panic(err)
	}
	t := sandmmo.NewWorld(common.W_WINDOWS, common.H_WINDOWS, common.CHUNK_SIZE)
	world = &t
	fmt.Println("Server setup ...")
	fmt.Printf("Tcp: %v, Udp: %v\n", n.Addr(), udp.LocalAddr())

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
func callbackAddUdp(tcp, udp net.Addr) {
	addrsUdp[tcp] = udp
}
func callbackRemoveUdp(tcp net.Addr) {
	delete(addrsUdp, tcp)
}
func handlerConnection(conn net.Conn) {
	fmt.Printf("New tcp connection with %v\n", conn.RemoteAddr())
	defer conn.Close()
	engine := chain.NewResponsibilityChainEngine(world, chain.GetHandlers(), conn, udp)
	engine.SetCallbackAddUdp(callbackAddUdp)
	engine.SetCallbackRemoveUdp(callbackRemoveUdp)
	for {
		r, err := common.ReadFromTcpSocket(conn)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Closing tcp connection with %v\n", conn.RemoteAddr())
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
			time.Sleep(common.SLEEP * time.Millisecond)
			sandmmo.GTouchedId = uint8(rand.Intn(256))
			//Lock world to out communications
			//Loop simulation
			chunksToSend := world.GetActiveChunksAndNeiboroud()
			//UnLock
			for _, iC := range chunksToSend {
				world.Simulate(uint16(iC))
			}
			chunksToSend = world.GetChunksToSend()
			for _, iC := range chunksToSend {
				chunk := world.GetChunkBytesToSend(uint16(iC))
				for _, addr := range addrsUdp {
					if addr == nil {
						continue
					}
					_, err := udp.WriteTo(chunk, addr.(*net.UDPAddr))
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	}()
}
