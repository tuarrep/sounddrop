package main

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"os"
	"time"
)

func main() {
	done := make(chan struct{})

	f, _ := os.Open("test.wav")
	s, format, _ := wav.Decode(f)

	buf := make([][2]float64, 512)
	s.Stream(buf)

	fmt.Printf("%v\n", buf)

	_ = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	speaker.Play(beep.Seq(s, beep.Callback(func() {
		close(done)
	})))

	<- done
}