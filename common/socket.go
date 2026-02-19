package common

import (
	"encoding/binary"
	"fmt"
	"net"
)

func ReadFromSocket(socket net.Conn) (res Package, err error) {
	var brush []byte = make([]byte, 4)
	n, err := socket.Read(brush)
	if err != nil {
		return res, err
	}
	if n != 4 {
		return res, err
	}
	t := binary.BigEndian.Uint32(brush)
	return Decode(t), nil
}

func SendToSocket(p uint32, socket net.Conn) error {
	var toSend []byte = make([]byte, 4)
	binary.BigEndian.PutUint32(toSend, p)
	fmt.Printf("Sending %x\n", toSend)
	_, err := socket.Write([]byte(toSend))

	return err
}
