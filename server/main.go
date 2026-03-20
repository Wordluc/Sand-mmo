package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/coder/websocket"
	ws "github.com/coder/websocket"
	"github.com/joho/godotenv"

	"io"
	"sand-mmo/common"
	"sand-mmo/core"
	handlers "sand-mmo/core/handlers"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var w *core.ServerWorld
var netCode *core.NetCode
var m *sync.Mutex = &sync.Mutex{}

func handler(write http.ResponseWriter, r *http.Request) {
	c, err := ws.Accept(write, r, &ws.AcceptOptions{
		InsecureSkipVerify: true, // allow all origins for dev
	})
	if err != nil {
		fmt.Println(err)
	}
	if netCode.AddClient(r.RemoteAddr, c) == 1 {
		go UpdateClientWorlds()
	}
	fmt.Println("N: ", netCode.GetLenClients())
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
	w = new(core.NewServerWorld(common.W_WINDOWS, common.H_WINDOWS, common.CHUNK_SIZE))
	netCode = new(core.NewNetCode(w, redis))

	http.HandleFunc("/ws", handler)
	err = http.ListenAndServe(":8000", nil)
	if err != nil {
		panic(err)
	}

}

func handlerConnection(conn *ws.Conn, addr string) {
	defer conn.CloseNow()
	defer netCode.RemoveClient(addr)
	engine := handlers.NewCoreHandlers(w, handlers.GetHandlers(), conn)

	for {
		r, err := common.ReadFromWebSocketPackage(conn)
		if err != nil {
			fmt.Println("Error ", addr, ": ", err.Error())
			netCode.RemoveClient(addr)
			return
		}
		m.Lock()
		err = engine.Run(r)
		m.Unlock()
		if err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) {
				return
			}
			fmt.Println(err.Error())
		}
	}
}

var timerSaving common.Timer
var timerLoop common.Timer

func loop() {
	var waitG sync.WaitGroup
	if netCode.GetLenClients() == 0 {
		defer timerSaving.Stop()
		defer timerLoop.Stop()
		return
	}
	m.Lock()
	err := w.Loop()
	if err != nil {
		fmt.Println(err.Error())
	}
	chunksToSend := w.GetChunksToSend()
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
	waitG.Add(netCode.GetLenClients())
	for addr, ws := range netCode.GetClients() {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), common.SLEEP*time.Millisecond)
			defer cancel()
			var err error
			defer func() {
				waitG.Done()
				if err != nil {
					fmt.Println("Removing for :", err.Error())
					netCode.RemoveClient(addr)
					return
				}
			}()
			for _, chunk := range chunks {
				if ws == nil {
					continue
				}
				err = ws.Write(ctx, websocket.MessageBinary, chunk)
				if err != nil {
					return
				}
			}
		}()
	}
	waitG.Wait()
}

func UpdateClientWorlds() {
	timerSaving = common.NewTimer(time.Minute, netCode.SaveSnapshot)
	timerLoop = common.NewTimer(common.SLEEP*time.Millisecond, loop)
	timerSaving.Start()
	timerLoop.Start()
}
