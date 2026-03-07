package common

import (
	"encoding/binary"

	ws "github.com/gorilla/websocket"
)

// TODO: create the udp version
func ReadFromTcpSocket(socket *ws.Conn) (res Package, err error) {
	_, brush, err := socket.ReadMessage()
	if err != nil {
		return res, err
	}
	t := binary.BigEndian.Uint64(brush)
	return Decode(t), nil
}

// TODO: create the udp version
func SendToTcpSocket(p Package, socket *ws.Conn) error {
	u64 := Encode(p)
	var toSend []byte = make([]byte, 8)
	binary.BigEndian.PutUint64(toSend, u64)
	err := socket.WriteMessage(ws.TextMessage, toSend)

	return err
}
