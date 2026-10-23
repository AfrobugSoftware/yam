package yui

import (
	"yam/y3d"
)

const (
	UI_NONE                 = 0
	UI_FIXED                = 2
	UI_FIT                  = 4
	UI_LAYOUT_LEFT_TO_RIGHT = 8
	UI_LAYOUT_TOP_TO_BOTTOM = 16
	UI_HIDDEN               = 32
	UI_GROW_X               = 64
	UI_GROW_Y               = 128
	UI_ALIGN_LEFT           = 256
	UI_ALIGN_RIGHT          = 512
	UI_ALIGN_CENTER_X       = 1024
	UI_ALIGN_CENTER_Y       = 2048
	UI_ALIGN_TOP            = 4096
	UI_ALIGN_BOTTOM         = 8192
	UI_FOCUSED              = 16384
	UI_ACTIVE               = 1 << 15
	UI_SCROLLABLE           = 1 << 16
	UI_FLOATING             = 1 << 17
	UI_TEXT                 = 1 << 18
	UI_ABSOLUTE_POSITION    = 1 << 19
)
const (
	UI_GROW         = UI_GROW_X | UI_GROW_Y
	UI_ALIGN_CENTER = UI_ALIGN_CENTER_X | UI_ALIGN_CENTER_Y
)

const (
	UI_PRESSED = 1 << iota
	UI_HOVERED
)

type UIElement struct {
	Position      y3d.Vec2
	Size          y3d.Vec2
	BaseSize      y3d.Vec2
	MinSize       y3d.Vec2
	MaxSize       y3d.Vec2
	Radius        float32
	Color         [4]uint8
	TextureBase   [2]float32
	Children      []*UIElement
	Padding       y3d.Vec4
	Flag          uint32
	RenderFlag    uint32
	Text          string
	ChildGap      int
	minSize       y3d.Vec2
	Lines         []string
	MouseCallback func(any)
	Font          *Font
	Name          string
	Tex           uint32
}

func NewUIElement(baseSize y3d.Vec2) *UIElement {
	return &UIElement{
		Flag:     UI_FIT | UI_LAYOUT_LEFT_TO_RIGHT,
		ChildGap: 0,
		BaseSize: baseSize,
		Size:     y3d.Vec2{X: 1, Y: 1},
		Padding:  y3d.Vec4{X: 2, Y: 2, Z: 2, W: 2},
	}
}
func (ui *UIElement) Add(eui ...*UIElement) {
	ui.Children = append(ui.Children, eui...)
}
func (ui *UIElement) BoundingBox() y3d.Rect {
	return y3d.Rect{
		X:      int(ui.Position.X),
		Y:      int(ui.Position.Y),
		Width:  int(ui.Position.X + ui.Size.X),
		Height: int(ui.Position.Y + ui.Size.Y),
	}
}

func (ui *UIElement) Set(flag uint32) {
	ui.Flag |= flag
}

func (ui *UIElement) Unset(flag uint32) {
	ui.Flag = ui.Flag &^ flag
}
