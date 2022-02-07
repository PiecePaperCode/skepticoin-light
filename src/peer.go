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
		checkErrorReturn(bufioReader.Discard(4))
		logger.Error("Wrong MAJI", maji)
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
		go receiveHello(reply)
		go requestPeers(peer)
		break
	case MSG.GetPeers:
		go sendPeers(peer)
		go requestPeers(peer)
		break
	case MSG.Peers:
		go receivePeers(reply)
		break
	case MSG.Data:
		go receiveData(reply)
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
			UserAgentLength:  14,
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
	logger.Debug("Received", peersMessage.LenPeers, "Peers")
}
func receiveData(b []byte) {
	message, _ := deserialize(b, DataBlockMessage{})
	dataBlockMessage := message.(DataBlockMessage)
	if dataBlockMessage.DataType == DATA.Block {
		logger.Debug(
			"New Block at Height",
			dataBlockMessage.Block.Header.Height,
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
