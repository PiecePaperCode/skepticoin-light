package main

import (
	"bufio"
	"encoding/binary"
	"github.com/wonderivan/logger"
	"io"
	"math/rand"
	"net"
	"time"
)

type RemotePeer struct {
	Socket    net.Conn
	IP        [16]byte
	Port      uint16
	Connected bool
}

var peers = []RemotePeer{
	{
		IP:        [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 23, 88, 99, 67},
		Port:      2412,
		Connected: false,
	},
}

func receiveLoop(i int) {
	for {
		err := receive(peers[i])
		if err != nil {
			logger.Warn(err)
			peers[i].Connected = false
			return
		}
	}
}
func receive(peer RemotePeer) error {
	bufioReader := bufio.NewReader(peer.Socket)
	maji, err := bufioReader.Peek(4)
	if err != nil {
		logger.Error("Connection got Dropped", ip(peer.IP, peer.Port))
		return err
	}
	if string(maji) != "MAJI" {
		logger.Warn("WRONG MAJI", maji)
		for {
			checkErrorReturn(bufioReader.Discard(1))
			maji, err = bufioReader.Peek(4)
			checkError(err)
			if string(maji) == "MAJI" {
				break
			}
		}
		return nil
	}
	sizeByte, err := bufioReader.Peek(8)
	checkError(err)
	size := int(binary.BigEndian.Uint32(sizeByte[4:8])) + 8
	reply := make([]byte, size)
	checkErrorReturn(io.ReadFull(bufioReader, reply))
	message, _ := deserialize(reply, Message{})
	switch message.(Message).Type {
	case MSG.Hello:
		receiveHello(reply)
		requestPeers(peer)
		//requestBlocks(peer)
		requestBlockData(peer)
		break
	case MSG.GetPeers:
		sendPeers(peer)
		requestPeers(peer)
		break
	case MSG.Inventory:
		receiveInventory(reply)
	case MSG.Peers:
		receivePeers(reply)
		break
	case MSG.Data:
		receiveData(reply)
		break
	}
	logger.Info(
		"Received", size,
		"Message Type", message.(Message).Type,
		"From", ip(peer.IP, peer.Port),
	)
	return nil
}
func sendHello() {
	for i, peer := range peers {
		if peer.Connected {
			continue
		}
		helloMessage := HelloMessage{
			Header:           constructHeader(),
			Type:             MSG.Hello,
			Version:          0,
			IP:               peer.IP,
			Port:             peer.Port,
			MyIp:             [18]byte{},
			Nonce:            rand.Uint32(),
			LenUserAgent:     14,
			UserAgent:        [14]byte{'p', 'i', 'e', 'c', 'e', 'g', 'o', ' ', '0', '.', '1', '.', '2', '1'},
			SupportedVersion: 0,
			Zeros:            [256]byte{},
		}
		helloMessage.Header.Len = uint32(len(serialize(helloMessage))) - 8
		con, err := connect(peer)
		if err != nil {
			peers[i].Connected = false
			continue
		}
		checkErrorReturn(con.Write(serialize(helloMessage)))
		peers[i].Socket = con
		peers[i].Connected = true
		go receiveLoop(i)
	}
}
func receiveHello(b []byte) {
	message, _ := deserialize(b, HelloMessage{})
	helloMessage := message.(HelloMessage)
	logger.Info(
		ip(helloMessage.IP, helloMessage.Port),
		"says Hello",
		string(helloMessage.UserAgent[:]),
	)
}
func requestBlocks(peer RemotePeer) {
	getBlocksMessage := GetBlocksMessage{
		Header:       constructHeader(),
		Type:         MSG.GetBlocks,
		Version:      0,
		LenStartHash: 1,
		StartHashes: [32]byte{0, 4, 130, 131, 177, 219, 116, 43, 192, 108, 216,
			194, 193, 51, 62, 126, 122, 127, 18, 74, 26, 233, 20, 147, 205, 167,
			223, 222, 146, 196, 60, 181},
		Zeros: [32]byte{},
	}
	getBlocksMessage.Header.Len = uint32(
		len(serialize(getBlocksMessage)),
	) - 8
	checkErrorReturn(peer.Socket.Write(serialize(getBlocksMessage)))
	logger.Debug("Get Blocks")
}
func requestBlockData(peer RemotePeer) {
	getDataMessage := GetDataMessage{
		Header:   constructHeader(),
		Type:     MSG.GetBlocks,
		Version:  0,
		DataType: DATA.Block,
		Hash: [32]byte{98, 32, 183, 18, 107, 163, 66, 204, 151, 224, 229, 157,
			13, 19, 250, 232, 100, 217, 25, 222, 9, 89, 107, 133, 232, 53, 138,
			254, 184, 81, 97, 166},
	}
	getDataMessage.Header.Len = uint32(
		len(serialize(getDataMessage)),
	) - 8
	checkErrorReturn(peer.Socket.Write(serialize(getDataMessage)))
	logger.Debug("Get DATA Blocks")
}
func receiveInventory(b []byte) {
	message, _ := deserialize(b, InventoryMessage{})
	inventoryMessage := message.(InventoryMessage)
	logger.Debug(
		"Inventory Type",
		inventoryMessage.Inventory[0].Type,
		"Message",
		inventoryMessage.Inventory[0].Hash,
	)
}
func requestPeers(peer RemotePeer) {
	getPeersMessage :=
		GetPeersMessage{
			Header:  constructHeader(),
			Type:    MSG.GetPeers,
			Version: 0,
		}
	getPeersMessage.Header.Len = uint32(
		len(serialize(getPeersMessage)),
	) - 8
	getPeersMessage.Header.Id = 2
	getPeersMessage.Header.ResponseId = 2
	getPeersMessage.Header.Context = 124312381912
	checkErrorReturn(peer.Socket.Write(serialize(getPeersMessage)))
	logger.Info("Get Peers from", ip(peer.IP, peer.Port))
}
func sendPeers(peer RemotePeer) {
	var peerMessage []PeerMessage
	for _, p := range peers {
		if p.Connected {
			peerMessage = append(peerMessage, PeerMessage{
				LastSeen: uint32(time.Now().Unix()),
				IP:       p.IP,
				Port:     p.Port,
			})
		}
	}
	peersMessage := PeersMessage{
		Header:   constructHeader(),
		Type:     MSG.Peers,
		Version:  0,
		LenPeers: vlqInt(len(peerMessage)),
		Peers:    peerMessage,
	}
	peersMessage.Header.Len = uint32(
		len(serialize(peersMessage)),
	) - 8
	checkErrorReturn(peer.Socket.Write(serialize(peersMessage)))
	logger.Info(
		ip(peer.IP, peer.Port),
		"Send", peersMessage.LenPeers,
		"Peers",
	)
}
func receivePeers(b []byte) {
	message, _ := deserialize(b, PeersMessage{})
	peersMessage := message.(PeersMessage)
	for _, peerMessage := range peersMessage.Peers {
		known := false
		for _, p := range peers {
			if string(p.IP[:]) == string(peerMessage.IP[:]) {
				known = true
				break
			}
		}
		lastSeen := time.Now().Sub(
			time.Unix(int64(peerMessage.LastSeen), 0),
		)
		if !known && lastSeen.Minutes() < 5 {
			peers = append(
				peers,
				RemotePeer{
					IP:        peerMessage.IP,
					Port:      peerMessage.Port,
					Connected: false,
				},
			)
		}
	}
	logger.Info("Received", peersMessage.LenPeers, "Peers")
}
func receiveData(b []byte) {
	message, _ := deserialize(b, DataBlockMessage{})
	dataBlockMessage := message.(DataBlockMessage)
	if dataBlockMessage.DataType == DATA.Block {
		logger.Debug(
			"New Block at Height",
			dataBlockMessage.Block.Header.Height,
			dataBlockMessage.Block.Header.BlockHash,
		)
	}
}
func connect(peer RemotePeer) (net.Conn, error) {
	interfaceIp := make([]interface{}, 16)
	for i, b := range peer.IP {
		interfaceIp[i] = b
	}
	con, err := net.DialTimeout(
		"tcp",
		ip(peer.IP, peer.Port),
		300*time.Millisecond,
	)
	if err != nil {
		return nil, err
	}
	return con, err
}
