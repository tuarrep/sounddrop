package util

import (
	"github.com/syncthing/syncthing/lib/protocol"
)

// ServiceBag various global object injected to services
type ServiceBag struct {
	DeviceID protocol.DeviceID
	Config   *Config
}

var instance *ServiceBag

// GetServiceBag get servicebag singleton
func GetServiceBag() *ServiceBag {
	if instance == nil {
		config := InitConfig()
		instance = &ServiceBag{Config: config}
	}

	return instance
}
