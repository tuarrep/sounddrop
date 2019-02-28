package service

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/tuarrep/sounddrop/message"
	"github.com/tuarrep/sounddrop/util"
	"net"
	"time"
)

//Peer structure represents discovered peer on mesh network
type Peer struct {
	id       string
	address  *net.UDPAddr
	lastSeen time.Time
}

// Server UDP server service
type Server struct {
	message   chan proto.Message
	ticker    chan bool
	Messenger *Messenger
	log       *logrus.Entry
	sc        *net.UDPConn
	sb        *util.ServiceBag
	peers     map[string]*Peer
}

var tickInterval = 1 * time.Second

// Stop clean service when stopped by supervisor
func (srv *Server) Stop() {
	srv.sc.Close()
	srv.log.Info("Server stopped.")
}

// Serve main service code
func (srv *Server) Serve() {
	srv.log = util.GetContextLogger("service/server.go", "Services/Server")

	srv.log.Info("Server starting...")
	srv.sb = util.GetServiceBag()

	ServerAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%v", srv.sb.Config.Discover.Port))
	util.CheckError(err, srv.log)
	srv.sc, err = net.ListenUDP("udp", ServerAddr)
	util.CheckError(err, srv.log)

	srv.message = make(chan proto.Message)
	srv.ticker = make(chan bool)
	srv.peers = make(map[string]*Peer)

	go srv.listenerLoop()
	go srv.tick()

	srv.Messenger.Register(message.WriteRequestMessage, srv)

	srv.log.Info("Server started. Listening at", srv.sc.LocalAddr().String())

	for {
		select {
		case _ = <-srv.ticker:
			srv.sendAnnounce()
			srv.checkPeersHealth()
		case msg := <-srv.message:
			switch m := msg.(type) {
			case *message.WriteRequest:
				var addresses []*net.UDPAddr

				if m.DeviceName == "*" {
					for _, peer := range srv.peers {
						addresses = append(addresses, peer.address)
					}
				} else if peer, exists := srv.peers[m.DeviceName]; exists {
					addresses = append(addresses, peer.address)
				}

				for _, address := range addresses {
					_, err := srv.sc.WriteToUDP(m.Message, address)

					if err != nil {
						srv.log.Warn(fmt.Sprintf("Failed to send message %d due to %v", m.Message[0], err.Error()))
					}
				}
			}
		}
	}
}

// GetChan returns messaging chan
func (srv *Server) GetChan() chan proto.Message {
	return srv.message
}

func (srv *Server) tick() {
	for {
		srv.ticker <- true
		time.Sleep(tickInterval)
	}
}

func (srv *Server) listenerLoop() {
	buf := make([]byte, 65526)

	for {
		n, addr, err := srv.sc.ReadFromUDP(buf)
		msg, err := message.FromBuffer(buf[0:n])
		util.CheckError(err, srv.log)
		switch m := msg.(type) {
		case *message.Announce:
			if m.DeviceName == srv.sb.DeviceID.String() {
				continue
			}

			if m.ServiceNumber != message.ServiceNumber {
				continue
			}

			if _, found := srv.peers[m.DeviceName]; found {
				srv.peers[m.DeviceName].address = addr
				srv.peers[m.DeviceName].lastSeen = time.Now()
			} else {
				srv.log.Debug("New device discovered: ", m.DeviceName)
				srv.peers[m.DeviceName] = &Peer{id: m.DeviceName, address: addr, lastSeen: time.Now()}

				notification := &message.PeerOnline{Id: m.DeviceName}
				srv.Messenger.Message <- notification
			}
		default:
			srv.Messenger.Message <- msg
		}
	}
}

func (srv *Server) sendAnnounce() {
	announce := &message.Announce{ServiceNumber: message.ServiceNumber, DeviceName: srv.sb.DeviceID.String()}
	data, err := message.ToBuffer(announce)
	util.CheckError(err, srv.log)
	_, err = srv.sc.WriteToUDP(data, &net.UDPAddr{IP: net.IP{255, 255, 255, 255}, Port: srv.sb.Config.Discover.Port})
	util.CheckError(err, srv.log)
}

func (srv *Server) checkPeersHealth() {
	for id, device := range srv.peers {
		if time.Now().After(device.lastSeen.Add(3 * tickInterval)) {
			srv.log.Warn(fmt.Sprintf("Device %s not announced since a while. Romoving it from known peers.", id))
			delete(srv.peers, id)
			notification := &message.PeerOffline{Id: id}
			srv.Messenger.Message <- notification
		}
	}
}
