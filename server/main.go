package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/coder/websocket"
	ws "github.com/coder/websocket"

	"io"
	"math/rand"
	"sand-mmo/common"
	chain "sand-mmo/responsibilityChain"
	"sand-mmo/world"
	"sync"
	"time"
)

var w *world.ServerWorld
var m sync.Mutex
var webSockets map[string]*ws.Conn = map[string]*ws.Conn{}

func handler(w http.ResponseWriter, r *http.Request) {
	c, err := ws.Accept(w, r, &ws.AcceptOptions{
		InsecureSkipVerify: true, // allow all origins for dev
	})
	if err != nil {
		fmt.Println(err)
	}
	m.Lock()
	addr := r.RemoteAddr
	webSockets[addr] = c
	m.Unlock()
	if len(webSockets) == 1 {
		go UpdateClientWorlds()
	}
	go handlerConnection(c, addr)

}

func main() {
	http.HandleFunc("/ws", handler)
	w = new(world.NewServerWorld(common.W_WINDOWS, common.H_WINDOWS, common.CHUNK_SIZE))
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		panic(err)
	}

}

func handlerConnection(conn *ws.Conn, addr string) {
	defer conn.CloseNow()
	engine := chain.NewResponsibilityChainEngine(w, chain.GetHandlers(), conn)

	for {
		r, err := common.ReadFromWebSocketPackage(conn)
		if err != nil {
			fmt.Printf("Error receiving %s\n", err.Error())
			if strings.Contains(err.Error(), "EOF") {
				fmt.Println("End " + addr)
			}
			return
		}
		m.Lock()
		err = engine.Run(r)
		m.Unlock()
		if err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) {
				return
			}
			fmt.Print(err.Error(), "\n\n")
		}
	}
}
func UpdateClientWorlds() {
	var waitG sync.WaitGroup
	go func() {
		for {
			time.Sleep(common.SLEEP * time.Millisecond)

			m.Lock()
			if len(webSockets) == 0 {
				m.Unlock()
				return
			}
			err := w.ApplyGenerators()
			if err != nil {
				fmt.Println(err)
			}
			chunksToSend := w.GetActiveChunksAndNeiboroud()
			for _, iC := range chunksToSend {
				w.Simulate(uint16(iC))
			}
			chunksToSend = w.GetChunksToSend()
			t := uint8(rand.Intn(256))
			for {
				if t == common.GTouchedId {
					t = uint8(rand.Intn(256))
					continue
				}
				break
			}
			common.GTouchedId = t

			var chunks [][]byte = make([][]byte, len(chunksToSend))
			waitG.Add(len(chunksToSend))
			for i, iC := range chunksToSend {
				go func() {
					chunks[i] = w.GetChunkBytesToSend(uint16(iC))
					waitG.Done()
				}()
			}
			waitG.Wait()
			m.Unlock()
			waitG.Add(len(webSockets))
			for addr, ws := range webSockets {
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
					defer cancel()
					var err error
					defer func() {
						waitG.Done()
						if err != nil {
							delete(webSockets, addr)
							return
						}
					}()
					for _, chunk := range chunks {
						if ws == nil {
							continue
						}
						err = ws.Write(ctx, websocket.MessageBinary, chunk)
						if err != nil {
							fmt.Println(err)
							return
						}
					}
				}()
			}
			waitG.Wait()
		}
	}()
}
