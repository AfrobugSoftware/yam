package yecs

import (
	"bytes"
	"log"
	"os"
	"yam/y3d"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
)

// channels
const (
	MONO   = 1
	STEREO = 2
)

type AudioData struct {
	Pos           y3d.Vec3
	Player        *oto.Player
	CurrentVolume int
}

func (a AudioData) Play() {
	if !a.Player.IsPlaying() {
		a.Player.Play()
	}
}

type AudioSystem struct {
	Channels   int
	SampleRate int
	Format     oto.Format
	Context    *oto.Context
}

func (as *AudioSystem) Init() {
	op := &oto.NewContextOptions{}
	op.SampleRate = as.SampleRate
	op.ChannelCount = as.Channels
	op.Format = as.Format
	c, ready, err := oto.NewContext(op)
	if err != nil {
		panic(err)
	}
	//wait for hardware to be ready
	log.Println("Starting audio system")
	<-ready
	as.Context = c
}

func (as *AudioSystem) CreateAudoDataStream(audioFile string, pos y3d.Vec3) (*AudioData, error) {
	file, err := os.Open(audioFile)
	if err != nil {
		return nil, err
	}
	//defer file.Close() //file is an io.reader, for streams

	decodedMp3, err := mp3.NewDecoder(file)
	if err != nil {
		return nil, err
	}
	a := &AudioData{
		Player: as.Context.NewPlayer(decodedMp3),
		Pos:    pos,
	}
	return a, nil
}

func (as *AudioSystem) CreateAudoData(audioFile string, pos y3d.Vec3) (*AudioData, error) {
	file, err := os.ReadFile(audioFile)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(file)
	decodedMp3, err := mp3.NewDecoder(reader)
	if err != nil {
		return nil, err
	}
	a := &AudioData{
		Player: as.Context.NewPlayer(decodedMp3),
		Pos:    pos,
	}
	return a, nil
}

func (as *AudioSystem) Shutdown() {}
func (as *AudioSystem) Query() []ComponentId {
	return []ComponentId{AudioComponent}
}
func (as *AudioSystem) Update(w *World, dt float64, entities []EntityId) {
	for _, e := range entities {
		a := w.GetComponent(e, AudioComponent).(AudioData)
		a.Play()
	}
}
