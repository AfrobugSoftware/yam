package yui

import (
	"yam/y3d"
)

func Button(text string, size y3d.Vec2, color [4]uint8, font *Font) *UIElement {
	button := NewUIElement(
		y3d.Vec2{},
		size,
	).Set(UI_ALIGN_CENTER | UI_FIT).SetRadius(10).SetColor(
		color,
	).Add(
		&UIElement{
			Flag:  UI_ALIGN_CENTER,
			Text:  text,
			Font:  font,
			Color: [4]uint8{255, 255, 255, 255},
		},
	)
	return button
}

func IconButton(icon rune, size y3d.Vec2, font *Font) *UIElement {
	button := NewUIElement(
		y3d.Vec2{},
		size,
	).Set(UI_ALIGN_CENTER | UI_FIT).SetRadius(10).SetColor(
		[4]uint8{0, 0, 0, 255},
	).Add(
		&UIElement{
			Flag:     UI_ALIGN_CENTER,
			IconRune: icon,
			Font:     font,
			Color:    [4]uint8{255, 255, 255, 255},
		},
	)
	return button
}
