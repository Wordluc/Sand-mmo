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
func SendToTcpSocket(p uint64, socket net.Conn) error {
	var toSend []byte = make([]byte, 8)
	binary.BigEndian.PutUint64(toSend, p)
	fmt.Printf("Sending %x\n", toSend)
	_, err := socket.Write([]byte(toSend))

	return err
}
