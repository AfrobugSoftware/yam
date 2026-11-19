package yui

func Text(text string, font *Font, color [4]uint8) *UIElement {
	return &UIElement{
		Text:  text,
		Font:  font,
		Color: color,
	}
}
