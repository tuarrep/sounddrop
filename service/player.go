package service

import (
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
	"github.com/tuarrep/sounddrop/message"
	"github.com/tuarrep/sounddrop/structure"
	"github.com/tuarrep/sounddrop/util"
	"sync"
)

// Player audio player service
type Player struct {
	Message   chan proto.Message
	log       *logrus.Entry
	Messenger *Messenger
	format    beep.Format
	data      [][2]float64
	tsq       *structure.TimedSampleQueue
	samples   chan *message.StreamData
	dataMutex sync.Mutex
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
	p.data = make([][2]float64, 0)
	p.tsq = structure.NewTimedSampleQueue()
	p.samples = make(chan *message.StreamData)
	p.tsq.Subscribe(p.samples)
	p.dataMutex = sync.Mutex{}

	_ = speaker.Init(p.format.SampleRate, 512)
	speaker.Play(beep.Seq(beep.Callback(func() {
		p.tsq.Start()
	}), p, beep.Callback(func() {
		p.log.Warn("Speaker ended stream. This should not have happened!")
	})))

	for {
		select {
		case msg := <-p.Message:
			switch m := msg.(type) {
			case *message.StreamData:
				popAt, _ := ptypes.Timestamp(m.NextAt)
				p.tsq.Push(m, popAt)
			}
		case sample := <-p.samples:
			go p.handleStreamData(sample)
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

	keepSamplesNb := len(p.data) - len(samples)
	if keepSamplesNb < 0 {
		keepSamplesNb = 0
	}

	p.dataMutex.Lock()
	defer p.dataMutex.Unlock()
	p.data = append(p.data[keepSamplesNb:], samples...)
}

// Stream stream audio samples from received data
func (p *Player) Stream(samples [][2]float64) (n int, ok bool) {
	realLength := len(samples)
	if realLength > len(p.data) {
		//fmt.Print("!")
		realLength = len(p.data)
	}

	p.dataMutex.Lock()
	defer p.dataMutex.Unlock()
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
