package service

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/mafzst/sounddrop/message"
	"github.com/mafzst/sounddrop/util"
	"sync"
)

type Messenger struct {
	Message   chan proto.Message
	log       *logrus.Entry
	listeners map[byte][]message.Receiver
	listenersMutex sync.Mutex
}

func (this *Messenger) Stop() {
	this.log.Info("Messenger stopped.")
}

func (this *Messenger) Serve() {
	this.listeners = make(map[byte][]message.Receiver)

	this.log = util.GetContextLogger("service/messenger.go", "Services/Messenger")
	this.log.Info("Messenger starting...")

	for {
		select {
		case msg := <-this.Message:
			opCode, err := message.FindOpCode(msg)
			if err != nil {
				this.log.Warn(fmt.Sprintf("Unable to find opCode for message %#v", msg))
			}

			if listeners, ok := this.listeners[opCode]; ok {
				for _, listener := range listeners {
					listener.GetChan() <- msg
				}
			} else {
				this.log.Debug(fmt.Sprintf("Message of type %d received but no listeners was registered", opCode))
			}
		}
	}
}

func (this *Messenger) Register(opCode byte, listener message.Receiver) {
	this.listenersMutex.Lock()
	this.listeners[opCode] = append(this.listeners[opCode], listener)
	this.listenersMutex.Unlock()
}
