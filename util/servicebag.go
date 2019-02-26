package util

import (
	"github.com/syncthing/syncthing/lib/protocol"
)

type ServiceBag struct {
	DeviceID protocol.DeviceID
	Config   *Config
}

var instance *ServiceBag

func GetServiceBag() *ServiceBag {
	if instance == nil {
		config := InitConfig()
		instance = &ServiceBag{Config: config}
	}

	return instance
}
