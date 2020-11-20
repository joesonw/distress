package main

import (
	"encoding/binary"
	"io"
	"net"

	"github.com/golang/protobuf/proto"

	"github.com/joesonw/distress/examples/netproto/message"
)

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			panic(err)
		}

		go handle(conn)
	}
}

func handle(conn net.Conn) {
	for {
		var b []byte

		b = make([]byte, 4)
		_, err := conn.Read(b)
		if err == io.EOF {
			return
		}
		if err != nil {
			panic(err)
		}
		size := binary.BigEndian.Uint32(b)
		b = make([]byte, int(size))

		_, err = conn.Read(b)
		if err == io.EOF {
			return
		}
		if err != nil {
			panic(err)
		}

		m := message.Message{}
		if err := proto.Unmarshal(b, &m); err != nil {
			panic(err)
		}

		m.Body = "you said: " + m.Body
		b, err = proto.Marshal(&m)
		if err != nil {
			panic(err)
		}

		b2 := make([]byte, 4+len(b))
		binary.BigEndian.PutUint32(b2, uint32(len(b)))
		copy(b2[4:], b)
		_, err = conn.Write(b2)
		if err != nil {
			panic(err)
		}
	}
}
