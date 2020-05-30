package service

import (
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/tuarrep/sounddrop/message"
	"github.com/tuarrep/sounddrop/util"
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

	msh.Messenger.RegisterSome([]byte{message.PeerOnlineMessage, message.PeerOfflineMessage, message.DeviceStatusMessage}, msh)

	msh.devices[msh.sb.DeviceID.String()] = &Device{id: msh.sb.DeviceID.String(), online: true, allowed: msh.sb.Config.Mesh.AutoAccept}

	for {
		select {
		case msg := <-msh.message:
			switch m := msg.(type) {
			case *message.PeerOnline:
				msh.handlePeerOnline(m)
			case *message.PeerOffline:
				msh.handlePeerOffline(m)
			case *message.DeviceStatus:
				msh.handleDeviceStatus(m)
			}
		}
	}
}

func (msh *Mesher) handleDeviceStatus(m *message.DeviceStatus) {
	if device, found := msh.devices[m.Id]; found && device.allowed == false {
		device.allowed = m.Allowed
		msh.devices[m.Id] = device
		msh.log.Warn("Accepted device ", m.Id)
	}
}

func (msh *Mesher) handlePeerOffline(m *message.PeerOffline) {
	if device, found := msh.devices[m.Id]; found {
		device.online = false
		msh.devices[m.Id] = device
		msh.log.Warn("Offline device ", m.Id)
	}
}

func (msh *Mesher) handlePeerOnline(m *message.PeerOnline) {
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
}

// GetChan returns messaging chan
func (msh *Mesher) GetChan() chan proto.Message {
	return msh.message
}

func (msh *Mesher) sendMeshState() {
	for _, device := range msh.devices {
		notification := &message.DeviceStatus{Id: device.id, Allowed: device.allowed}
		notificationData, _ := message.ToBuffer(notification)
		msh.Messenger.Message <- &message.WriteRequest{DeviceName: "*", Message: notificationData}
	}
}
