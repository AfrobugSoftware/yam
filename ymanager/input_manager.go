package ymanager

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

type InputManager struct {
	CurKeyState       []uint8
	PrevKeyState      []uint8
	CurMouseKeyState  uint32
	PrevMouseKeyState uint32
	MousePosition     y3d.Vec3
	ShowCursor        int
	ScreenWidth       int32
	ScreenHeight      int32
	IsRelative        bool
	ScrollWheelPos    y3d.Vec3
	ScrollWheelDir    uint32
	MaxMouseSpeed     float32
}

func NewInputManager() *InputManager {
	return &InputManager{
		CurKeyState:  make([]uint8, sdl.NUM_SCANCODES),
		PrevKeyState: make([]uint8, sdl.NUM_SCANCODES),
	}
}

func (im *InputManager) Update() {
	copy(im.PrevKeyState, im.CurKeyState)
	im.PrevMouseKeyState = im.CurMouseKeyState
	clear(im.CurKeyState)
	im.CurMouseKeyState = 0
	im.ScrollWheelDir = 0
	im.ScrollWheelPos = y3d.Vec3{}

	var x, y int32
	var state uint32
	if im.IsRelative {
		x, y, state = sdl.GetRelativeMouseState()
	} else {
		x, y, state = sdl.GetMouseState()
	}
	im.MousePosition = im.convertToOpenGLCoords(x, y)
	im.CurMouseKeyState = state
}
func (im *InputManager) convertToOpenGLCoords(x, y int32) y3d.Vec3 {
	return y3d.Vec3{
		X: float32(x) - float32(im.ScreenWidth)/2,
		Y: float32(im.ScreenHeight)/2 - float32(y),
	}
}
func (im *InputManager) GetKeyState(key int) uint8 {
	if key < 0 || key >= sdl.NUM_SCANCODES {
		return BUTTON_NONE
	}
	p := im.PrevKeyState[key]
	c := im.CurKeyState[key]
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

func (im *InputManager) GetMouseButtonState(button int) uint8 {
	if button < 1 || button > MOUSE_MAX {
		return BUTTON_NONE
	}
	p := im.PrevMouseKeyState & (1 << (button - 1))
	c := im.CurMouseKeyState & (1 << (button - 1))
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
