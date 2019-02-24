package service

import (
	"fmt"
	"github.com/faiface/beep/wav"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"os"
	"sounddrop/message"
	"sounddrop/util"
	"time"
)

type Streamer struct {
	Message   chan proto.Message
	log       *logrus.Entry
	Messenger *Messenger
}

func (this *Streamer) Stop() {
	this.log.Info("Streamer stopped.")
}

func (this *Streamer) Serve() {
	this.log = util.GetContextLogger("service/streamer.go", "Services/Streamer")
	this.log.Info("Streamer starting...")

	file, err := os.Open("test.wav")
	util.CheckError(err, this.log)
	stream, format, err := wav.Decode(file)
	util.CheckError(err, this.log)

	this.log.Info(fmt.Sprintf("Audio format is: channels=%d, sampleRate=%d, precision=%d", format.NumChannels, format.SampleRate, format.Precision))

	buff := make([][2]float64, 512)

	ok := true
	n := 512

	for ok == true {
		n, ok = stream.Stream(buff)

		samplesLeft := make([]float64, n)
		samplesRight := make([]float64, n)

		for i := 0; i < n; i++ {
			samplesLeft[i] =  buff[i][0]
			samplesRight[i] = buff[i][1]
		}

		msg := &message.StreamData{SamplesLeft: samplesLeft, SamplesRight: samplesRight}
		msgData, _ := message.ToBuffer(msg)
		this.Messenger.Message <- &message.WriteRequest{DeviceName: "*", Message: msgData}

		time.Sleep(format.SampleRate.D(n))
	}
}
