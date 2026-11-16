package ycore

type EnterExitStateFunc func(w *World, a EntityId) error
type UpdateStateFunc func(w *World, a EntityId, dt float64) error
type State int
type StateMachine struct {
	PrevState State
	CurState  State
	OnUpdate  []UpdateStateFunc
	OnEnter   [][]EnterExitStateFunc
	OnLeave   [][]EnterExitStateFunc
}

func (s *StateMachine) Update(w *World, dt float64, e EntityId) {
	if s.OnUpdate != nil {
		s.OnUpdate[s.CurState](w, e, dt)
	}
}

func (s *StateMachine) ChangeState(w *World, e EntityId, state State) {
	if s.OnLeave == nil || s.OnEnter == nil {
		return
	}
	err := s.OnLeave[s.PrevState][s.CurState](w, e)
	if err != nil {
		return
	}
	save := s.PrevState
	s.PrevState, s.CurState = s.CurState, state
	err = s.OnEnter[s.PrevState][s.CurState](w, e)
	if err != nil {
		s.CurState = s.PrevState
		s.PrevState = save
	}
}
