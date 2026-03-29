package yecs

import (
	"bytes"
	"log"
	"os"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
)

// channels
const (
	MONO   = 1
	STEREO = 2
)

type AudioData struct {
	DisTanceToListener float32
	Player             string
	CurrentVolume      float64
}

type AudioPlayer struct {
	file   *os.File
	player *oto.Player
}

type AudioSystem struct {
	Channels   int
	SampleRate int
	Format     oto.Format
	Context    *oto.Context
	Players    map[string]AudioPlayer
}

func NewAudioSystem(channels int, sampleRate int, format oto.Format) *AudioSystem {
	return &AudioSystem{
		Channels:   channels,
		SampleRate: sampleRate,
		Format:     format,
		Players:    make(map[string]AudioPlayer),
	}
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

func (as *AudioSystem) CreateAudoDataStream(audioFile string, name string) error {
	file, err := os.Open(audioFile)
	if err != nil {
		return err
	}
	decodedMp3, err := mp3.NewDecoder(file)
	if err != nil {
		return err
	}
	as.Players[name] = AudioPlayer{
		file:   file,
		player: as.Context.NewPlayer(decodedMp3),
	}
	return nil
}

func (as *AudioSystem) CreateAudoData(audioFile string, name string) error {
	file, err := os.ReadFile(audioFile)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(file)
	decodedMp3, err := mp3.NewDecoder(reader)
	if err != nil {
		return err
	}
	as.Players[name] = AudioPlayer{
		file:   nil,
		player: as.Context.NewPlayer(decodedMp3),
	}
	return nil
}

func (as *AudioSystem) RemovePlayer(name string) {
	player, ok := as.Players[name]
	if ok {
		if player.file != nil {
			player.file.Close()
		}
		delete(as.Players, name)
	}
}

func (as *AudioSystem) Shutdown() {
	for _, p := range as.Players {
		if p.file != nil {
			p.file.Close()
		}
	}
}
func (as *AudioSystem) Query() []ComponentId {
	return []ComponentId{AudioComponent}
}
func (as *AudioSystem) Update(w *World, dt float64, entities []EntityId) {
	for _, e := range entities {
		a := w.GetComponent(e, AudioComponent).(AudioData)
		player, ok := as.Players[a.Player]
		if ok {
			player.player.SetVolume(a.CurrentVolume)
			player.player.Play()
		}
	}
}
