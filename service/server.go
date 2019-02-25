package service

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"net"
	"github.com/mafzst/sounddrop/message"
	"github.com/mafzst/sounddrop/util"
	"time"
)

type Peer struct {
	id       string
	address  *net.UDPAddr
	lastSeen time.Time
}

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

func (this *Server) Stop() {
	this.sc.Close()
	this.log.Info("Server stopped.")
}

func (this *Server) Serve() {
	this.log = util.GetContextLogger("service/server.go", "Services/Server")

	this.log.Info("Server starting...")
	this.sb = util.GetServiceBag()

	ServerAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%v", this.sb.Config.Discover.Port))
	util.CheckError(err, this.log)
	this.sc, err = net.ListenUDP("udp", ServerAddr)
	util.CheckError(err, this.log)

	this.message = make(chan proto.Message)
	this.ticker = make(chan bool)
	this.peers = make(map[string]*Peer)

	go this.listenerLoop()
	go this.tick()

	this.Messenger.Register(message.WriteRequestMessage, this)

	this.log.Info("Server started. Listening at", this.sc.LocalAddr().String())

	for {
		select {
		case _ = <-this.ticker:
			this.sendAnnounce()
			this.checkPeersHealth()
		case msg := <-this.message:
			switch m := msg.(type) {
			case *message.WriteRequest:
				var addresses []*net.UDPAddr

				if m.DeviceName == "*" {
					for _, peer := range this.peers {
						addresses = append(addresses, peer.address)
					}
				} else if peer, exists := this.peers[m.DeviceName]; exists {
					addresses = append(addresses, peer.address)
				}

				for _, address := range addresses {
					_, err := this.sc.WriteToUDP(m.Message, address)

					if err != nil {
						this.log.Warn(fmt.Sprintf("Failed to send message %d due to %v", m.Message[0], err.Error()))
					}
				}
			}
		}
	}
}

func (this *Server) GetChan() chan proto.Message {
	return this.message
}

func (this *Server) tick() {
	for {
		this.ticker <- true
		time.Sleep(tickInterval)
	}
}

func (this *Server) listenerLoop() {
	buf := make([]byte, 65526)

	for {
		n, addr, err := this.sc.ReadFromUDP(buf)
		msg, err := message.FromBuffer(buf[0:n])
		util.CheckError(err, this.log)
		switch m := msg.(type) {
		case *message.Announce:
			if m.DeviceName == this.sb.DeviceID.String() {
				continue
			}

			if m.ServiceNumber != message.ServiceNumber {
				continue
			}

			if _, found := this.peers[m.DeviceName]; found {
				this.peers[m.DeviceName].address = addr
				this.peers[m.DeviceName].lastSeen = time.Now()
			} else {
				this.log.Debug("New device discovered: ", m.DeviceName)
				this.peers[m.DeviceName] = &Peer{id: m.DeviceName, address: addr, lastSeen: time.Now()}

				notification := &message.PeerOnline{Id: m.DeviceName}
				this.Messenger.Message <- notification
			}
		default:
			this.Messenger.Message <- msg
		}
	}
}

func (this *Server) sendAnnounce() {
	announce := &message.Announce{ServiceNumber: message.ServiceNumber, DeviceName: this.sb.DeviceID.String()}
	data, err := message.ToBuffer(announce)
	util.CheckError(err, this.log)
	_, err = this.sc.WriteToUDP(data, &net.UDPAddr{IP: net.IP{255, 255, 255, 255}, Port: this.sb.Config.Discover.Port})
	util.CheckError(err, this.log)
}

func (this *Server) checkPeersHealth() {
	for id, device := range this.peers {
		if time.Now().After(device.lastSeen.Add(3 * tickInterval)) {
			this.log.Warn(fmt.Sprintf("Device %this not announced since a while. Romoving it from known peers.", id))
			delete(this.peers, id)
			notification := &message.PeerOffline{Id: id}
			this.Messenger.Message <- notification
		}
	}
}
