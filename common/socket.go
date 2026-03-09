package common

import (
	"context"
	"encoding/binary"

	ws "github.com/coder/websocket"
)

// TODO: create the udp version
func ReadFromWebSocketPackage(socket *ws.Conn) (res Package, err error) {
	_, brush, err := socket.Read(context.Background())
	if err != nil {
		return res, err
	}
	t := binary.BigEndian.Uint64(brush)
	return Decode(t), nil
}

// TODO: create the udp version
func SendToWebSocketPackages(socket *ws.Conn, ps ...Package) error {
	for _, p := range ps {
		u64 := Encode(p)
		var toSend []byte = make([]byte, 8)
		binary.BigEndian.PutUint64(toSend, u64)
		err := socket.Write(context.Background(), ws.MessageBinary, toSend)
		if err != nil {
			return err
		}
	}

	return nil
}
