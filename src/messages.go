package main

import (
	"time"
)

var MSG = struct {
	Hello     uint16
	GetBlocks uint16
	Inventory uint16
	GetData   uint16
	Data      uint16
	GetPeers  uint16
	Peers     uint16
}{
	0, 1, 2, 3, 4, 5, 6,
}
var DATA = struct {
	Block       uint16
	Header      uint16
	Transaction uint16
}{
	0, 1, 2,
}

type Header struct {
	Magic      [4]byte
	Len        uint32
	Version    byte
	Timestamp  uint32
	Id         uint32
	ResponseId uint32
	Context    uint64
	Zero       [32]byte
}

func constructHeader() Header {
	header := Header{
		Magic:      [4]byte{'M', 'A', 'J', 'I'},
		Len:        56,
		Version:    0,
		Timestamp:  uint32(time.Now().Unix()),
		Id:         1,
		ResponseId: 0,
		Context:    0,
		Zero:       [32]byte{},
	}
	return header
}

type Message struct {
	Header  Header
	Type    uint16
	Version byte
}
type HelloMessage struct {
	Header           Header
	Type             uint16
	Version          byte
	IP               [16]byte
	Port             uint16
	MyIp             [18]byte
	Nonce            uint32
	LenUserAgent     byte
	UserAgent        [14]byte
	SupportedVersion byte
	Zeros            [256]byte
}
type GetBlocksMessage struct {
	Header       Header
	Type         uint16
	Version      byte
	LenStartHash uint8
	StartHashes  [32]byte
	Zeros        [32]byte
}
type InventoryItem struct {
	Type uint16
	Hash [32]byte
}
type InventoryMessage struct {
	Header       Header
	Type         uint16
	Version      byte
	LenInventory vlqInt
	Inventory    []InventoryItem
}
type GetDataMessage struct {
	Header   Header
	Type     uint16
	Version  byte
	DataType uint16
	Hash     [32]byte
}
type DataMessage struct {
	Header   Header
	Type     uint16
	Version  byte
	DataType uint16
}
type DataBlockMessage struct {
	Header   Header
	Type     uint16
	Version  byte
	DataType uint16
	Block    Block
}
type DataTransactionMessage struct {
	Header      Header
	Type        uint16
	Version     byte
	DataType    uint16
	Transaction Transaction
}
type GetPeersMessage struct {
	Header
	Type    uint16
	Version byte
}
type PeerMessage struct { // 22 len
	LastSeen uint32
	IP       [16]byte
	Port     uint16
}
type PeersMessage struct {
	Header   Header
	Type     uint16
	Version  byte
	LenPeers vlqInt
	Peers    []PeerMessage
}
