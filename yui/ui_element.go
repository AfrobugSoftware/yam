package yui

import "yam/y3d"

const (
	UI_NONE          = 0
	UI_FIXED         = 2
	UI_FIT           = 4
	UI_LEFT_TO_RIGHT = 8
	UI_TOP_TO_BOTTOM = 16
	UI_CENTER        = 32
)

type UIElement struct {
	Position y3d.Vec2
	Size     y3d.Vec2
	Color    [4]uint8
	Children []UIElement
	Padding  y3d.Vec4
	Flag     uint32
	ChildGap int
}
