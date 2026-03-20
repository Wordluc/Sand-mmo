package core

import (
	"context"
	"fmt"
	"maps"
	"sand-mmo/common"
	"sync"
	"time"

	ws "github.com/coder/websocket"
	"github.com/redis/go-redis/v9"
)

const REDIS_KEY_BYTES_BYTES = "world:bytes"
const REDIS_KEY_BYTES_GENERATOR = "world:generator"

type NetCode struct {
	webSockets     map[string]*ws.Conn
	webSocketMutex *sync.Mutex
	redis          *redis.Client
	world          *ServerWorld
}

func NewNetCode(world *ServerWorld, redisClient *redis.Client) (res NetCode) {
	res.webSockets = map[string]*ws.Conn{}
	res.webSocketMutex = &sync.Mutex{}
	res.redis = redisClient
	res.world = world
	return res
}

func (w *NetCode) SaveSnapshot() {
	if w.redis == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	worldBytes := w.world.GetWorldBytes()
	err := w.redis.Set(ctx, REDIS_KEY_BYTES_BYTES, string(worldBytes), 0).Err()
	if err != nil {
		println(err.Error())
		return
	}
	generatorsBytes := w.world.GetGeneratorsBytes()
	err = w.redis.Set(ctx, REDIS_KEY_BYTES_GENERATOR, string(generatorsBytes), 0).Err()
	if err != nil {
		println(err.Error())
		return
	}
	println("World Saved")
}

func (w *NetCode) LoadSnapshot() error {
	get := func(key string) ([]byte, error) {
		ctx, p := context.WithTimeout(context.Background(), common.SLEEP*time.Millisecond)
		defer p()
		worldBytes, err := w.redis.Get(ctx, key).Result()
		if err == nil {
			return []byte(worldBytes), nil
		}
		switch err {
		case redis.Nil:
			return []byte{}, nil
		default:
			return []byte{}, err
		}
	}
	worldBytes, err := get(REDIS_KEY_BYTES_BYTES)
	if err != nil {
		return err
	}
	w.world.ImportCells(worldBytes)

	generatorBytes, err := get(REDIS_KEY_BYTES_GENERATOR)
	if err != nil {
		return err
	}
	w.world.ImportGenerators(generatorBytes)
	return nil
}

func (w *NetCode) AddClient(addr string, conn *ws.Conn) int {
	w.webSocketMutex.Lock()
	defer w.webSocketMutex.Unlock()
	w.webSockets[addr] = conn
	return len(w.webSockets)
}

func (w *NetCode) RemoveClient(addr string) {
	w.webSocketMutex.Lock()
	defer w.webSocketMutex.Unlock()
	delete(w.webSockets, addr)
}

func (w *NetCode) GetLenClients() int {
	w.webSocketMutex.Lock()
	defer w.webSocketMutex.Unlock()
	return len(w.webSockets)
}

func (w *NetCode) GetClients() (conns map[string]*ws.Conn) {
	w.webSocketMutex.Lock()
	conns = maps.Clone(w.webSockets)
	w.webSocketMutex.Unlock()
	return conns
}

func (w *NetCode) SendChunks(chunksToSend []uint16) {
	var waitG sync.WaitGroup
	var chunks [][]byte = make([][]byte, len(chunksToSend))
	waitG.Add(len(chunksToSend))
	for i, iC := range chunksToSend {
		go func() {
			chunks[i] = w.world.GetChunkBytesToSend(uint16(iC))
			waitG.Done()
		}()
	}
	waitG.Wait()
	waitG.Add(w.GetLenClients())
	for addr, client := range w.GetClients() {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), common.SLEEP*time.Millisecond)
			defer cancel()
			var err error
			defer func() {
				waitG.Done()
				if err != nil {
					fmt.Println("Removing for :", err.Error())
					w.RemoveClient(addr)
					return
				}
			}()
			for _, chunk := range chunks {
				if client == nil {
					continue
				}
				err = client.Write(ctx, ws.MessageBinary, chunk)
				if err != nil {
					return
				}
			}
		}()
	}
	waitG.Wait()
}
