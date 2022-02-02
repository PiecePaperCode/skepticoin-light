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
			logger.Warn(ip(peer.IP, peer.Port), "Got Dropped")
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
		message := deSerialize.Message(reply)
		logger.Info(
			"Received", size,
			"Message Type", message.Type,
			"From", ip(peer.IP, peer.Port),
		)
		switch message.Type {
		case MSG.Hello:
			helloMessage := deSerialize.Hello(reply)
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
			peersMessage, err := deSerialize.Peers(reply)
			if err != nil {
				peers[i].Connected = false
				logger.Error(ip(peer.IP, peer.Port), "Got Dropped")
				go PeerEvent.Hello()
			}
			logger.Debug("Received", peersMessage.LenPeers, "Peers")
			PeerEvent.AddPeers(peersMessage)
			go PeerEvent.Hello()
			break
		case MSG.Data:
			dataBlockMessage := deSerialize.DataBlockMessage(reply)
			if dataBlockMessage.DataType == DATA.Block {
				logger.Debug(
					"New Block at Height",
					dataBlockMessage.Block.Header.Height,
				)
			}
			break
		}
	}
}
