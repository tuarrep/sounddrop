package main

import (
	"github.com/golang/protobuf/proto"
	"github.com/thejerf/suture"
	"os"
	"os/signal"
	"github.com/mafzst/sounddrop/service"
	"github.com/mafzst/sounddrop/util"
)

func main() {
	util.InitLogger()

	log := util.GetContextLogger("main.go", "main")

	log.Info("Starting main process...")
	myID, err := util.GetMyId()
	if err != nil {
		log.Fatal("Cannot obtain device ID. Aborting startup.", err)
		os.Exit(-1)
	}

	log.Info("I'm known on mesh by: ", myID.String())

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	sb := util.GetServiceBag()
	sb.DeviceID = myID

	supervisor := suture.NewSimple("supervisor")

	messenger := &service.Messenger{Message: make(chan proto.Message)}
	supervisor.Add(messenger)

	server := &service.Server{Messenger: messenger}
	supervisor.Add(server)

	mesher := &service.Mesher{Messenger: messenger}
	supervisor.Add(mesher)

	player := &service.Player{Messenger:messenger}
	supervisor.Add(player)

	if sb.Config.Streamer.AutoStart {
		streamer := &service.Streamer{Messenger: messenger}
		supervisor.Add(streamer)
	}

	supervisor.ServeBackground()

	log.Info("Main process started.")

	for {
		select {
		case <-stop:
			supervisor.Stop()
			log.Info("Clean exit. \n")
			os.Exit(0)
		}
	}
}
