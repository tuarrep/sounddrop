package service

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/tuarrep/sounddrop/message"
	"github.com/tuarrep/sounddrop/structure"
	"github.com/tuarrep/sounddrop/util"
	"math"
	"time"
)

// Player audio player service
type Player struct {
	Message   chan proto.Message
	log       *logrus.Entry
	Messenger *Messenger
	format    beep.Format
	tsq       *structure.TimedSampleQueue
	silence   beep.Streamer
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
	p.tsq = structure.NewTimedSampleQueue(10 * int(p.format.SampleRate))
	p.silence = beep.Silence(-1)

	_ = speaker.Init(p.format.SampleRate, 512)
	speaker.Play(beep.Seq(beep.Callback(func() {
		//p.tsq.Start()
	}), p, beep.Callback(func() {
		p.log.Warn("Speaker ended stream. This should not have happened!")
	})))

	for {
		select {
		case msg := <-p.Message:
			switch m := msg.(type) {
			case *message.StreamData:
				for index, sample := range m.SamplesLeft {
					samples := [2]float64{sample, m.SamplesRight[index]}
					p.tsq.Add(samples, m.NextAt+int64(p.format.SampleRate.D(index)*time.Nanosecond))
				}
			}
		}
	}
}

// GetChan returns messaging chan
func (p *Player) GetChan() chan proto.Message {
	return p.Message
}

// Stream stream audio samples from received data
func (p *Player) Stream(samples [][2]float64) (n int, ok bool) {
	neededLength := len(samples)
	silenceCount := 0
	now := time.Now().UnixNano()

	_, t := p.tsq.Peek()

	for now-t > int64(10*time.Millisecond) {
		// We are late, dropping samples
		p.tsq.Remove()
		now = time.Now().UnixNano()
		_, t = p.tsq.Peek()
	}

	if t-now > int64(10*time.Millisecond) {
		// We are before the time of the next sample, padding with silence
		p.log.Debug(fmt.Sprintf("Next packet is scheduled in %v", time.Duration(t-now)))
		silenceCount = int(math.Min(float64(p.format.SampleRate.N(time.Duration(t-now))), float64(len(samples))))
		p.silence.Stream(samples[:silenceCount])
	}

	for i := silenceCount; i < neededLength; i++ {
		samples[i], _ = p.tsq.Remove()
	}

	return len(samples), len(samples) > 0
}

// Err return streaming error (never)
func (p *Player) Err() error {
	return nil
}
