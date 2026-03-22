package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	ws "github.com/coder/websocket"
	"github.com/joho/godotenv"

	"sand-mmo/common"
	"sand-mmo/core"
	handlers "sand-mmo/core/handlers"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"net/http/pprof"
)

var w *core.ServerWorld
var netCode *core.NetCode
var m *sync.Mutex = &sync.Mutex{}
var schedulerSaving common.Scheduler
var schedulerLoop common.Scheduler
var err error

func handler(write http.ResponseWriter, r *http.Request) {
	c, err := ws.Accept(write, r, &ws.AcceptOptions{
		InsecureSkipVerify: true, // allow all origins for dev
	})
	if err != nil {
		fmt.Println(err)
	}
	client := netCode.AddClient(r.RemoteAddr, c)
	if netCode.GetLenClients() == 1 {
		go StartLoop()
	}
	fmt.Println("N: ", netCode.GetLenClients())
	go handlerConnection(client)

}

func getRedisClient() (*redis.Client, error) {
	println("Redis: ", fmt.Sprintf("%v:%v", os.Getenv("redis_address"), os.Getenv("redis_port")))
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", os.Getenv("redis_address"), os.Getenv("redis_port")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0, // use default DB
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
	w = new(core.NewServerWorld(common.W_CELLS_TOTAL, common.H_CELLS_TOTAL, common.CHUNK_SIZE))
	netCode = new(core.NewNetCode(w, redis))
	if os.Getenv("load") == "true" {
		netCode.LoadSnapshot()
	}

	schedulerSaving = common.NewTimer(time.Minute, "TimerSaving", netCode.SaveSnapshot)
	schedulerLoop = common.NewTimer(common.SLEEP*time.Millisecond, "TimerLoop", loop)
	http.HandleFunc("/profile", pprof.Profile)
	http.HandleFunc("/ws", handler)
	err = http.ListenAndServe(":8000", nil)
	if err != nil {
		panic(err)
	}

}

func handlerConnection(client *core.Client) {
	defer netCode.RemoveClient(client)
	engine := handlers.NewCoreHandlers(w, handlers.GetHandlers(), client, netCode)

	for {
		r, err := common.ReadFromWebSocketPackage(client.Conn)
		if err != nil {
			fmt.Println("Error ", client.Addr, ": ", err.Error())
			return
		}
		m.Lock()
		err = engine.Run(r)
		m.Unlock()
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func loop() {
	if netCode.GetLenClients() == 0 {
		defer StopLoop()
		return
	}
	m.Lock()
	err = w.Loop()
	if err != nil {
		fmt.Println(err.Error())
	}
	netCode.SendChunks(w.GetChunksToSend())
	common.UntouchEverything()
	m.Unlock()

}

func StopLoop() {
	schedulerSaving.Stop()
	schedulerLoop.Stop()
}

func StartLoop() {
	if os.Getenv("save") == "true" {
		schedulerSaving.Start()
	}
	schedulerLoop.Start()
}
