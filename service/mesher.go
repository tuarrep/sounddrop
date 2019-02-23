package service

import (
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"sounddrop/message"
	"sounddrop/util"
)

type Device struct {
	id      string
	online  bool
	allowed bool
}

type Mesher struct {
	message   chan proto.Message
	Messenger *Messenger
	log       *logrus.Entry
	devices   map[string]*Device
	sb        *util.ServiceBag
}

func (this *Mesher) Stop() {
	this.log.Info("Mesher stopped.")
}

func (this *Mesher) Serve() {
	this.log = util.GetContextLogger("service/mesher.go", "Services/Mesher")

	this.log.Info("Mesher starting...")

	this.sb = util.GetServiceBag()
	this.message = make(chan proto.Message)
	this.devices = make(map[string]*Device)
	this.Messenger.Register(message.PeerOnlineMessage, this)
	this.Messenger.Register(message.PeerOfflineMessage, this)
	this.Messenger.Register(message.DeviceAllowedMessage, this)

	this.devices[this.sb.DeviceID.String()] = &Device{id: this.sb.DeviceID.String(), online: true, allowed: this.sb.Config.Mesh.AutoAccept}

	for {
		select {
		case msg := <-this.message:
			switch m := msg.(type) {
			case *message.PeerOnline:
				if device, found := this.devices[m.Id]; found {
					device.online = true
					this.devices[m.Id] = device
					this.log.Warn("Online device ", m.Id)
				} else {
					device := &Device{id: m.Id, online: true, allowed: this.sb.Config.Mesh.AutoAccept}
					this.devices[m.Id] = device
					this.log.Debug("New device ", device.id)

					if this.sb.Config.Mesh.AutoAccept {
						this.log.Warn("Auto accepting device ", m.Id)
						this.sendMeshState()
					}
				}
			case *message.PeerOffline:
				if device, found := this.devices[m.Id]; found {
					device.online = false
					this.devices[m.Id] = device
					this.log.Warn("Offline device ", m.Id)
				}
			case *message.DeviceAllowed:
				if device, found := this.devices[m.Id]; found && device.allowed == false {
					device.allowed = true
					this.devices[m.Id] = device
					this.log.Warn("Accepted device ", m.Id)
				}
			case *message.DeviceDisallowed:
				if device, found := this.devices[m.Id]; found && device.allowed == true {
					device.allowed = false
					this.devices[m.Id] = device
					this.log.Warn("Evicted device ", m.Id)
				}
			}
		}
	}
}

func (this *Mesher) GetChan() chan proto.Message {
	return this.message
}

func (this *Mesher) sendMeshState() {
	for _, device := range this.devices {
		var notification proto.Message

		if device.allowed {
			notification = &message.DeviceAllowed{Id: device.id}
		} else {
			notification = &message.DeviceDisallowed{Id: device.id}
		}

		notificationData, _ := message.ToBuffer(notification)
		this.Messenger.Message <- &message.WriteRequest{DeviceName: "*", Message: notificationData}
	}
}
