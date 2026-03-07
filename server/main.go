package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	ws "github.com/gorilla/websocket"

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

var upgrader = ws.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handler(w http.ResponseWriter, r *http.Request) {
	add := r.RemoteAddr
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed establishing connection with ", add)
		return
	}
	m.Lock()
	webSockets[add] = conn
	m.Unlock()
	go handlerConnection(conn)

}

func main() {
	http.HandleFunc("/ws", handler)
	w = new(world.NewServerWorld(common.W_WINDOWS, common.H_WINDOWS, common.CHUNK_SIZE))
	go UpdateClientWorlds()
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		panic(err)
	}

}

func handlerConnection(conn *ws.Conn) {
	defer conn.Close()
	engine := chain.NewResponsibilityChainEngine(w, chain.GetHandlers(), conn)

	for {
		r, err := common.ReadFromWebSocketPackage(conn)
		if err != nil {
			fmt.Printf("Error receiving %s\n", err.Error())
			continue
		}
		m.Lock()
		err = engine.Run(r)
		m.Unlock()
		if err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) {
				delete(webSockets, conn.RemoteAddr().String())
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
			for _, chunk := range chunks {
				for _, ws := range webSockets {
					if ws == nil {
						continue
					}
					err := ws.WriteMessage(websocket.TextMessage, chunk)
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	}()
}
