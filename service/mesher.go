package service

import (
	"github.com/golang/protobuf/proto"
	"github.com/mafzst/sounddrop/message"
	"github.com/mafzst/sounddrop/util"
	"github.com/sirupsen/logrus"
)

// Device mesh device
type Device struct {
	id      string
	online  bool
	allowed bool
}

// Mesher mesher service
type Mesher struct {
	message   chan proto.Message
	Messenger *Messenger
	log       *logrus.Entry
	devices   map[string]*Device
	sb        *util.ServiceBag
}

// Stop clean service when stopped by supervisor
func (msh *Mesher) Stop() {
	msh.log.Info("Mesher stopped.")
}

// Serve main service code
func (msh *Mesher) Serve() {
	msh.log = util.GetContextLogger("service/mesher.go", "Services/Mesher")

	msh.log.Info("Mesher starting...")

	msh.sb = util.GetServiceBag()
	msh.message = make(chan proto.Message)
	msh.devices = make(map[string]*Device)
	msh.Messenger.Register(message.PeerOnlineMessage, msh)
	msh.Messenger.Register(message.PeerOfflineMessage, msh)
	msh.Messenger.Register(message.DeviceAllowedMessage, msh)

	msh.devices[msh.sb.DeviceID.String()] = &Device{id: msh.sb.DeviceID.String(), online: true, allowed: msh.sb.Config.Mesh.AutoAccept}

	for {
		select {
		case msg := <-msh.message:
			switch m := msg.(type) {
			case *message.PeerOnline:
				if device, found := msh.devices[m.Id]; found {
					device.online = true
					msh.devices[m.Id] = device
					msh.log.Warn("Online device ", m.Id)
				} else {
					device := &Device{id: m.Id, online: true, allowed: msh.sb.Config.Mesh.AutoAccept}
					msh.devices[m.Id] = device
					msh.log.Debug("New device ", device.id)

					if msh.sb.Config.Mesh.AutoAccept {
						msh.log.Warn("Auto accepting device ", m.Id)
						msh.sendMeshState()
					}
				}
			case *message.PeerOffline:
				if device, found := msh.devices[m.Id]; found {
					device.online = false
					msh.devices[m.Id] = device
					msh.log.Warn("Offline device ", m.Id)
				}
			case *message.DeviceAllowed:
				if device, found := msh.devices[m.Id]; found && device.allowed == false {
					device.allowed = true
					msh.devices[m.Id] = device
					msh.log.Warn("Accepted device ", m.Id)
				}
			case *message.DeviceDisallowed:
				if device, found := msh.devices[m.Id]; found && device.allowed == true {
					device.allowed = false
					msh.devices[m.Id] = device
					msh.log.Warn("Evicted device ", m.Id)
				}
			}
		}
	}
}

// GetChan returns messaging chan
func (msh *Mesher) GetChan() chan proto.Message {
	return msh.message
}

func (msh *Mesher) sendMeshState() {
	for _, device := range msh.devices {
		var notification proto.Message

		if device.allowed {
			notification = &message.DeviceAllowed{Id: device.id}
		} else {
			notification = &message.DeviceDisallowed{Id: device.id}
		}

		notificationData, _ := message.ToBuffer(notification)
		msh.Messenger.Message <- &message.WriteRequest{DeviceName: "*", Message: notificationData}
	}
}
