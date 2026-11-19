package yui

import "yam/y3d"

func Row(padding y3d.Vec4, spacing int, items ...*UIElement) *UIElement {
	row := NewUIElement(y3d.Vec2{}, y3d.Vec2{})
	row.Set(UI_LAYOUT_LEFT_TO_RIGHT | UI_FIT | UI_INVISIBLE).
		//Set(UI_GROW).
		SetChildGap(spacing).
		SetPadding(padding).
		Add(items...)
	return row
}

func Column(padding y3d.Vec4, spacing int, items ...*UIElement) *UIElement {
	column := NewUIElement(y3d.Vec2{}, y3d.Vec2{})
	column.Set(UI_LAYOUT_TOP_TO_BOTTOM | UI_FIT | UI_INVISIBLE).
		//Set(UI_GROW).
		SetChildGap(spacing).
		SetPadding(padding).
		Add(items...)
	return column
}

func RowSpacer() *UIElement {
	spacer := NewUIElement(y3d.Vec2{}, y3d.Vec2{})
	spacer.Set(UI_GROW_X | UI_INVISIBLE)
	return spacer
}

func ColumnSpacer() *UIElement {
	spacer := NewUIElement(y3d.Vec2{}, y3d.Vec2{})
	spacer.Set(UI_GROW_Y | UI_INVISIBLE)
	return spacer
}

func SizedBox(size y3d.Vec2) *UIElement {
	box := NewUIElement(y3d.Vec2{}, size)
	box.Set(UI_INVISIBLE)
	return box
}
