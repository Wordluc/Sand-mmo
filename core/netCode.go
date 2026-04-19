package core

import (
	"context"
	"fmt"
	"sand-mmo/common"
	"sync"
	"time"

	ws "github.com/coder/websocket"
	"github.com/redis/go-redis/v9"
)

const REDIS_KEY_WORLD = "world:bytes"
const REDIS_KEY_GENERATOR = "world:generator"
const REDIS_KEY_CLIENT_HISTORY = "client:history"

type NetCode struct {
	clients *sync.Map
	redis   *redis.Client
	world   *ServerWorld
}

type Client struct {
	Addr      string
	Conn      *ws.Conn
	AtChunkId int
	IsGod     bool
}

func NewNetCode(world *ServerWorld, redisClient *redis.Client) (res NetCode) {
	res.clients = &sync.Map{}
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
	err := w.redis.Set(ctx, REDIS_KEY_WORLD, string(worldBytes), 0).Err()
	if err != nil {
		println(err.Error())
		return
	}
	generatorsBytes := w.world.GetGeneratorsBytes()
	err = w.redis.Set(ctx, REDIS_KEY_GENERATOR, string(generatorsBytes), 0).Err()
	if err != nil {
		println(err.Error())
		return
	}
	println("World Saved")
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err = w.redis.ZAdd(ctx, REDIS_KEY_CLIENT_HISTORY, redis.Z{
		Score:  float64(time.Now().Unix()), // unix timestamp as score
		Member: fmt.Sprint(w.GetLenClients(), ";", time.Now().Format(time.DateTime)),
	}).Err()
	if err != nil {
		fmt.Println(err)
	}
	println("Save Clients Number")
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
	worldBytes, err := get(REDIS_KEY_WORLD)
	if err != nil {
		return err
	}
	w.world.ImportCells(worldBytes)

	generatorBytes, err := get(REDIS_KEY_GENERATOR)
	if err != nil {
		return err
	}
	w.world.ImportGenerators(generatorBytes)
	println("World Loaded")
	return nil
}

func (w *NetCode) AddClient(addr string, conn *ws.Conn) (c *Client) {
	c = &Client{
		Addr:      addr,
		Conn:      conn,
		AtChunkId: 0,
	}
	w.clients.Store(addr, c)
	return c
}

func (w *NetCode) RemoveClient(client *Client) {
	fmt.Println("Removed " + client.Addr)
	w.clients.Delete(client.Addr)
	client.Conn.CloseNow()
}

func (w *NetCode) GetLenClients() (count int) {
	w.clients.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}

func (w *NetCode) SendInitialChunks(client *Client) (err error) {
	defer func() {
		if err != nil {
			fmt.Println("Removing for :", err.Error())
			w.RemoveClient(client)
			return
		}
	}()
	xClient, yClient := common.GetServerXYChunk(client.AtChunkId)
	x, y := xClient, yClient
	idChunk := client.AtChunkId
	var chunks map[int][]byte = make(map[int][]byte, common.W_CHUNKS_CLIENT*common.H_CHUNKS_CLIENT)
	for {
		chunks[idChunk] = w.world.GetChunkBytesToSend(idChunk)
		x++
		if x >= xClient+common.W_CHUNKS_CLIENT {
			y++
			x = xClient
		}
		if y >= yClient+common.H_CHUNKS_CLIENT {
			break
		}
		idChunk = x + y*common.W_CHUNKS_TOTAL
	}
	return w.SendChunksTo(chunks, client)
}

func (w *NetCode) SendInitialChunksForGod(client *Client) (err error) {
	defer func() {
		if err != nil {
			fmt.Println("Removing for :", err.Error())
			w.RemoveClient(client)
			return
		}
	}()
	xClient, yClient := common.GetServerXYChunk(client.AtChunkId)
	x, y := xClient, yClient
	idChunk := client.AtChunkId
	var chunks map[int][]byte = make(map[int][]byte, common.W_CHUNKS_CLIENT*common.H_CHUNKS_CLIENT)
	for {
		chunks[idChunk] = w.world.GetChunkBytesToSend(idChunk)
		x++
		if x >= xClient+common.W_CHUNKS_TOTAL {
			y++
			x = xClient
		}
		if y >= yClient+common.H_CHUNKS_TOTAL {
			break
		}
		idChunk = x + y*common.W_CHUNKS_TOTAL
	}
	return w.SendChunksTo(chunks, client)
}
func (w *NetCode) SendVisibleChunksToAll(chunksToSend []int) {
	var chunks map[int][]byte = make(map[int][]byte, len(chunksToSend))
	for _, iC := range chunksToSend {
		chunks[iC] = w.world.GetChunkBytesToSend(iC)
	}
	clients := w.getClients()
	for _, client := range clients {
		if client.Conn == nil {
			continue
		}
		go func() {
			w.SendChunksTo(chunks, client)
		}()
	}
}

func (w *NetCode) SendChunksTo(chunksToSend map[int][]byte, client *Client) (err error) {
	xClient, yClient := common.GetServerXYChunk(client.AtChunkId)
	var x, y int
	var timeout time.Duration = 200
	if client.IsGod {
		timeout = 500
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Millisecond)
	defer cancel()
	defer func() {
		if err != nil {
			fmt.Println("Removing for :", err.Error())
			w.RemoveClient(client)
		}
	}()
	chunks_batched := []byte{}
	if len(chunksToSend) == 0 {
		return nil
	}
	for idChunk, chunk := range chunksToSend {
		if !client.IsGod {
			x, y = common.GetServerXYChunk(idChunk)
			if x < xClient-1 || x > xClient+common.W_CHUNKS_CLIENT {
				continue
			}
			if y < yClient-1 || y > yClient+common.H_CHUNKS_CLIENT {
				continue
			}
		}
		chunks_batched = append(chunks_batched, chunk...)
	}
	err = client.Conn.Write(ctx, ws.MessageBinary, chunks_batched)
	if err != nil {
		return
	}
	return err
}

func (w *NetCode) getClients() (conns map[string]*Client) {
	conns = map[string]*Client{}
	w.clients.Range(func(key, value any) bool {
		conns[key.(string)] = value.(*Client)
		return true
	})
	return conns
}
