package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/coder/websocket"
	ws "github.com/coder/websocket"
	"github.com/joho/godotenv"

	"io"
	"sand-mmo/common"
	chain "sand-mmo/responsibilityChain"
	"sand-mmo/world"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var w *world.ServerWorld
var m *sync.Mutex = &sync.Mutex{}

func handler(write http.ResponseWriter, r *http.Request) {
	c, err := ws.Accept(write, r, &ws.AcceptOptions{
		InsecureSkipVerify: true, // allow all origins for dev
	})
	if err != nil {
		fmt.Println(err)
	}
	if w.AddClient(r.RemoteAddr, c) == 1 {
		go UpdateClientWorlds()
	}
	go handlerConnection(c, r.RemoteAddr)

}

func getRedisClient() (*redis.Client, error) {
	println("Redis: ", fmt.Sprintf("%v:%v", os.Getenv("redis_address"), os.Getenv("redis_port")))
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", os.Getenv("redis_address"), os.Getenv("redis_port")),
		Password: "", // no password
		DB:       0,  // use default DB
		Protocol: 2,
	})
	if client == nil {
		return nil, errors.New("Error creating redis client")
	}
	return client, nil
}

func main() {
	godotenv.Load(".env")
	redis, err := getRedisClient()
	if err != nil {
		panic(err)
	}
	w = new(world.NewServerWorld(common.W_WINDOWS, common.H_WINDOWS, common.CHUNK_SIZE, redis))

	http.HandleFunc("/ws", handler)
	err = http.ListenAndServe(":8000", nil)
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

			if w.GetLenSockets() == 0 {
				w.SaveSnapshot()

				return
			}
			m.Lock()
			err := w.ApplyGenerators()
			if err != nil {
				fmt.Println(err)
			}
			chunksToSend := w.GetActiveChunksAndNeiboroud()
			for _, iC := range chunksToSend {
				w.SimulateChunk(uint16(iC))
			}
			chunksToSend = w.GetChunksToSend()
			common.UntouchEverything()

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
			waitG.Add(w.GetLenSockets())
			for addr, ws := range w.GetClients() {
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), common.SLEEP*time.Millisecond)
					defer cancel()
					var err error
					defer func() {
						waitG.Done()
						if err != nil {
							w.RemoveClient(addr)
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
