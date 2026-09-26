package ycore

import (
	"bytes"
	"errors"
	"os"
	"yam/y3d"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
)

type AudioListener struct {
	Position y3d.Vec3
	Velocity y3d.Vec3
}

type AudioPlayer struct {
	Player    *oto.Player
	Position  y3d.Vec3
	Direction y3d.Vec3
	Radius    float32
}

func NewAudioPlayer(player *oto.Player) *AudioPlayer {
	return &AudioPlayer{
		Player: player,
	}
}

type AudioManager struct {
	Players           map[string]*AudioPlayer
	PlayerStreamFiles map[string]*os.File
	Listener          *AudioListener
	OtoContext        *oto.Context
}

func NewAudioManager() *AudioManager {
	op := &oto.NewContextOptions{}
	op.SampleRate = 44100
	op.ChannelCount = 2
	op.Format = oto.FormatSignedInt16LE
	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		panic("oto.NewContext failed: " + err.Error())
	}
	<-readyChan
	if err := otoCtx.Err(); err != nil {
		panic("oto initialization failed: " + err.Error())
	}
	return &AudioManager{
		Players:    make(map[string]*AudioPlayer),
		OtoContext: otoCtx,
	}
}

func (a *AudioManager) LoadMP3(name, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	decoded, err := mp3.NewDecoder(file)
	if err != nil {
		return err
	}

	player := a.OtoContext.NewPlayer(decoded)
	a.Players[name] = NewAudioPlayer(player)
	a.PlayerStreamFiles[name] = file
	return nil
}

func (a *AudioManager) LoadMP3Static(name, filename string) error {
	fbytes, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(fbytes)
	decoded, err := mp3.NewDecoder(buf)
	if err != nil {
		return err
	}
	p := a.OtoContext.NewPlayer(decoded)
	a.Players[name] = NewAudioPlayer(p)
	return nil
}

func (a *AudioManager) Play(name string) error {
	p, ok := a.Players[name]
	if !ok {
		return errors.New("no audio player with that name")
	}
	p.Player.Play()
	return nil
}

func (a *AudioManager) ClosePlayer(name string) {
	f, ok := a.PlayerStreamFiles[name]
	if ok {
		f.Close()
		delete(a.PlayerStreamFiles, name)
	}
	delete(a.Players, name)
}
