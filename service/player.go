package service

import (
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"os"
	"github.com/mafzst/sounddrop/message"
	"github.com/mafzst/sounddrop/util"
	"time"
)

type Player struct {
	Message   chan proto.Message
	log       *logrus.Entry
	Messenger *Messenger
	format    beep.Format
	data      [][2]float64
	s         beep.StreamSeekCloser
}

func (this *Player) Stop() {
	this.log.Info("Player stopped.")
}

func (this *Player) Serve() {
	this.log = util.GetContextLogger("service/player.go", "Services/Player")
	this.log.Info("Player starting...")

	this.Message = make(chan proto.Message)
	this.Messenger.Register(message.StreamDataMessage, this)
	this.format = beep.Format{SampleRate: 44100, NumChannels: 2, Precision: 2}
	this.data = make([][2]float64, this.format.SampleRate.N(5*time.Second))

	f, _ := os.Open("test.wav")
	this.s, _, _ = wav.Decode(f)
	//this.s.Stream(this.data)

	_ = speaker.Init(this.format.SampleRate, this.format.SampleRate.N(time.Second/10))
	speaker.Play(beep.Seq(this, beep.Callback(func() {
		this.log.Warn("Speaker ended stream. This should not have happened!")
	})))

	for {
		select {
		case msg := <-this.Message:
			switch m := msg.(type) {
			case *message.StreamData:
				go this.handleStreamData(m)
			}
		}
	}
}

func (this *Player) GetChan() chan proto.Message {
	return this.Message
}

func (this *Player) handleStreamData(m *message.StreamData) {
	bufferLength := len(m.SamplesRight)
	if bufferLength > len(m.SamplesLeft) {
		bufferLength = len(m.SamplesLeft)
	}

	samples := make([][2]float64, bufferLength)

	for i := 0; i < bufferLength; i++ {
		samples[i] = [2]float64{m.SamplesLeft[i], m.SamplesRight[i]}
	}

	this.data = append(this.data, samples...)
}

func (this *Player) Stream(samples [][2]float64) (n int, ok bool) {
	realLength := len(samples)
	if realLength > len(this.data) {
		realLength = len(this.data)
	}

	fetchedSamples := this.data[:realLength]

	for i := 0; i < len(fetchedSamples); i++ {
		samples[i][0] = fetchedSamples[i][0]
		samples[i][1] = fetchedSamples[i][1]
	}

	this.data = this.data[realLength:]

	return len(samples), len(samples) > 0
}

func (this *Player) Err() error {
	return nil
}
