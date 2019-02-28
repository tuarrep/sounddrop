package service

import (
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/tuarrep/sounddrop/message"
	"github.com/tuarrep/sounddrop/util"
	"os"
	"time"
)

// Player audio player service
type Player struct {
	Message   chan proto.Message
	log       *logrus.Entry
	Messenger *Messenger
	format    beep.Format
	data      [][2]float64
	s         beep.StreamSeekCloser
}

// Stop clean service when stopped by supervisor
func (p *Player) Stop() {
	p.log.Info("Player stopped.")
}

// Serve main service code
func (p *Player) Serve() {
	p.log = util.GetContextLogger("service/player.go", "Services/Player")
	p.log.Info("Player starting...")

	p.Message = make(chan proto.Message)
	p.Messenger.Register(message.StreamDataMessage, p)
	p.format = beep.Format{SampleRate: 44100, NumChannels: 2, Precision: 2}
	p.data = make([][2]float64, p.format.SampleRate.N(5*time.Second))

	f, _ := os.Open("test.wav")
	p.s, _, _ = wav.Decode(f)
	//p.s.Stream(p.data)

	_ = speaker.Init(p.format.SampleRate, p.format.SampleRate.N(time.Second/10))
	speaker.Play(beep.Seq(p, beep.Callback(func() {
		p.log.Warn("Speaker ended stream. This should not have happened!")
	})))

	for {
		select {
		case msg := <-p.Message:
			switch m := msg.(type) {
			case *message.StreamData:
				go p.handleStreamData(m)
			}
		}
	}
}

// GetChan returns messaging chan
func (p *Player) GetChan() chan proto.Message {
	return p.Message
}

func (p *Player) handleStreamData(m *message.StreamData) {
	bufferLength := len(m.SamplesRight)
	if bufferLength > len(m.SamplesLeft) {
		bufferLength = len(m.SamplesLeft)
	}

	samples := make([][2]float64, bufferLength)

	for i := 0; i < bufferLength; i++ {
		samples[i] = [2]float64{m.SamplesLeft[i], m.SamplesRight[i]}
	}

	p.data = append(p.data, samples...)
}

// Stream stream audio samples from received data
func (p *Player) Stream(samples [][2]float64) (n int, ok bool) {
	realLength := len(samples)
	if realLength > len(p.data) {
		realLength = len(p.data)
	}

	fetchedSamples := p.data[:realLength]

	for i := 0; i < len(fetchedSamples); i++ {
		samples[i][0] = fetchedSamples[i][0]
		samples[i][1] = fetchedSamples[i][1]
	}

	p.data = p.data[realLength:]

	return len(samples), len(samples) > 0
}

// Err return streaming error (never)
func (p *Player) Err() error {
	return nil
}
