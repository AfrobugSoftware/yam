package yecs

// state machine holds state tables for entering, update and leaving
// for example Enter[a][b](s, actor) runs for a an outgoing state for
// a and incoming state for b

type EnterExitStateFunc func(w *World, a EntityId) error
type UpdateStateFunc func(w *World, a EntityId, dt float64) error

type StateMachine struct {
	PrevState int
	CurState  int
	OnUpdate  []UpdateStateFunc
	OnEnter   [][]EnterExitStateFunc
	OnLeave   [][]EnterExitStateFunc
}

func (s *StateMachine) Update(w *World, dt float64, e EntityId) {
	if s.OnUpdate != nil {
		f := s.OnUpdate[s.CurState]
		f(w, e, dt)
	}
}

func (s *StateMachine) ChangeState(w *World, e EntityId, state int) {
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

type StateSystem struct{}

func (ss *StateSystem) Init()     {}
func (ss *StateSystem) Shutdown() {}

func (ss *StateSystem) Query() []ComponentId {
	return []ComponentId{StateComponent}
}

func (ss *StateSystem) Update(w *World, dt float64, entities []EntityId) {
	for _, e := range entities {
		machine, ok := w.GetComponent(e, StateComponent).(StateMachine)
		if !ok {
			continue
		}
		machine.Update(w, dt, e)
	}
}
