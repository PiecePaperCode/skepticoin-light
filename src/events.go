package main

import (
	"bytes"
	"encoding/binary"
	"github.com/wonderivan/logger"
	"math/rand"
	"net"
	"time"
)

var PeerEvent = struct {
	Hello     func()
	GetPeers  func(peer RemotePeer)
	SendPeers func(peer RemotePeer)
	AddPeers  func(peersMessage PeersMessage)
}{
	Hello: func() {
		for i, peer := range peers {
			if peer.Connected {
				continue
			}
			hello := new(bytes.Buffer)
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
			helloMessage.Header.Len = uint32(binary.Size(helloMessage)) - 8
			checkError(
				binary.Write(
					hello,
					binary.BigEndian,
					helloMessage,
				),
			)
			con, err := connect(peer)
			if err != nil {
				peers[i].Connected = false
				continue
			}
			checkErrorReturn(con.Write(hello.Bytes()))
			peers[i].Socket = con
			peers[i].Connected = true
		}
	},
	GetPeers: func(peer RemotePeer) {
		logger.Info("Get Peers from", ip(peer.IP, peer.Port))
		buf := new(bytes.Buffer)
		getPeersMessage :=
			GetPeersMessage{
				Header:  constructHeader(),
				Type:    MSG.GetPeers,
				Version: 0,
			}
		getPeersMessage.Header.Len = uint32(
			binary.Size(getPeersMessage),
		) - 8
		getPeersMessage.Header.Id = 2
		getPeersMessage.Header.ResponseId = 1
		getPeersMessage.Header.Context = 14939624176500638228
		checkError(binary.Write(buf, binary.BigEndian, getPeersMessage))
		checkErrorReturn(peer.Socket.Write(buf.Bytes()))
	},
	SendPeers: func(peer RemotePeer) {
		var peerMessage [10]PeerMessage
		for i, p := range peers {
			if 10 < i {
				break
			}
			if p.Connected {
				peerMessage[i] = PeerMessage{
					LastSeen: uint32(time.Now().Unix()),
					IP:       [16]byte{},
					Port:     0,
				}
			}
		}
		peersMessage := PeersMessage{
			Header:   constructHeader(),
			Type:     MSG.Peers,
			Version:  0,
			LenPeers: 10,
			Peers:    peerMessage,
		}
		peersMessage.Header.Len = uint32(
			binary.Size(peersMessage),
		) - 8
		buf := new(bytes.Buffer)
		checkError(binary.Write(buf, binary.BigEndian, peersMessage))
		checkErrorReturn(peer.Socket.Write(buf.Bytes()))
		logger.Info(ip(peer.IP, peer.Port), "Received", len(buf.Bytes()), "Peers")
	},
	AddPeers: func(peersMessage PeersMessage) {
		for _, peerMessage := range peersMessage.Peers {
			known := false
			for _, p := range peers {
				if string(p.IP[:]) == string(peerMessage.IP[:]) {
					known = true
					break
				}
			}
			if !known {
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
	},
}

func connect(peer RemotePeer) (net.Conn, error) {
	interfaceIp := make([]interface{}, 16)
	for i, b := range peer.IP {
		interfaceIp[i] = b
	}
	con, err := net.DialTimeout(
		"tcp",
		ip(peer.IP, peer.Port),
		2*time.Second,
	)
	if err != nil {
		return nil, err
	}
	return con, err
}
