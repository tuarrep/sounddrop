package service

import (
	"fmt"
	"github.com/faiface/beep/wav"
	"github.com/golang/protobuf/proto"
	"github.com/mafzst/sounddrop/message"
	"github.com/mafzst/sounddrop/util"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"time"
)

// Streamer audio streamer service
type Streamer struct {
	Message   chan proto.Message
	log       *logrus.Entry
	Messenger *Messenger
	sb        *util.ServiceBag
}

// Stop clean service when stopped by supervisor
func (s *Streamer) Stop() {
	s.log.Info("Streamer stopped.")
}

// Serve main service code
func (s *Streamer) Serve() {
	s.log = util.GetContextLogger("service/streamer.go", "Services/Streamer")
	s.log.Info("Streamer starting...")

	s.sb = util.GetServiceBag()

	files, err := ioutil.ReadDir(s.sb.Config.Streamer.PlaylistDir)
	util.CheckError(err, s.log)

	s.log.Info(fmt.Sprintf("Streamer started. %d files found", len(files)))

	for _, file := range files {
		fileData, err := os.Open(fmt.Sprintf("%s/%s", s.sb.Config.Streamer.PlaylistDir, file.Name()))
		util.CheckError(err, s.log)
		stream, format, err := wav.Decode(fileData)
		util.CheckError(err, s.log)

		s.log.Info(fmt.Sprintf("Audio format is: channels=%d, sampleRate=%d, precision=%d", format.NumChannels, format.SampleRate, format.Precision))

		buff := make([][2]float64, 512)

		ok := true
		n := 512

		for ok == true {
			n, ok = stream.Stream(buff)

			samplesLeft := make([]float64, n)
			samplesRight := make([]float64, n)

			for i := 0; i < n; i++ {
				samplesLeft[i] = buff[i][0]
				samplesRight[i] = buff[i][1]
			}

			msg := &message.StreamData{SamplesLeft: samplesLeft, SamplesRight: samplesRight}
			msgData, _ := message.ToBuffer(msg)
			s.Messenger.Message <- &message.WriteRequest{DeviceName: "*", Message: msgData}

			time.Sleep(format.SampleRate.D(n))
		}
	}
}
