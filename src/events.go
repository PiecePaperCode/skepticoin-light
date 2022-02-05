package main

import (
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
			helloMessage.Header.Len = uint32(len(SERIALIZE(helloMessage))) - 8
			con, err := connect(peer)
			if err != nil {
				peers[i].Connected = false
				continue
			}
			checkErrorReturn(con.Write(SERIALIZE(helloMessage)))
			peers[i].Socket = con
			peers[i].Connected = true
		}
	},
	GetPeers: func(peer RemotePeer) {
		logger.Info("Get Peers from", ip(peer.IP, peer.Port))
		getPeersMessage :=
			GetPeersMessage{
				Header:  constructHeader(),
				Type:    MSG.GetPeers,
				Version: 0,
			}
		getPeersMessage.Header.Len = uint32(
			len(SERIALIZE(getPeersMessage)),
		) - 8
		getPeersMessage.Header.Id = 2
		getPeersMessage.Header.ResponseId = 2
		getPeersMessage.Header.Context = 124312381912
		checkErrorReturn(peer.Socket.Write(SERIALIZE(getPeersMessage)))
	},
	SendPeers: func(peer RemotePeer) {
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
			len(SERIALIZE(peersMessage)),
		) - 8
		checkErrorReturn(peer.Socket.Write(SERIALIZE(peersMessage)))
		logger.Info(
			ip(peer.IP, peer.Port),
			"Send", peersMessage.LenPeers,
			"Peers",
		)
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
		300*time.Millisecond,
	)
	if err != nil {
		return nil, err
	}
	return con, err
}
