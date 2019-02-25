package service

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/xtaci/kcp-go"
	"net"
	"sounddrop/message"
	"sounddrop/util"
	"time"
)

type Peer struct {
	id       string
	us       *kcp.UDPSession
	lastSeen time.Time
}

type Server struct {
	message   chan proto.Message
	ticker    chan bool
	Messenger *Messenger
	log       *logrus.Entry
	sc        *net.UDPConn
	sb        *util.ServiceBag
	listener  *kcp.Listener
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

	//ServerAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%v", this.sb.Config.Discover.Port))
	//util.CheckError(err, this.log)
	//this.sc, err = net.ListenUDP("udp", ServerAddr)
	//util.CheckError(err, this.log)

	////

	this.log.Debug("Before all KCP")

	var err error

	this.listener, err = kcp.ListenWithOptions(fmt.Sprintf(":%v", this.sb.Config.Discover.Port), nil, 10, 3)
	util.CheckError(err, this.log)
	//listener.SetDSCP(0)
	err = this.listener.SetReadBuffer(4194304)
	util.CheckError(err, this.log)
	err = this.listener.SetWriteBuffer(4194304)
	util.CheckError(err, this.log)
	//listener.SetDeadline(time.Now().Add(time.Second))
	this.log.Debug("After ServeCon")
	util.CheckError(err, this.log)

	this.log.Debug("After all KCP")

	////

	this.message = make(chan proto.Message)
	this.ticker = make(chan bool)
	this.peers = make(map[string]*Peer)

	go this.listenerLoop()
	go this.tick()

	this.Messenger.Register(message.WriteRequestMessage, this)

	this.log.Info("Server started. Port: ", this.sb.Config.Discover.Port)

	for {
		select {
		case _ = <-this.ticker:
			this.sendAnnounce()
			this.checkPeersHealth()
		case msg := <-this.message:
			switch m := msg.(type) {
			case *message.WriteRequest:
				var sessions []*kcp.UDPSession

				if m.DeviceName == "*" {
					for _, peer := range this.peers {
						sessions = append(sessions, peer.us)
					}
				} else if peer, exists := this.peers[m.DeviceName]; exists {
					sessions = append(sessions, peer.us)
				}

				for _, us := range sessions {
					//us, _ := kcp.DialWithOptions(us, nil, 10, 3)
					_, err := us.Write(m.Message)
					//_, err := this.sc.WriteToUDP(m.Message, us)

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
		if us, err := this.listener.AcceptKCP(); err == nil {
			n, err := us.Read(buf)
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
					this.peers[m.DeviceName].us = us
					this.peers[m.DeviceName].lastSeen = time.Now()
				} else {
					this.log.Debug("New device discovered: ", m.DeviceName)
					this.peers[m.DeviceName] = &Peer{id: m.DeviceName, us: us, lastSeen: time.Now()}

					notification := &message.PeerOnline{Id: m.DeviceName}
					this.Messenger.Message <- notification
				}
			default:
				this.Messenger.Message <- msg
			}
		} else {
			this.log.Warn(fmt.Sprintf("Error establishing peer connection: %+v", err))
		}
	}
}

func (this *Server) sendAnnounce() {
	announce := &message.Announce{ServiceNumber: message.ServiceNumber, DeviceName: this.sb.DeviceID.String(), Raddr: this.listener.Addr().String()}
	data, err := message.ToBuffer(announce)
	util.CheckError(err, this.log)
	us, err := kcp.DialWithOptions(fmt.Sprintf("255.255.255.255:%d", this.sb.Config.Discover.Port), nil, 10, 3)
	util.CheckError(err, this.log)
	_, err = us.Write(data)
	//_, err = this.sc.WriteToUDP(data, &net.UDPAddr{IP: net.IP{255, 255, 255, 255}, Port: this.sb.Config.Discover.Port})
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
