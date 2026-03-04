package yam

import (
	"errors"
	"yam/ygame"

	"github.com/veandco/go-sdl2/sdl"
)

// states
const (
	Idle = iota
	Running
	Jumping
	MaxState
)
const (
	FPS = 10
)

func NilPlayerState(a ygame.Actor) error {
	return nil
}

func EnterIdle(a ygame.Actor) error {
	as, err := a.(*ygame.AnimSprite)
	if !err {
		return errors.New("actor not a AnimeSprite")
	}
	as.SelectSheet(Idle)
	as.FPS = FPS
	as.NumFrames = 10
	return nil
}

func EnterRunning(a ygame.Actor) error {
	as, err := a.(*ygame.AnimSprite)
	if !err {
		return errors.New("actor not a AnimeSprite")
	}
	as.SelectSheet(Running)
	as.FPS = FPS
	as.NumFrames = 8
	return nil
}

func EnterJumping(a ygame.Actor) error {
	as, err := a.(*ygame.AnimSprite)
	if !err {
		return errors.New("actor not a AnimeSprite")
	}
	as.SelectSheet(Jumping)
	as.FPS = FPS
	as.NumFrames = 12
	return nil
}

func UpdateIdle(a ygame.Actor, dt float64) error {
	as, err := a.(*ygame.AnimSprite)
	if !err {
		return errors.New("actor not a AnimeSprite")
	}
	g := ygame.GetGame()
	if g.KeyState[sdl.SCANCODE_A] != 0 || g.KeyState[sdl.SCANCODE_D] != 0 {
		as.StateMachine.ChangeState(Running)
	}
	if g.KeyState[sdl.SCANCODE_W] != 0 {
		as.StateMachine.ChangeState(Jumping)
	}
	return nil
}

func UpdateRunning(a ygame.Actor, dt float64) error {
	as, err := a.(*ygame.AnimSprite)
	if !err {
		return errors.New("actor not a AnimeSprite")
	}
	g := ygame.GetGame()
	if g.KeyState[sdl.SCANCODE_S] != 0 {
		as.StateMachine.ChangeState(Idle)
	}
	if g.KeyState[sdl.SCANCODE_W] != 0 {
		as.StateMachine.ChangeState(Jumping)
	}
	return nil
}
func UpdateJumping(a ygame.Actor, dt float64) error {
	as, err := a.(*ygame.AnimSprite)
	if !err {
		return errors.New("actor not a AnimeSprite")
	}
	g := ygame.GetGame()
	if g.KeyState[sdl.SCANCODE_D] != 0 {
		as.StateMachine.ChangeState(Running)
	}
	if g.KeyState[sdl.SCANCODE_S] != 0 {
		as.StateMachine.ChangeState(Idle)
	}
	return nil
}

var (
	EnterAnimStateTable = [][]ygame.EnterExitStateFunc{
		//leaving idle
		{
			EnterIdle,
			EnterRunning,
			EnterJumping,
		},
		//leaving running
		{
			EnterIdle,
			NilPlayerState,
			EnterJumping,
		},
		{
			EnterIdle,
			EnterRunning,
			NilPlayerState,
		},
	}
	ExitAnimStateTable = [][]ygame.EnterExitStateFunc{
		//leaving idle
		{
			NilPlayerState,
			NilPlayerState,
			NilPlayerState,
		},
		//leaving running
		{
			NilPlayerState,
			NilPlayerState,
			NilPlayerState,
		},
		{
			NilPlayerState,
			NilPlayerState,
			NilPlayerState,
		},
	}

	UpdateAnimStateTable = []ygame.UpdateStateFunc{
		UpdateIdle,
		UpdateRunning,
		UpdateJumping,
	}
)
