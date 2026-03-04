package ygame

import "errors"

var (
	ErrorInvalidState = errors.New("invalid state")
)

// state machine holds state tables for entering, update and leaving
// for example Enter[a][b](s, actor) runs for a an outgoing state for
// a and incoming state for b

type EnterExitStateFunc func(a Actor) error
type UpdateStateFunc func(a Actor, dt float64) error

type StateMachine struct {
	Owner     Actor
	PrevState int
	CurState  int
	OnUpdate  []UpdateStateFunc
	OnEnter   [][]EnterExitStateFunc
	OnLeave   [][]EnterExitStateFunc
}

func InvalidState(a Actor, dt float64) error {
	return ErrorInvalidState
}

func NilState(a Actor, dt float64) error {
	return nil
}

func (s *StateMachine) Update(dt float64) {
	if s.OnUpdate != nil {
		f := s.OnUpdate[s.CurState]
		f(s.Owner, dt)
	}
}

func (s *StateMachine) ChangeState(state int) {
	err := s.OnLeave[s.PrevState][s.CurState](s.Owner)
	if err != nil {
		return
	}
	save := s.PrevState
	s.PrevState = s.CurState
	s.CurState = state
	err = s.OnEnter[s.PrevState][s.CurState](s.Owner)
	if err != nil {
		s.CurState = s.PrevState
		s.PrevState = save
	}
}
