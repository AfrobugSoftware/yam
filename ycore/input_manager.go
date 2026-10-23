package ycore

import (
	"fmt"
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

const (
	CTRL = iota
	SHIFT
	ALT
	SUPER
)

type PadController struct {
	Control    *sdl.GameController
	PreButtons []byte
	CurButtons []byte
}

type InputManager struct {
	CurKeyState       []uint8
	Window            *sdl.Window
	PrevKeyState      []uint8
	ModKeyState       [4]uint8
	PrevModKeyState   [4]uint8
	CurMouseKeyState  [3]uint8
	PrevMouseKeyState [3]uint8
	MousePosition     y3d.Vec2
	ShowCursor        int
	IsRelative        bool
	IsResized         bool
	ScrollWheelPos    y3d.Vec3
	ScrollWheelDir    uint32
	MaxMouseSpeed     float32
	MouseCage         y3d.Rect
	TextInput         string
	Controllers       map[int]*PadController
}

func NewInputManager(window *sdl.Window, width, height int) *InputManager {
	return &InputManager{
		CurKeyState:  make([]uint8, sdl.NUM_SCANCODES),
		PrevKeyState: make([]uint8, sdl.NUM_SCANCODES),
		Controllers:  make(map[int]*PadController),
		MouseCage: y3d.Rect{
			X:      0,
			Y:      0,
			Width:  width,
			Height: height,
		},
		Window: window,
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
func (im *InputManager) GetKeyStateMod(key int) uint8 {
	if key < 0 || key >= SUPER {
		return BUTTON_NONE
	}
	p := im.PrevModKeyState[key]
	c := im.ModKeyState[key]
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
		return sdl.RELEASED
	}
	return im.CurMouseKeyState[button-1]
}

func (im *InputManager) ConnectController(idx int) {
	if sdl.IsGameController(idx) {
		controller := sdl.GameControllerOpen(idx)
		if controller != nil {
			im.Controllers[idx] = &PadController{
				PreButtons: make([]byte, 7),
				CurButtons: make([]byte, 7),
				Control:    controller,
			}
		}
	}
}

func (im *InputManager) DisconnectController(idx int) {
	c, ok := im.Controllers[idx]
	if ok {
		c.Control.Close()
		delete(im.Controllers, idx)
	}
}

func (im *InputManager) ProcessInput() bool {
	copy(im.PrevKeyState, im.CurKeyState)
	copy(im.PrevModKeyState[:], im.ModKeyState[:])
	copy(im.PrevMouseKeyState[:], im.CurMouseKeyState[:])
	clear(im.CurKeyState)
	clear(im.ModKeyState[:])

	im.ScrollWheelDir = 0
	im.ScrollWheelPos = y3d.Vec3{}
	im.IsResized = false
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch event.GetType() {
		case sdl.QUIT:
			return false
		case sdl.CONTROLLERDEVICEADDED:
			c := event.(*sdl.ControllerDeviceEvent)
			im.ConnectController(int(c.Which))
		case sdl.CONTROLLERDEVICEREMOVED:
			c := event.(*sdl.ControllerDeviceEvent)
			im.DisconnectController(int(c.Which))
		case sdl.MOUSEBUTTONDOWN:
			m := event.(*sdl.MouseButtonEvent)
			im.CurMouseKeyState[int(m.Button-1)] = m.State
		case sdl.MOUSEBUTTONUP:
			m := event.(*sdl.MouseButtonEvent)
			im.CurMouseKeyState[int(m.Button-1)] = m.State
		case sdl.MOUSEWHEEL:
			w := event.(*sdl.MouseWheelEvent)
			im.ScrollWheelPos = y3d.Vec3{
				X: float32(w.X),
				Y: float32(w.Y),
			}
			im.ScrollWheelDir = w.Direction
		case sdl.WINDOWEVENT_SIZE_CHANGED:
			fmt.Println("resized")
			im.IsResized = true
		case sdl.TEXTINPUT:
			c := event.(*sdl.TextInputEvent)
			im.TextInput = c.GetText()
		case sdl.KEYDOWN:
			e := event.(*sdl.KeyboardEvent)
			mod := sdl.Keymod(e.Keysym.Mod)
			if mod&sdl.KMOD_CTRL != 0 {
				im.ModKeyState[CTRL] = e.State
			}
			if mod&sdl.KMOD_ALT != 0 {
				im.ModKeyState[ALT] = e.State
			}
			if mod&sdl.KMOD_SHIFT != 0 {
				im.ModKeyState[SHIFT] = e.State
			}
			if mod&sdl.KMOD_GUI != 0 {
				im.ModKeyState[SUPER] = e.State
			}
		}

	}
	state := sdl.GetKeyboardState()
	if state != nil {
		if state[sdl.SCANCODE_ESCAPE] != 0 {
			return false
		}
		copy(im.CurKeyState, state)
	}
	var x, y int32
	if im.IsRelative {
		x, y, _ = sdl.GetRelativeMouseState()
	} else {
		x, y, _ = sdl.GetMouseState()
	}
	im.MousePosition.X = float32(x) //float32(min(max(x, int32(im.MouseCage.X)), int32(im.MouseCage.Width)))
	im.MousePosition.Y = float32(y) //float32(min(max(y, int32(im.MouseCage.Y)), int32(im.MouseCage.Height)))
	return true
}
