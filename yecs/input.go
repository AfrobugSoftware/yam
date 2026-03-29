package yecs

import (
	"yam/y3d"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	BUTTON_NONE uint8 = iota
	BUTTON_RELEASED
	BUTTON_PRESSED
	BUTTON_HELD
)

const (
	MOUSE_LEFT   = 1
	MOUSE_RIGHT  = 2
	MOUSE_MIDDLE = 3
	MOUSE_X1     = 4
	MOUSE_X2     = 5
	MOUSE_MAX    = 6
)

type Input struct {
	isystem *InputSystem
}

func (i Input) GetKeyValue(key int) uint8 {
	if key < 0 || key >= sdl.NUM_SCANCODES {
		return BUTTON_NONE
	}
	return i.isystem.CurKeyState[key]
}

func (i Input) GetKeyState(key int) uint8 {
	if key < 0 || key >= sdl.NUM_SCANCODES {
		return BUTTON_NONE
	}
	p := i.isystem.PrevKeyState[key]
	c := i.isystem.CurKeyState[key]
	r := BUTTON_NONE
	switch c {
	case 0:
		switch p {
		case 0:
			r = BUTTON_NONE
		case 1:
			r = BUTTON_RELEASED
		}
	case 1:
		switch p {
		case 0:
			r = BUTTON_PRESSED
		case 1:
			r = BUTTON_HELD
		}
	}
	return r
}

func (i Input) GetMousePosition() y3d.Vec3 {
	return i.isystem.MousePosition
}

func (i Input) GetMouseKeyState(key uint32) uint8 {
	if key >= MOUSE_MAX {
		return BUTTON_NONE
	}
	p := sdl.Button(key) & i.isystem.PrevMouseKeyState
	c := sdl.Button(key) & i.isystem.CurMouseKeyState
	r := BUTTON_NONE
	switch c {
	case 0:
		switch p {
		case 0:
			r = BUTTON_NONE
		case 1:
			r = BUTTON_RELEASED
		}
	case 1:
		switch p {
		case 0:
			r = BUTTON_PRESSED
		case 1:
			r = BUTTON_HELD
		}
	}
	return r
}

type InputSystem struct {
	CurKeyState       []uint8
	PrevKeyState      []uint8
	CurMouseKeyState  uint32
	PrevMouseKeyState uint32
	MousePosition     y3d.Vec3
	ShowCursor        int
	ScreenWidth       int32
	ScreenHeight      int32
	IsRelative        bool
}

func (is *InputSystem) Init() {
	is.CurKeyState = make([]uint8, sdl.NUM_SCANCODES)
	is.PrevKeyState = make([]uint8, sdl.NUM_SCANCODES)
	sdl.ShowCursor(is.ShowCursor)
}

func convertToOpenGLCoords(is *InputSystem, x, y int32) y3d.Vec3 {
	return y3d.Vec3{
		X: float32(x) - float32(is.ScreenWidth)/2,
		Y: float32(is.ScreenHeight)/2 - float32(y),
	}
}

func (is *InputSystem) UpdateMouse() {
	var x, y int32
	var state uint32
	if is.IsRelative {
		x, y, state = sdl.GetRelativeMouseState()
	} else {
		x, y, state = sdl.GetMouseState()
	}
	is.MousePosition = convertToOpenGLCoords(is, x, y)
	is.CurMouseKeyState = state
}

func (is *InputSystem) SetRelavtive(set bool) {
	is.IsRelative = set
	sdl.SetRelativeMouseMode(set)
}

func (is *InputSystem) CreateInput() Input {
	return Input{
		isystem: is,
	}
}

func (is *InputSystem) PrepareToUpdate() {
	copy(is.PrevKeyState, is.CurKeyState)
	is.PrevMouseKeyState = is.CurMouseKeyState

	clear(is.CurKeyState)
	is.CurMouseKeyState = 0
}

func (is *InputSystem) Query() []ComponentId {
	return []ComponentId{InputComponent}
}

func (is *InputSystem) Update(w *World, dt float64, entites []EntityId) {}
func (is *InputSystem) Shutdown()                                       {}
