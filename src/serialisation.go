package main

import (
	"bytes"
	"encoding/binary"
	"math"
)

var Serialisation = struct {
	Hello            func([]byte) HelloMessage
	Message          func([]byte) Message
	Peers            func([]byte) (PeersMessage, error)
	DataBlockMessage func([]byte) DataBlockMessage
}{
	Hello: func(b []byte) HelloMessage {
		buf := bytes.NewBuffer(b)
		var helloMessage HelloMessage
		checkError(binary.Read(buf, binary.BigEndian, &helloMessage))
		return helloMessage
	},
	Message: func(b []byte) Message {
		buf := bytes.NewBuffer(b)
		var message Message
		checkError(binary.Read(buf, binary.BigEndian, &message))
		return message
	},
	Peers: func(b []byte) (PeersMessage, error) {
		buf := bytes.NewBuffer(b)
		var peersMessage PeersMessage
		err := binary.Read(buf, binary.BigEndian, &peersMessage)
		return peersMessage, err
	},
	DataBlockMessage: func(b []byte) DataBlockMessage {
		// DataBlockMessage Height is VLQ it needs to be converted to fixed sized [uint32]
		offset := binary.Size(DataMessage{}) + 1
		_, buf, n := variantLengthQuantity(b[offset:])
		blockUint32Bytes := b[:offset]
		blockUint32Bytes = append(blockUint32Bytes, buf...)
		blockUint32Bytes = append(
			blockUint32Bytes,
			b[offset+n:]...,
		)
		reader := bytes.NewReader(blockUint32Bytes)
		var dataBlock DataBlockMessage
		checkError(binary.Read(reader, binary.BigEndian, &dataBlock))
		return dataBlock
	},
}

func variantLengthQuantity(byteArr []byte) (uint32, []byte, int) {
	var variant []uint32
	i := 0
	for {
		b := binary.BigEndian.Uint32([]byte{0, 0, 0, byteArr[i]})
		if b < 128 {
			variant = append(variant, b)
			break
		} else {
			b = binary.BigEndian.Uint32([]byte{0, 0, 0, byteArr[i]}) % 128
			variant = append(variant, b)
		}
		i++
	}
	var result uint32 = 0
	var p float64 = 0
	for i := len(variant) - 1; i > -1; i-- {
		result += variant[i] * uint32(math.Pow(128, p))
		p++
	}
	buf := new(bytes.Buffer)
	checkError(binary.Write(buf, binary.BigEndian, result))
	return result, buf.Bytes(), len(variant)
}
