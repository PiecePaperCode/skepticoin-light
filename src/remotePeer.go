package main

import (
	"bufio"
	"encoding/binary"
	"github.com/wonderivan/logger"
	"io"
	"net"
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

func receive() {
	for i, peer := range peers {
		if peer.Connected != true {
			continue
		}
		bufioReader := bufio.NewReader(peer.Socket)
		maji, err := bufioReader.Peek(4)
		if err != nil {
			peers[i].Connected = false
			logger.Warn(ip(peer.IP, peer.Port), "Got Dropped", err)
			go PeerEvent.Hello()
			continue
		}
		if string(maji) != "MAJI" || err != nil {
			logger.Error("Wrong MAJI", maji)
			checkErrorReturn(bufioReader.Discard(4))
			continue
		}
		sizeByte, err := bufioReader.Peek(8)
		checkError(err)
		size := int(binary.BigEndian.Uint32(sizeByte[4:8])) + 8
		reply := make([]byte, size)
		checkErrorReturn(io.ReadFull(bufioReader, reply))
		message, _ := DESERIALIZE(reply, Message{})
		logger.Info(
			"Received", size,
			"Message Type", message.(Message).Type,
			"From", ip(peer.IP, peer.Port),
		)
		switch message.(Message).Type {
		case MSG.Hello:
			message, _ := DESERIALIZE(reply, HelloMessage{})
			helloMessage := message.(HelloMessage)
			go PeerEvent.GetPeers(peer)
			logger.Info(
				ip(helloMessage.IP, helloMessage.Port),
				"says Hello",
				string(helloMessage.UserAgent[:]),
			)
			break
		case MSG.GetPeers:
			go PeerEvent.SendPeers(peer)
			break
		case MSG.Peers:
			message, _ := DESERIALIZE(reply, PeersMessage{})
			peersMessage := message.(PeersMessage)
			logger.Debug("Received", peersMessage.LenPeers, "Peers")
			PeerEvent.AddPeers(peersMessage)
			go PeerEvent.Hello()
			break
		case MSG.Data:
			message, _ := DESERIALIZE(reply, DataBlockMessage{})
			dataBlockMessage := message.(DataBlockMessage)
			if dataBlockMessage.DataType == DATA.Block {
				logger.Debug(
					"New Block at Height",
					dataBlockMessage.Block.Header.Height,
					dataBlockMessage.Block.Transactions[0].Outputs[0].Value,
				)
			}
			break
		}
	}
}
