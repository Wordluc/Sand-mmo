package core

import (
	"context"
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
	w.redis.Set(ctx, REDIS_KEY_BYTES_BYTES, string(worldBytes), 0)
	generatorsBytes := w.world.GetGeneratorsBytes()
	w.redis.Set(ctx, REDIS_KEY_BYTES_GENERATOR, string(generatorsBytes), 0)
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
