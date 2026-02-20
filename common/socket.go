package common

import (
	"encoding/binary"
	"fmt"
	"net"
)

// TODO: create the udp version
func ReadFromTcpSocket(socket net.Conn) (res Package, err error) {
	var brush []byte = make([]byte, 8)
	n, err := socket.Read(brush)
	if err != nil {
		return res, err
	}
	if n != 8 {
		return res, err
	}
	t := binary.BigEndian.Uint64(brush)
	return Decode(t), nil
}

// TODO: create the udp version
func SendToTcpSocket(p Package, socket net.Conn) error {
	u64 := Encode(p)
	var toSend []byte = make([]byte, 8)
	binary.BigEndian.PutUint64(toSend, u64)
	fmt.Printf("Sending %x\n", toSend)
	_, err := socket.Write([]byte(toSend))

	return err
}
