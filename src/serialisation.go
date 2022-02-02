package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"time"
)

// From wikipedia: https://en.wikipedia.org/wiki/Variable-length_quantity
type vlqInt uint64

var deSerialize = struct {
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
		var message Message
		err := binary.Read(buf, binary.BigEndian, &message)
		lenPeers, _, n := variantLengthQuantity(b[binary.Size(message):])
		offset := binary.Size(message) + n
		size := binary.Size(PeerMessage{})
		var peerMessage []PeerMessage
		for i := offset; i < (size*int(lenPeers))+offset; i += size {
			buf = bytes.NewBuffer(b[i:])
			var peer PeerMessage
			checkError(binary.Read(buf, binary.BigEndian, &peer))
			peerMessage = append(peerMessage, peer)
		}
		peersMessage := PeersMessage{
			Header:   message.Header,
			Type:     message.Type,
			Version:  message.Version,
			LenPeers: lenPeers,
			Peers:    peerMessage,
		}
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
var serialize = struct {
	Hello            func(HelloMessage) []byte
	Peers            func() []byte
	DataBlockMessage func(DataBlockMessage) []byte
}{
	Hello: func(helloMessage HelloMessage) []byte {
		buf := new(bytes.Buffer)
		checkError(binary.Write(buf, binary.BigEndian, helloMessage))
		return buf.Bytes()
	},
	Peers: func() []byte {
		const maxPeers = 63
		peerMessage := [maxPeers]PeerMessage{}
		i := 0
		for _, p := range peers {
			if p.Connected && i < maxPeers {
				peerMessage[i] = PeerMessage{
					LastSeen: uint32(time.Now().Unix()),
					IP:       p.IP,
					Port:     p.Port,
				}
				i++
			}
		}
		peersMessage := struct {
			Header   Header
			Type     uint16
			Version  byte
			LenPeers uint8 // Variable l q
			Peers    [maxPeers]PeerMessage
		}{
			Header:   constructHeader(),
			Type:     MSG.Peers,
			Version:  0,
			LenPeers: uint8(i),
			Peers:    peerMessage,
		}
		peersMessage.Header.Len = uint32(
			binary.Size(peersMessage),
		) - 8
		buf := new(bytes.Buffer)
		checkError(binary.Write(buf, binary.BigEndian, peersMessage))
		return buf.Bytes()
	},
}

func DESERIALIZE(byteArr []byte, t interface{}) int {
	fields := reflect.TypeOf(t)
	// The application must call Elem() twice to get the struct value
	values := reflect.ValueOf(&t).Elem()
	num := reflect.ValueOf(t).NumField()
	counter := 0

	for i := 0; i < num; i++ {
		field := fields.Field(i)
		value := values.Elem().Field(i)
		tmp := reflect.New(values.Elem().Type()).Elem()
		fmt.Println(field.Name, value.Type().Size())

		switch value.Kind() {
		case reflect.Struct:
			counter += DESERIALIZE(byteArr[counter:], value.Interface())
		case reflect.Uint8, reflect.Uint16, reflect.Uint64:
			size := int(value.Type().Size())
			counter += size
		case reflect.Uint32:
			fmt.Println(tmp.Field(i).CanSet())
			tmp.Field(i).SetUint(uint64(binary.BigEndian.Uint32(byteArr[counter:])))
			values.Set(tmp)
			size := int(value.Type().Size())
			counter += size
		case reflect.Array:
			counter += value.Len()
		}
	}
	fmt.Println(counter, len(byteArr))
	return counter
}

func variantLengthQuantity(byteArr []byte) (vlqInt, []byte, int) {
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
	return vlqInt(result), buf.Bytes(), len(variant)
}
