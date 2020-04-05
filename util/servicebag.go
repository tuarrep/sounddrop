package util

import (
	"github.com/google/uuid"
)

// ServiceBag various global object injected to services
type ServiceBag struct {
	DeviceID uuid.UUID
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
